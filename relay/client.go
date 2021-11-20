package relay

type Client interface {
	String() string

	TrySend(msg *NotificationTask) bool

	Disconnect()
}
