package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	"github.com/csw/semrelay"
	"github.com/esiqveland/notify"
	dbus "github.com/godbus/dbus/v5"
)

var dConn *dbus.Conn
var notifier notify.Notifier

type registration struct {
	id  uint32
	url string
}

var registry = make(map[uint32]string)

var registerCh = make(chan registration, 8)
var clickCh = make(chan uint32, 8)

var icon *DBusIcon

func notifyUser(payload []byte) error {
	var semN semrelay.Notification
	if err := json.Unmarshal(payload, &semN); err != nil {
		return err
	}
	if semN.Pipeline.Id != semN.Workflow.InitialPipelineId {
		log.Printf("Ignoring result for pipeline %s\n", semN.Pipeline.YamlFileName)
		return nil
	}
	urgency := dbus.MakeVariant(1)
	if semN.Pipeline.Result == "failed" {
		urgency = dbus.MakeVariant(byte(2))
	}
	url := fmt.Sprintf("https://%s.semaphoreci.com/workflows/%s?pipeline_id=%s",
		semN.Organization.Name, semN.Workflow.Id, semN.Pipeline.Id)
	tag := fmt.Sprintf("%s/%s", semN.Project.Name, semN.Revision.Branch.Name)
	titleText, err := title(&semN)
	if err != nil {
		return err
	}
	n := notify.Notification{
		AppName:    "Semaphore",
		ReplacesID: uint32(0),
		Summary:    titleText,
		Body:       body(&semN),
		Actions: []notify.Action{
			{Key: "default", Label: "Open"},
		},
		Hints: map[string]dbus.Variant{
			"urgency":           urgency,
			"category":          dbus.MakeVariant(tag),
			"x-dunst-stack-tag": dbus.MakeVariant(tag),
			"image-data":        dbus.MakeVariant(icon),
		},
	}
	id, err := notifier.SendNotification(n)
	if err != nil {
		return err
	}
	registerCh <- registration{id: id, url: url}
	return nil
}

func onAction(action *notify.ActionInvokedSignal) {
	clickCh <- action.ID
}

func runHandler() {
	for {
		select {
		case reg := <-registerCh:
			registry[reg.id] = reg.url
		case id := <-clickCh:
			url := registry[id]
			delete(registry, id)
			log.Printf("Opening URL on click: %s\n", url)
			cmd := exec.Command("/usr/bin/xdg-open", url)
			err := cmd.Run()
			if err != nil {
				log.Printf("Opening URL %s failed: %v\n", url, err)
			}
		}
	}
}

func initNotify() error {
	var err error
	dConn, err = dbus.SessionBusPrivate()
	if err != nil {
		return err
	}

	if err = dConn.Auth(nil); err != nil {
		dConn.Close()
		return err
	}

	if err = dConn.Hello(); err != nil {
		dConn.Close()
		return err
	}

	notifier, err = notify.New(dConn, notify.WithOnAction(onAction))
	if err != nil {
		return err
	}

	icon = buildIcon(semrelay.IconImage)

	go runHandler()

	return nil
}

func cleanupNotify() error {
	if err := notifier.Close(); err != nil {
		log.Printf("Failed to close notifier: %v\n", err)
	}
	if err := dConn.Close(); err != nil {
		log.Printf("Failed to close DBus connection: %v\n", err)
		return err
	}
	return nil
}
