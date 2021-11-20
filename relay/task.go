package relay

import (
	"time"
)

type NotificationTask struct {
	Id      uint64
	User    string
	Sent    *time.Time
	Payload []byte
}

func NewNotificationTask(id uint64, user string, payload []byte) *NotificationTask {
	return &NotificationTask{
		Id:      id,
		User:    user,
		Sent:    nil,
		Payload: payload,
	}
}
