package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type session struct {
	user   string
	client *Client
}

type dispatch struct {
	user    string
	payload []byte
}

type dispatcher struct {
	joinCh     chan session
	leaveCh    chan session
	dispatchCh chan dispatch

	clients map[string][]*Client
}

func newDispatcher() *dispatcher {
	return &dispatcher{
		joinCh:     make(chan session, 8),
		leaveCh:    make(chan session, 8),
		dispatchCh: make(chan dispatch, 8),
		clients:    make(map[string][]*Client),
	}
}

func (d *dispatcher) register(user string, client *Client) {
	log.Printf("Registering %s @ %s\n", user, client.conn.RemoteAddr())
	d.joinCh <- session{user: user, client: client}
}

func (d *dispatcher) unregister(user string, client *Client) {
	log.Printf("Unregistering %s @ %s\n", user, client.conn.RemoteAddr())
	d.leaveCh <- session{user: user, client: client}
}

func (d *dispatcher) send(user string, payload []byte) {
	log.Printf("Sending %d bytes to %s\n", len(payload), user)
	d.dispatchCh <- dispatch{user: user, payload: payload}
}

func (d *dispatcher) onRegister(sess session) {
	d.clients[sess.user] = append(d.clients[sess.user], sess.client)
}

func (d *dispatcher) onUnregister(sess session) {
	oldClients := d.clients[sess.user]
	newClients := make([]*Client, 0, len(oldClients))
	for _, client := range d.clients[sess.user] {
		if client != sess.client {
			newClients = append(newClients, client)
		}
	}
	d.clients[sess.user] = newClients
	// bye
	close(sess.client.send)
}

func (d *dispatcher) run() {
	for {
		select {
		case sess := <-d.joinCh:
			d.onRegister(sess)
		case sess := <-d.leaveCh:
			d.onUnregister(sess)
		case dispatch := <-d.dispatchCh:
			var prepared *websocket.PreparedMessage
			var err error
			for _, client := range d.clients[dispatch.user] {
				if prepared == nil {
					prepared, err = websocket.NewPreparedMessage(websocket.TextMessage, dispatch.payload)
					if err != nil {
						log.Printf("Unable to prepare message: %s\n", dispatch.payload)
						panic(err)
					}
				}
				log.Printf("Sending %d bytes to %s @ %s\n", len(dispatch.payload), dispatch.user, client.conn.RemoteAddr())
				select {
				case client.send <- prepared:
					// sent
				default:
					// queue is full, drop the connection
					d.onUnregister(session{user: dispatch.user, client: client})
				}
			}
		}
	}
}
