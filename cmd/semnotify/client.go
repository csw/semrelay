package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	flags "github.com/jessevdk/go-flags"

	"github.com/csw/semrelay"
	internal "github.com/csw/semrelay/internal"
)

var opts struct {
	User     string `short:"u" long:"user"`
	Password string `short:"p" long:"password"`
	Server   string `short:"s" long:"server"`
}

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
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

func run() {
	for {
		if err := runConnection(); err != nil {
			log.Printf("Connection error: %s", err.Error())
		}
		time.Sleep(5 * time.Second)
	}
}

func runConnection() error {
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("wss://%s/ws", opts.Server), nil)
	if err != nil {
		log.Printf("connect failed: %v\n", err)
		return err
	}
	defer conn.Close()
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
		log.Fatal("registration failed", err)
	}
	log.Println("Registered.")

	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			log.Printf("ReadMessage failed: %v\n", err)
			return err
		}
		if payload == nil {
			log.Println("Received empty payload.")
			continue
		}
		fmt.Printf("Received: %s\n", payload)
		if err := notifyUser(payload); err != nil {
			log.Fatal("notifyUser failed: ", err)
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
	return w.Close()
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal("error parsing args: ", err)
	}
	if err := initNotify(); err != nil {
		log.Fatal("DBus connection error: ", err)
	}
	defer func() {
		if err := cleanupNotify(); err != nil {
			log.Printf("cleanup failed: %v\n", err)
		}
	}()
	if len(args) > 0 {
		switch args[0] {
		case "success":
			if err := notifyUser(internal.ExampleSuccess); err != nil {
				log.Fatal("error in notifyUser: ", err)
			}
			time.Sleep(5 * time.Second)
			os.Exit(0)
		case "failure":
			if err := notifyUser(internal.ExampleFailure); err != nil {
				log.Fatal("error in notifyUser: ", err)
			}
			time.Sleep(5 * time.Second)
			os.Exit(0)
		default:
			log.Fatal("unhandled argument", args[0])
		}
	}

	if opts.User == "" || opts.Password == "" || opts.Server == "" {
		log.Fatal("must specify --user, --password, and --server")
	}
	// websocket.DefaultDialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	run()
}
