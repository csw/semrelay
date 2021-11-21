// Adapted from:
// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"github.com/csw/semrelay"
	"github.com/csw/semrelay/relay"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	dispatcher *relay.Dispatcher
	user       *relay.User

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

func (c *Client) String() string {
	return c.conn.RemoteAddr().String()
}

func (c *Client) TrySend(msg *relay.NotificationTask) bool {
	select {
	case c.send <- msg.Payload:
		// sent
		return true
	default:
		// queue is full, drop the connection
		return false
	}
}

func (c *Client) Disconnect() {
	log.WithField("conn", c.String()).Info("Disconnecting client")
	close(c.send)
}

// readPump reads registration and acknowledgemnt messages from the notification
// client.
func (c *Client) readPump() {
	defer c.conn.Close()
	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		panic(err)
	}
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	username, err := c.awaitRegister()
	ulog := log.WithField("user", username).WithField("conn", c.String())
	if err != nil {
		ulog.WithError(err).Error("Registration failed")
		return
	}
	c.user = c.dispatcher.Register(username, c)
	defer func() {
		c.user.Leave(c)
	}()
	for {
		var msg semrelay.Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				ulog.Info("Connection closed")
				break
			}
			log.WithError(err).Error("Error reading message")
		}
		switch msg.Type {
		case semrelay.AckMsg:
			c.user.Ack(msg.Id)
		default:
			ulog.WithField("type", msg.Type).Error("Unexpected message from client")
		}
	}
}

func (c *Client) awaitRegister() (string, error) {
	var msg semrelay.Message
	err := c.conn.ReadJSON(&msg)
	if err != nil {
		return "", err
	}
	if msg.Type != semrelay.RegistrationMsg {
		return "", fmt.Errorf("Expected registration message, got %s", msg.Type)
	}
	var reg semrelay.Registration
	if err := json.Unmarshal(msg.Payload, &reg); err != nil {
		return "", err
	}
	if reg.User == "" {
		return "", errors.New("no user specified")
	}
	if reg.Password != password {
		return "", errors.New("password mismatch")
	}
	return reg.User, nil
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				panic(err)
			}
			if !ok {
				// The hub closed the channel.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.WithError(err).WithField("conn", c.String()).Error("Error sending message")
			}

		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				panic(err)
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(disp *relay.Dispatcher, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{dispatcher: disp, conn: conn, send: make(chan []byte, 32)}

	go client.writePump()
	go client.readPump()
}
