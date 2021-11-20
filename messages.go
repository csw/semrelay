package semrelay

import "encoding/json"

const (
	RegistrationMsg = "registration"
	NotificationMsg = "notification"
	AckMsg          = "ack"
)

type Message struct {
	Type    string          `json:"type"`
	Id      uint64          `json:"id"`
	Payload json.RawMessage `json:"payload"`
}

func MakeRegistration(user, password string) *Message {
	reg, err := json.Marshal(Registration{User: user, Password: password})
	if err != nil {
		panic(err)
	}
	return &Message{
		Type:    RegistrationMsg,
		Payload: reg,
	}
}

func MakeNotification(id uint64, payload json.RawMessage) *Message {
	return &Message{
		Type:    NotificationMsg,
		Id:      id,
		Payload: payload,
	}
}

func MakeAck(id uint64) Message {
	return Message{Type: AckMsg, Id: id}
}
