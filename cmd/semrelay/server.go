package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/caddyserver/certmagic"

	internal "github.com/csw/semrelay/internal"
)

var password string
var token string

func main() {
	domain := os.Getenv("DOMAIN")
	if domain == "" {
		log.Fatal("Must specify DOMAIN")
	}
	password = os.Getenv("PASSWORD")
	if password == "" {
		log.Fatal("Must specify PASSWORD")
	}
	token = os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("Must specify TOKEN")
	}
	certmagic.DefaultACME.Email = os.Getenv("EMAIL")
	if os.Getenv("STAGING") != "" {
		certmagic.DefaultACME.CA = certmagic.LetsEncryptStagingCA
	}
	disp := newDispatcher()
	go disp.run()
	mux := http.NewServeMux()
	mux.HandleFunc("/hook", func(w http.ResponseWriter, r *http.Request) {
		handleHook(disp, w, r)
	})
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(disp, w, r)
	})
	if user := os.Getenv("TEST"); user != "" {
		go func() {
			for {
				time.Sleep(15 * time.Second)
				disp.send(user, internal.ExampleSuccess)
			}
		}()
	}
	err := certmagic.HTTPS([]string{domain}, mux)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
