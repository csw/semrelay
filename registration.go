package semrelay

// Registration is sent by the client as JSON to authenticate and request
// notifications for a given user.
type Registration struct {
	User     string `json:"user"`
	Password string `json:"password"`
}
