package relay

import (
	"encoding/json"
	"log"
	"math/rand"

	"github.com/csw/semrelay"
)

type User struct {
	name     string
	msgCh    chan *NotificationTask
	ackCh    chan uint64
	joinCh   chan Client
	leaveCh  chan Client
	queue    []*NotificationTask
	inFlight []*NotificationTask
	clients  []Client
}

const (
	queueMax = 8
)

func NewUser(name string) *User {
	return &User{
		name:    name,
		msgCh:   make(chan *NotificationTask, queueMax),
		ackCh:   make(chan uint64, 1),
		joinCh:  make(chan Client, 1),
		leaveCh: make(chan Client, 1),
	}
}

func (u *User) Dispatch(payload json.RawMessage) error {
	id := rand.Uint64()
	msg := semrelay.MakeNotification(id, payload)
	enc, err := json.Marshal(&msg)
	if err != nil {
		return err
	}
	u.msgCh <- NewNotificationTask(id, u.name, enc)
	return nil
}

func (u *User) Ack(id uint64) {
	u.ackCh <- id
}

func (u *User) Join(client Client) {
	u.joinCh <- client
}

func (u *User) Leave(client Client) {
	u.leaveCh <- client
}

func (u *User) Run() {
	for {
		select {
		case msg := <-u.msgCh:
			u.onDispatch(msg)
		case id := <-u.ackCh:
			u.onAck(id)
		case client := <-u.joinCh:
			u.register(client)
		case client := <-u.leaveCh:
			u.deregister(client)
		}
	}
}

func (u *User) onDispatch(msg *NotificationTask) {
	if len(u.clients) > 0 {
		// send to each active client
		sent := false
		for _, client := range u.clients {
			if !client.TrySend(msg) {
				log.Println("Failed to send message to", client)
				u.deregister(client)
				break
			}
			sent = true
		}
		if sent {
			u.inFlight = appendBounded(u.inFlight, msg)
		} else {
			u.queue = appendBounded(u.queue, msg)
		}
	} else {
		// no clients connected, queue up the message
		u.queue = appendBounded(u.queue, msg)
	}
}

func (u *User) onAck(id uint64) {
	for i, task := range u.inFlight {
		if task.Id == id {
			// copy rest of slice forward, truncate
			copy(u.inFlight[i:], u.inFlight[i+1:])
			u.inFlight = u.inFlight[:len(u.inFlight)-1]
			return
		}
	}
}

func (u *User) register(client Client) {
	if len(u.clients) == 0 {
		// send pending messages
		for _, msg := range u.inFlight {
			if !client.TrySend(msg) {
				log.Println("Failed to send pending in-flight messages to ", client)
				client.Disconnect()
				return
			}
		}
		for _, msg := range u.queue {
			if !client.TrySend(msg) {
				log.Println("Failed to send pending queued messages to ", client)
				client.Disconnect()
				return
			}
		}
		u.inFlight = append(u.inFlight, u.queue...)
		u.queue = u.queue[:0]
	}
	u.clients = append(u.clients, client)
	log.Printf("Registered %s for %s\n", client, u.name)
}

func (u *User) deregister(client Client) {
	var nClients []Client
	for _, existing := range u.clients {
		if existing != client {
			nClients = append(nClients, existing)
		}
	}
	u.clients = nClients
	client.Disconnect()
	log.Printf("Deregistered %s for %s\n", client, u.name)
}

func appendBounded(q []*NotificationTask, nt *NotificationTask) []*NotificationTask {
	if len(q) < queueMax {
		return append(q, nt)
	} else {
		copy(q, q[1:])
		q[queueMax-1] = nt
		return q
	}
}
