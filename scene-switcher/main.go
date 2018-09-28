package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/caseymrm/go-obs-websocket"
	ui "github.com/gizak/termui"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	Verbose  bool   `short:"v" long:"verbose" description:"Make verbose output"`
	Address  string `short:"a" long:"address" description:"Address of the OBS websocket" default:"localhost"`
	Port     int    `short:"p" long:"port" description:"Port of the OBS websocket" default:"4444"`
	Password string `short:"w" long:"password" description:"Password for the OBS websocket" default:""`
}

type MyUI struct {
	Echo   *EchoArea
	Info   *ui.Par
	Scenes *ui.List
}

var myui MyUI

func SetUpUI() {
	log.Printf("Creating UI")
	myui.Info = ui.NewPar("Press Q to quit")
	myui.Info.Height = 3
	myui.Info.TextFgColor = ui.ColorWhite
	myui.Info.BorderLabel = "Info"

	myui.Echo = NewEchoArea(10)
	myui.Echo.Label = "Logs"

	myui.Scenes = ui.NewList()
	myui.Scenes.BorderLabel = "Scenes"

	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(3, 0, myui.Info),
			ui.NewCol(9, 0, myui.Scenes)),
		ui.NewRow(
			ui.NewCol(12, 0, myui.Echo)))

	ui.Body.Align()

	ui.Render(ui.Body)

	ui.Handle("/sys/kbd/q", func(ui.Event) {
		log.Printf("Exiting")
		ui.StopLoop()
	})

}

var logfile *os.File

func SetupLog() error {
	filename := fmt.Sprintf("%s/obs-scene-switcher.%d.log", os.TempDir(), os.Getpid())
	var err error
	logfile, err = os.Create(filename)
	if err != nil {
		return err
	}
	log.SetOutput(io.MultiWriter(logfile, os.Stderr))
	return nil
}

func Execute() error {
	var opts Options
	if _, err := flags.Parse(&opts); err != nil {
		//mask help wanted
		if ferr, ok := err.(*flags.Error); ok == true && ferr.Type == flags.ErrHelp {
			return nil
		}
		return err
	}

	if err := ui.Init(); err != nil {
		return err
	}
	defer ui.Close()

	SetupLog()
	defer func() {
		log.SetOutput(os.Stderr)
		logfile.Close()
	}()

	SetUpUI()

	if opts.Verbose == true {
		log.SetOutput(io.MultiWriter(logfile, myui.Echo))
	} else {
		log.SetOutput(logfile)
	}

	log.Printf("Connecting to %s:%d", opts.Address, opts.Port)
	c, err := obsws.NewClient(opts.Address, opts.Port, opts.Password)
	if err != nil {
		return err
	}

	go func() {
		events := c.EventChannel()
		for e := range events {
			log.Printf("Received event: %v", e)
		}
	}()

	resp, err := c.GetSceneList()
	if err != nil {
		return err
	}

	for i, s := range resp.Scenes {
		if i >= 10 {
			break
		}
		key := (i + 1) % 10
		name := s.Name
		myui.Scenes.Items = append(myui.Scenes.Items, fmt.Sprintf("[%d] %s", key, name))
		eventaddress := fmt.Sprintf("/sys/kbd/%d", key)
		log.Printf("Found scene %d:%s, handling it with %s", key, name, eventaddress)
		ui.Handle(eventaddress,
			func(ui.Event) {

				log.Printf("Switching to scene '%s', %s", name, eventaddress)
				err := c.SetCurrentScene(name)
				if err != nil {
					log.Printf("Could not change to  scene '%s': %s", name, err)
				}
			})
	}
	myui.Scenes.Height = len(myui.Scenes.Items) + 2
	myui.Info.Height = len(myui.Scenes.Items) + 2
	ui.Body.Align()
	ui.Render(ui.Body)
	log.Print("Looping ui")
	ui.Loop()

	return nil
}

func main() {
	if err := Execute(); err != nil {
		log.Printf("Got unhandled error: %s", err)
		os.Exit(1)
	}
}
