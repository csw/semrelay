package relay

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/csw/semrelay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dummyClient struct {
	ok        bool
	helloCh   chan struct{}
	msgCh     chan *NotificationTask
	connected bool
}

func newDummyClient() *dummyClient {
	return &dummyClient{
		ok:        true,
		helloCh:   make(chan struct{}),
		msgCh:     make(chan *NotificationTask, 32),
		connected: true,
	}
}

func (dc *dummyClient) String() string {
	return "dummy"
}

func (dc *dummyClient) Hello() {
	dc.helloCh <- struct{}{}
}

func (dc *dummyClient) awaitHello() {
	<-dc.helloCh
}

func (dc *dummyClient) TrySend(msg *NotificationTask) bool {
	if dc.ok {
		select {
		case dc.msgCh <- msg:
			return true
		default:
			return false
		}
	} else {
		return false
	}
}

func (dc *dummyClient) Disconnect() {
	dc.connected = false
}

func TestUserQueue(t *testing.T) {
	user := NewUser("bob")
	go user.Run()
	require.NoError(t, user.Dispatch(json.RawMessage("1")))
	require.NoError(t, user.Dispatch(json.RawMessage("2")))
	c1 := newDummyClient()
	syncJoin(user, c1)
	<-c1.msgCh
	<-c1.msgCh
}

func TestUserQueueDropOldest(t *testing.T) {
	user := NewUser("bob")
	go user.Run()
	for i := 0; i < queueMax+1; i++ {
		require.NoError(t, user.Dispatch(json.RawMessage(fmt.Sprintf("%d", i))))
	}
	time.Sleep(20 * time.Millisecond)
	c1 := newDummyClient()
	syncJoin(user, c1)
	r1 := <-c1.msgCh
	var msg1 semrelay.Message
	require.NoError(t, json.Unmarshal(r1.Payload, &msg1))
	assert.Equal(t, uint8('1'), msg1.Payload[0])
}

func TestUserErrorToQueue(t *testing.T) {
	user := NewUser("bob")
	go user.Run()
	c1 := newDummyClient()
	syncJoin(user, c1)
	c1.ok = false
	require.NoError(t, user.Dispatch(json.RawMessage("1")))
	time.Sleep(20 * time.Millisecond)
	assert.False(t, c1.connected)
	c2 := newDummyClient()
	syncJoin(user, c2)
	r1 := <-c2.msgCh
	var msg1 semrelay.Message
	require.NoError(t, json.Unmarshal(r1.Payload, &msg1))
	assert.Equal(t, uint8('1'), msg1.Payload[0])
}

func TestUserErrorFromQueue(t *testing.T) {
	user := NewUser("bob")
	go user.Run()
	require.NoError(t, user.Dispatch(json.RawMessage("1")))
	c1 := newDummyClient()
	c1.ok = false
	syncJoin(user, c1)
	time.Sleep(20 * time.Millisecond)
	assert.False(t, c1.connected)
	c2 := newDummyClient()
	syncJoin(user, c2)
	r1 := <-c2.msgCh
	var msg1 semrelay.Message
	require.NoError(t, json.Unmarshal(r1.Payload, &msg1))
	assert.Equal(t, uint8('1'), msg1.Payload[0])
}

func TestUserDeregister(t *testing.T) {
	user := NewUser("bob")
	go user.Run()
	c1 := newDummyClient()
	syncJoin(user, c1)
	c2 := newDummyClient()
	syncJoin(user, c2)
	time.Sleep(20 * time.Millisecond)
	require.NoError(t, user.Dispatch(json.RawMessage("1")))
	<-c1.msgCh
	<-c2.msgCh
	user.Leave(c1)
	require.NoError(t, user.Dispatch(json.RawMessage("2")))
	<-c2.msgCh
}

func TestUserRedelivery(t *testing.T) {
	user := NewUser("bob")
	go user.Run()
	c1 := newDummyClient()
	syncJoin(user, c1)
	require.NoError(t, user.Dispatch(json.RawMessage("1")))
	nt1 := <-c1.msgCh
	user.Leave(c1)
	c2 := newDummyClient()
	syncJoin(user, c2)
	nt2 := <-c2.msgCh
	assert.Equal(t, nt1.Id, nt2.Id)
}

func TestUserRedeliveryFailure(t *testing.T) {
	user := NewUser("bob")
	go user.Run()
	c1 := newDummyClient()
	syncJoin(user, c1)
	require.NoError(t, user.Dispatch(json.RawMessage("1")))
	<-c1.msgCh
	user.Leave(c1)
	c2 := newDummyClient()
	c2.ok = false
	syncJoin(user, c2)
	time.Sleep(100 * time.Millisecond)
	assert.False(t, c2.connected)
}

func TestUserAck(t *testing.T) {
	user := NewUser("bob")
	go user.Run()
	c1 := newDummyClient()
	syncJoin(user, c1)
	require.NoError(t, user.Dispatch(json.RawMessage("1")))
	nt1 := <-c1.msgCh
	user.Ack(nt1.Id)
	user.Leave(c1)
	c2 := newDummyClient()
	syncJoin(user, c2)
	timeout := time.NewTimer(100 * time.Millisecond)
	select {
	case <-c2.msgCh:
		t.Fatal("saw redelivery")
	case <-timeout.C:
		// OK
	}
}

func TestDropSlowUser(t *testing.T) {
	user := NewUser("bob")
	go user.Run()
	c1 := newDummyClient()
	c1.ok = false
	syncJoin(user, c1)
	require.NoError(t, user.Dispatch(json.RawMessage("1")))
	time.Sleep(20 * time.Millisecond)
	assert.False(t, c1.connected)
}

func TestUserDispatchGarbage(t *testing.T) {
	user := NewUser("bob")
	assert.Error(t, user.Dispatch([]byte{0}))
}

func syncJoin(user *User, client *dummyClient) {
	go user.Join(client)
	client.awaitHello()
}
