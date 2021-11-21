package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/csw/semrelay"
	"github.com/csw/semrelay/relay"
)

func handleHook(d *relay.Dispatcher, w http.ResponseWriter, r *http.Request) {
	curToken := r.URL.Query().Get("token")
	if curToken != token {
		log.WithField("token", curToken).Error("Wrong token for webhook message")
		w.WriteHeader(400)
		fmt.Fprintln(w, "nope")
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.WithError(err).Error("Read error on webhook message")
		w.WriteHeader(500)
		fmt.Fprintln(w, "nope")
		return
	}
	log.Debugf("Got webhook notification: %s", body)
	var n semrelay.Notification
	if err := json.Unmarshal(body, &n); err != nil {
		log.WithError(err).Error("Failed to parse webhook message")
		w.WriteHeader(400)
		fmt.Fprintln(w, "nope")
		return
	}
	user := n.Revision.Sender.Login
	if user == "" {
		log.Error("No user in webhook message")
		w.WriteHeader(400)
		fmt.Fprintln(w, "nope")
		return
	}
	log.WithFields(log.Fields{
		"user":       user,
		"repository": n.Repository.Slug,
		"done_at":    n.Pipeline.DoneAt,
		"pipeline":   n.Pipeline.Id,
	}).Info("Received build notification")
	d.Dispatch(user, body)
	fmt.Fprintln(w, "Roger")
}
