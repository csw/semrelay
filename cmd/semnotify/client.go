package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"github.com/adrg/xdg"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/sys/unix"

	"github.com/csw/semrelay"
	internal "github.com/csw/semrelay/internal"
)

var (
	user       string
	password   string
	server     string
	insecure   bool
	promotions bool
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var (
	curClient atomic.Value
	clientCtx context.Context
)

func run() error {
	if err := initDBus(); err != nil {
		return err
	}
	defer func() {
		if err := cleanupDBus(); err != nil {
			log.WithError(err).Error("Cleanup failed")
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

func runConnection() error {
	var err error
	url := fmt.Sprintf("wss://%s/ws", server)
	conn, _, err := websocket.DefaultDialer.DialContext(clientCtx, url, nil)
	if err != nil {
		log.WithError(err).Debug("Connection failed")
		return nil
	}
	client := newClient(clientCtx, conn)
	defer client.close()
	curClient.Store(client)
	if err := clientCtx.Err(); err != nil {
		return err
	}
	if err := client.initPings(); err != nil {
		log.WithError(err).Error("Failed to set up ping handling")
	}
	log.Infof("Connected to %s.", server)

	register(client)
	log.Debug("Registered.")

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				log.Warn("Connection closed.")
				return nil
			}
			log.WithError(err).Error("ReadMessage failed")
			return nil
		}
		if err := handleMessage(client, raw); err != nil {
			log.WithError(err).Error("Error handling message")
			return nil
		}
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

type Client struct {
	conn   *websocket.Conn
	sendCh chan *semrelay.Message
	sendWG sync.WaitGroup
}

func newClient(parentCtx context.Context, conn *websocket.Conn) *Client {
	client := &Client{
		conn:   conn,
		sendCh: make(chan *semrelay.Message),
	}
	client.sendWG.Add(1)
	go client.runSend()
	return client
}

func (c *Client) close() error {
	if c.sendCh != nil {
		close(c.sendCh)
		c.sendCh = nil
	}
	c.sendWG.Wait()
	return c.conn.Close()
}

func (c *Client) runSend() {
	defer c.sendWG.Done()
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case msg, ok := <-c.sendCh:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				panic(err)
			}
			if !ok {
				// Connection is being closed
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(msg); err != nil {
				log.WithError(err).Error("Error sending message")
				return
			}
		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				panic(err)
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}
				log.WithError(err).Error("Error sending ping")
				return
			}
		}
	}
}

func (c *Client) initPings() error {
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return err
	}
	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			return err
		}
		return nil
	})
	return nil
}

func handleMessage(client *Client, raw []byte) error {
	log.Debugf("Received %d bytes: %s", len(raw), raw)
	var msg semrelay.Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		return fmt.Errorf("parsing message failed: %w", err)
	}
	switch msg.Type {
	case semrelay.NotificationMsg:
		var semN semrelay.Notification
		if err := json.Unmarshal(msg.Payload, &semN); err != nil {
			return err
		}
		if err := notifyUser(&semN); err != nil {
			return err
		}
		ack := semrelay.MakeAck(msg.Id)
		log.Debugf("Sending ack for message %d", msg.Id)
		client.sendCh <- &ack
	default:
		return fmt.Errorf("Unhandled message type: %s", msg.Type)
	}
	return nil
}

func register(client *Client) {
	client.sendCh <- semrelay.MakeRegistration(user, password)
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
	if err := initDBus(); err != nil {
		return fmt.Errorf("DBus connection error: %s", err)
	}
	var semN semrelay.Notification
	if err := json.Unmarshal(msg, &semN); err != nil {
		panic(err)
	}
	err := notifyUser(&semN)
	if err == nil {
		time.Sleep(5 * time.Second)
	}
	if cerr := cleanupDBus(); cerr != nil {
		return cerr
	}
	return err
}

// closer closes the current connection asynchronously when a signal is
// received, to interrupt a blocking read.
func closer() {
	<-clientCtx.Done()
	client := curClient.Load()
	if client == nil {
		return
	}
	_ = client.(*Client).close()
}

func parseConfig() error {
	cfg, err := xdg.SearchConfigFile("semnotify/config")
	if err != nil {
		// config file not found
		return nil
	}
	viper.SetConfigType("env")
	viper.SetConfigFile(cfg)
	return viper.ReadInConfig()
}

func processConfig() error {
	user = viper.GetString("user")
	password = viper.GetString("password")
	server = viper.GetString("server")
	insecure = viper.GetBool("insecure")
	promotions = viper.GetBool("promotions")

	if user == "" {
		return errors.New("must specify user in configuration")
	}
	if password == "" {
		return errors.New("must specify password in configuration")
	}
	if server == "" {
		return errors.New("must specify server in configuration")
	}

	if viper.GetBool("verbose") {
		log.SetLevel(log.DebugLevel)
	}

	return nil
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:          true,
		TimestampFormat:        time.Stamp,
		DisableLevelTruncation: true,
	})

	pflag.StringP("user", "u", "", "GitHub user to receive notifications for")
	pflag.StringP("password", "p", "", "semrelay password")
	pflag.StringP("server", "s", "", "semrelay hostname")
	pflag.BoolP("verbose", "v", false, "Verbose mode")
	pflag.Bool("insecure", false, "Disable TLS certificate verification")
	pflag.Bool("promotions", true, "Show promotion results")
	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		panic(err)
	}
	if err := parseConfig(); err != nil {
		log.WithError(err).Fatal("Error parsing configuration")
	}
	if err := processConfig(); err != nil {
		log.WithError(err).Fatal("Configuration error")
	}

	clientCtx, _ = signal.NotifyContext(context.Background(),
		os.Interrupt, os.Kill, unix.SIGTERM, unix.SIGHUP)
	go closer()
	args := pflag.Args()
	if len(args) > 0 {
		if err := sendExample(args[0]); err != nil {
			log.WithError(err).Fatal("Sending example failed")
		}
		os.Exit(0)
	}

	if insecure {
		websocket.DefaultDialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if err := run(); err != nil {
		if err != context.Canceled {
			log.WithError(err).Fatal("Exiting with error")
		}
	}
}
