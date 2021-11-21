package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/csw/semrelay"
	"github.com/esiqveland/notify"
	dbus "github.com/godbus/dbus/v5"
	log "github.com/sirupsen/logrus"
)

var dConn *dbus.Conn
var notifier notify.Notifier

type registration struct {
	id  uint32
	url string
}

// registry associates DBus notification IDs with the corresponding URLs for the
// Semaphore pipeline pages, to display when the user clicks on a notification.
var registry = make(map[uint32]string)

var registerCh = make(chan registration, 8)
var clickCh = make(chan uint32, 8)

var icon *DBusIcon

func notifyUserPlatform(semN *semrelay.Notification) error {
	if !promotions && semN.Pipeline.Id != semN.Workflow.InitialPipelineId {
		// Only display results for the original pipeline. This avoids
		// displaying notifications for automatic promotions that might validly
		// fail.
		log.Debugf("Ignoring result for pipeline %s", semN.Pipeline.YamlFileName)
		return nil
	}
	urgency := dbus.MakeVariant(1) // Normal
	if semN.Pipeline.Result == "failed" {
		urgency = dbus.MakeVariant(byte(2)) // Critical
	}
	url := fmt.Sprintf("https://%s.semaphoreci.com/workflows/%s?pipeline_id=%s",
		semN.Organization.Name, semN.Workflow.Id, semN.Pipeline.Id)
	// The stacking tag is for x-dunst-stack-tag, so that newer notifications
	// for a given branch will be displayed instead of older ones, rather than
	// alongside them.
	tag := fmt.Sprintf("%s/%s", semN.Project.Name, semN.Revision.Branch.Name)
	titleText, err := title(semN)
	if err != nil {
		return err
	}
	n := notify.Notification{
		AppName:    "Semaphore",
		ReplacesID: uint32(0),
		Summary:    titleText,
		Body:       body(semN),
		Actions: []notify.Action{
			{Key: "default", Label: "Open"},
		},
		Hints: map[string]dbus.Variant{
			"urgency":           urgency,
			"x-dunst-stack-tag": dbus.MakeVariant(tag),
			"image-data":        dbus.MakeVariant(icon),
		},
	}
	id, err := notifier.SendNotification(n)
	if err != nil {
		return err
	}
	// Register the URL to be displayed when the action is invoked.
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
			url, found := registry[id]
			if !found {
				continue
			}
			delete(registry, id)
			log.WithField("url", url).Debug("Opening URL on click")
			cmd := exec.Command("xdg-open", url)
			err := cmd.Run()
			if err != nil {
				log.WithField("url", url).WithError(err).Error("Error opening URL")
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
		dConn.Close()
		return err
	}
	server, err := notifier.GetServerInformation()
	if err != nil {
		dConn.Close()
		return err
	}
	log.Debugf("Notification daemon: %s (%s), version %s, specification version %s\n",
		server.Name, server.Vendor, server.Version, server.SpecVersion)
	caps, err := notifier.GetCapabilities()
	if err != nil {
		dConn.Close()
		return err
	}
	log.Debugf("Notification daemon capabilities: %s\n", strings.Join(caps, ", "))

	icon = buildIcon(semrelay.IconImage)

	go runHandler()

	return nil
}

func cleanupNotify() error {
	if err := notifier.Close(); err != nil {
		log.WithError(err).Error("Error closing notifier")
	}
	if err := dConn.Close(); err != nil {
		log.WithError(err).Error("Error closing DBus connection")
		return err
	}
	return nil
}
