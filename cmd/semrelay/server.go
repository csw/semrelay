package main

import (
	"net/http"
	"os"
	"time"

	"github.com/caddyserver/certmagic"
	log "github.com/sirupsen/logrus"

	internal "github.com/csw/semrelay/internal"
	"github.com/csw/semrelay/relay"
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
	if os.Getenv("VERBOSE") != "" {
		log.SetLevel(log.DebugLevel)
	}
	disp := relay.NewDispatcher()
	go disp.Run()
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
				disp.Dispatch(user, internal.ExampleSuccess)
			}
		}()
	}
	var err error
	if os.Getenv("HTTP_ONLY") != "" {
		port := "80"
		if portspec := os.Getenv("PORT"); portspec != "" {
			port = portspec
		}
		err = http.ListenAndServe(":"+port, mux)
	} else {
		err = certmagic.HTTPS([]string{domain}, mux)
	}
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
