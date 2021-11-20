package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/csw/semrelay"
	"github.com/csw/semrelay/relay"
)

func handleHook(d *relay.Dispatcher, w http.ResponseWriter, r *http.Request) {
	curToken := r.URL.Query().Get("token")
	if curToken != token {
		log.Printf("Wrong token %s for webhook message\n", curToken)
		w.WriteHeader(400)
		fmt.Fprintln(w, "nope")
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Read error on webhook message: %v\n", err)
		w.WriteHeader(500)
		fmt.Fprintln(w, "nope")
		return
	}
	log.Printf("Got webhook notification: %s\n", body)
	var n semrelay.Notification
	if err := json.Unmarshal(body, &n); err != nil {
		log.Printf("Failed to parse webhook message: %v\n", err)
		w.WriteHeader(400)
		fmt.Fprintln(w, "nope")
		return
	}
	user := n.Revision.Sender.Login
	if user == "" {
		log.Println("No user in webhook message")
		w.WriteHeader(400)
		fmt.Fprintln(w, "nope")
		return
	}
	d.Dispatch(user, body)
	fmt.Fprintln(w, "Roger")
}
