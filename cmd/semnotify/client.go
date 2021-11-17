package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	flags "github.com/jessevdk/go-flags"
	"golang.org/x/sys/unix"

	"github.com/csw/semrelay"
	internal "github.com/csw/semrelay/internal"
)

var opts struct {
	User     string `short:"u" long:"user" description:"GitHub username to show notifications for"`
	Password string `short:"p" long:"password" description:"Relay server password"`
	Server   string `short:"s" long:"server" description:"Relay server hostname"`
	Insecure bool   `long:"insecure" description:"Disable TLS certification verification (for staging certificates only)"`
}

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var (
	curConn   atomic.Value
	clientCtx context.Context
)

func notifyUser(payload []byte) error {
	var semN semrelay.Notification
	if err := json.Unmarshal(payload, &semN); err != nil {
		return err
	}
	return notifyUserPlatform(&semN)
}

func title(semN *semrelay.Notification) (string, error) {
	startT, err := time.Parse(time.RFC3339, semN.Pipeline.RunningAt)
	if err != nil {
		return "", err
	}
	doneT, err := time.Parse(time.RFC3339, semN.Pipeline.DoneAt)
	if err != nil {
		return "", err
	}
	mins := doneT.Sub(startT).Minutes()
	return fmt.Sprintf("Build %s for %s:%s in %.0fm",
		semN.Pipeline.Result, semN.Project.Name, semN.Revision.Branch.Name, mins), nil
}

func body(semN *semrelay.Notification) string {
	b := strings.Builder{}
	fmt.Fprintf(&b, "Commit %s: %s\n",
		semN.Revision.CommitSHA[:7], semN.Revision.CommitMessage)
	if semN.Pipeline.Result == "failed" {
		blockParts := []string{}
		for _, block := range semN.Blocks {
			jobParts := []string{}
			for _, job := range block.Jobs {
				if job.Result == "failed" {
					jobParts = append(jobParts, job.Name)
				}
			}
			if len(jobParts) > 0 {
				blockParts = append(blockParts,
					fmt.Sprintf("%s (%s)", block.Name, strings.Join(jobParts, ", ")))
			}
		}
		fmt.Fprintf(&b, "Failed in %s\n", strings.Join(blockParts, ", "))
	}
	return b.String()
}

func run() error {
	if err := initNotify(); err != nil {
		return err
	}
	defer func() {
		if err := cleanupNotify(); err != nil {
			log.Println("Cleanup failed: ", err)
		}
	}()
	for {
		if err := clientCtx.Err(); err != nil {
			// check for cancellation
			return err
		}
		if err := runConnection(); err != nil {
			return err
		}
		sleep(5 * time.Second)
	}
}

func sleep(d time.Duration) {
	t := time.NewTimer(d)
	// don't bother stopping the timer, if the context is cancelled we're about
	// to exit
	select {
	case <-t.C:
		return
	// allow interruption by context cancellation
	case <-clientCtx.Done():
		return
	}
}

func runConnection() error {
	var err error
	conn, _, err := websocket.DefaultDialer.DialContext(clientCtx, fmt.Sprintf("wss://%s/ws", opts.Server), nil)
	if err != nil {
		log.Println("connect failed: ", err)
		return nil
	}
	defer conn.Close()
	curConn.Store(conn)
	if err := clientCtx.Err(); err != nil {
		return err
	}
	fmt.Println("Connected.")
	if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return err
	}
	conn.SetPongHandler(func(string) error {
		if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			return err
		}
		return nil
	})
	go pinger(conn)

	if err := register(conn); err != nil {
		log.Println("registration failed: ", err)
		return nil
	}
	log.Println("Registered.")

	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			log.Println("ReadMessage failed: ", err)
			return nil
		}
		if len(payload) == 0 {
			log.Println("Received empty payload.")
			continue
		}
		log.Printf("Received %d bytes: %s\n", len(payload), payload)
		if err := notifyUser(payload); err != nil {
			return nil
		}
	}
}

func pinger(conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()
	for range ticker.C {
		if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
			log.Printf("Error setting write deadline: %v\n", err)
			return
		}
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Printf("Error sending ping: %v\n", err)
			return
		}
	}
}

func register(conn *websocket.Conn) error {
	w, err := conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	msg, err := json.Marshal(&semrelay.Registration{
		User:     opts.User,
		Password: opts.Password,
	})
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	// The authentication protocol is as simple as it gets. If authentication
	// fails, the server will just drop the connection.
	return w.Close()
}

func sendExample(name string) error {
	var msg []byte
	switch name {
	case "success":
		msg = internal.ExampleSuccess
	case "failure":
		msg = internal.ExampleFailure
	default:
		return fmt.Errorf("unhandled argument %s", name)
	}
	if err := initNotify(); err != nil {
		return fmt.Errorf("DBus connection error: %s", err)
	}
	err := notifyUser(msg)
	if err == nil {
		time.Sleep(5 * time.Second)
	}
	if cerr := cleanupNotify(); cerr != nil {
		return cerr
	}
	return err
}

// closer closes the current connection asynchronously when a signal is
// received, to interrupt a blocking read.
func closer() {
	<-clientCtx.Done()
	conn := curConn.Load()
	if conn == nil {
		return
	}
	_ = conn.(*websocket.Conn).Close()
}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	args, err := parser.Parse()
	if err != nil {
		if !flags.WroteHelp(err) {
			parser.WriteHelp(os.Stderr)
		}
		os.Exit(1)
	}
	clientCtx, _ = signal.NotifyContext(context.Background(),
		os.Interrupt, os.Kill, unix.SIGTERM, unix.SIGHUP)
	go closer()
	if len(args) > 0 {
		if err := sendExample(args[0]); err != nil {
			log.Fatal("Sending example failed: ", err)
		}
		os.Exit(0)
	}

	if opts.User == "" || opts.Password == "" || opts.Server == "" {
		log.Fatal("must specify --user, --password, and --server")
	}
	if opts.Insecure {
		websocket.DefaultDialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if err := run(); err != nil {
		log.Fatal("Error: ", err)
	}
}
