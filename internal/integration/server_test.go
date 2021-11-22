//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	"github.com/csw/semrelay"
	"github.com/csw/semrelay/internal"
)

const (
	testUser     = "csw"
	testPassword = "password"
	testToken    = "token"
)

func TestBasic(t *testing.T) {
	conn := wsConn(t, testUser, testPassword)
	defer conn.Close()
	sendHook(t, internal.ExampleSuccess)
	n, err := readNotification(t, conn)
	require.NoError(t, err)
	require.Equal(t, testUser, n.Revision.Sender.Login)
}

func TestMultiple(t *testing.T) {
	c1 := wsConn(t, testUser, testPassword)
	defer c1.Close()
	c2 := wsConn(t, testUser, testPassword)
	defer c2.Close()
	sendHook(t, internal.ExampleSuccess)
	n, err := readNotification(t, c1)
	require.NoError(t, err)
	require.Equal(t, testUser, n.Revision.Sender.Login)
	n2, err := readNotification(t, c2)
	require.NoError(t, err)
	require.Equal(t, n, n2)
}

func TestDifferentUser(t *testing.T) {
	c1 := wsConn(t, testUser, testPassword)
	defer c1.Close()
	c2 := wsConn(t, "bob", testPassword)
	defer c2.Close()
	sendHook(t, internal.ExampleSuccess)
	n, err := readNotification(t, c1)
	require.NoError(t, err)
	require.Equal(t, testUser, n.Revision.Sender.Login)
	go func() {
		time.Sleep(500 * time.Millisecond)
		c2.Close()
	}()
	n2, err := readNotification(t, c2)
	require.Nil(t, n2)
	require.True(t, errors.Is(err, net.ErrClosed) || err == io.EOF, "unexpected error: %v", err)
}

func wsConn(t *testing.T, user, password string) *websocket.Conn {
	wsUrl := fmt.Sprintf("ws://localhost:%s/ws", os.Getenv("TARGET_PORT"))
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	require.NoError(t, err)
	reg := semrelay.MakeRegistration(user, password)
	err = conn.WriteJSON(&reg)
	require.NoError(t, err)
	return conn
}

func sendHook(t *testing.T, body []byte) {
	hookUrl := fmt.Sprintf("http://localhost:%s/hook?token=%s",
		os.Getenv("TARGET_PORT"), testToken)
	hookBody := bytes.NewReader(body)
	res, err := http.Post(hookUrl, "application/json", hookBody)
	require.NoError(t, err)
	require.Equal(t, 200, res.StatusCode)
	require.NoError(t, res.Body.Close())
}

func readNotification(t *testing.T, conn *websocket.Conn) (*semrelay.Notification, error) {
	var msg semrelay.Message
	if err := conn.ReadJSON(&msg); err != nil {
		return nil, err
	}
	require.Equal(t, msg.Type, semrelay.NotificationMsg)
	var n semrelay.Notification
	if err := json.Unmarshal(msg.Payload, &n); err != nil {
		return nil, err
	}
	return &n, nil
}
