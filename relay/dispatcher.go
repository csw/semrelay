package relay

import (
	log "github.com/sirupsen/logrus"
)

type session struct {
	user   string
	client Client
	userCh chan<- *User
}

type dispatch struct {
	user    string
	payload []byte
}

type Dispatcher struct {
	joinCh     chan session
	dispatchCh chan dispatch

	users map[string]*User
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		joinCh:     make(chan session, 8),
		dispatchCh: make(chan dispatch, 8),
		users:      make(map[string]*User),
	}
}

func (d *Dispatcher) Register(user string, client Client) *User {
	log.WithFields(log.Fields{"user": user, "client": client}).Debug("Registering")
	userCh := make(chan *User, 1)
	d.joinCh <- session{user: user, client: client, userCh: userCh}
	return <-userCh
}

func (d *Dispatcher) Dispatch(user string, payload []byte) {
	d.dispatchCh <- dispatch{user: user, payload: payload}
}

func (d *Dispatcher) onRegister(sess session) {
	user := d.users[sess.user]
	if user == nil {
		user = NewUser(sess.user)
		d.users[sess.user] = user
		go user.Run()
	}
	user.Join(sess.client)
	sess.userCh <- user
}

func (d *Dispatcher) onDispatch(msg dispatch) {
	if user := d.users[msg.user]; user != nil {
		if err := user.Dispatch(msg.payload); err != nil {
			log.Println("Error dispatching message: ", err)
		}
	}
}

func (d *Dispatcher) Run() {
	for {
		select {
		case sess := <-d.joinCh:
			d.onRegister(sess)
		case msg := <-d.dispatchCh:
			d.onDispatch(msg)
		}
	}
}
