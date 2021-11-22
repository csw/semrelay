package relay

type Client interface {
	String() string

	Hello()

	TrySend(msg *NotificationTask) bool

	Disconnect()
}
