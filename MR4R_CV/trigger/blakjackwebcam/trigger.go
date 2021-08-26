/*
 * Copyright © 2019. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package webcam

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/blackjack/webcam"

	"github.com/P-f1/LC/flogo-lib/core/activity"
	"github.com/P-f1/LC/flogo-lib/core/data"

	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"

	"gocv.io/x/gocv"
)

const (
	DeviceID = "DeviceID"
)

//-============================================-//
//   Entry point register Trigger & factory
//-============================================-//

var triggerMd = trigger.NewMetadata(&Settings{}, &HandlerSettings{}, &Output{})

func init() {
	_ = trigger.Register(&WebcamVideoReceiver{}, &Factory{})
}

//-===============================-//
//     Define Trigger Factory
//-===============================-//

type Factory struct {
}

// Metadata implements trigger.Factory.Metadata
func (*Factory) Metadata() *trigger.Metadata {
	return triggerMd
}

// New implements trigger.Factory.New
func (*Factory) New(config *trigger.Config) (trigger.Trigger, error) {
	settings := &Settings{}
	err := metadata.MapToStruct(config.Settings, settings, true)
	if err != nil {
		return nil, err
	}

	return &WebcamVideoReceiver{settings: settings}, nil
}

//-=========================-//
//      Define Trigger
//-=========================-//

var logger log.Logger

type WebcamVideoReceiver struct {
	deviceID string
	webcam   *gocv.VideoCapture
	metadata *trigger.Metadata
	config   *trigger.Config
	mux      sync.Mutex

	settings *Settings
	handlers []trigger.Handler
}

// implements trigger.Initializable.Initialize
func (this *WebcamVideoReceiver) Initialize(ctx trigger.InitContext) error {

	this.handlers = ctx.GetHandlers()
	logger = ctx.Logger()

	return nil
}

// implements ext.Trigger.Start
func (this *WebcamVideoReceiver) Start() error {

	logger.Info("(WebcamVideoReceiver.Start) .....")
	handlers := this.handlers

	var err error
	this.deviceID, err = data.CoerceToString(handlers[0].Settings()[DeviceID])
	if nil != err {
		return activity.NewError("Frame rate is not configured", "TELLO-IR-4004", nil)
	}

	logger.Info("FLOGO_HOME = ", os.Getenv("FLOGO_HOME"))

	go this.readFromWebcam()

	return nil
}

func (this *WebcamVideoReceiver) OnVideoData(video []byte, frameWidth int, frameHeight int) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	logger.Info("(WebcamVideoReceiver.OnImage) Entering, frameWidth = ", frameWidth, ", frameHeight = ", frameHeight)

	_, err := this.handlers[0].Handle(
		context.Background(),
		&Output{
			VideoFrame:        string(video),
			OutputFrameWidth:  int(frameWidth),
			OutputFrameHeight: int(frameHeight),
		},
	)
	if nil != err {
		logger.Info("Error -> ", err)
	}

	logger.Info("(WebcamVideoReceiver.OnImage) Exit .............")

	return err
}

func (this *WebcamVideoReceiver) OnClose() error {
	return this.Stop()
}

// implements ext.Trigger.Stop
func (this *WebcamVideoReceiver) Stop() error {
	logger.Debug("Stopping camera")
	return nil
}

func (this *WebcamVideoReceiver) readFromWebcam() error {
	cam, err := webcam.Open(this.deviceID)
	if err != nil {
		panic(err.Error())
	}
	defer cam.Close()

	format_desc := cam.GetSupportedFormats()
	var formats []webcam.PixelFormat
	for f := range format_desc {
		formats = append(formats, f)
	}

	format := formats[0]
	frames := FrameSizes(cam.GetSupportedFrameSizes(format))
	sort.Sort(frames)
	size := frames[0]

	f, w, h, err := cam.SetImageFormat(format, uint32(size.MaxWidth), uint32(size.MaxHeight))

	if err != nil {
		panic(err.Error())
	} else {
		fmt.Fprintf(os.Stderr, "Resulting image format: %s (%dx%d)\n", format_desc[f], w, h)
	}

	err = cam.StartStreaming()
	if err != nil {
		panic(err.Error())
	}

	timeout := uint32(5) //5 seconds
	width := int(w)
	height := int(h)
	for {
		err = cam.WaitForFrame(timeout)

		switch err.(type) {
		case nil:
		case *webcam.Timeout:
			fmt.Fprint(os.Stderr, err.Error())
			continue
		default:
			panic(err.Error())
		}

		frame, err := cam.ReadFrame()
		if len(frame) != 0 {
			this.OnVideoData(frame, width, height)
		} else if err != nil {
			panic(err.Error())
		}
	}
}

func readChoice(s string) int {
	var i int
	for true {
		print(s)
		_, err := fmt.Scanf("%d\n", &i)
		if err != nil || i < 1 {
			println("Invalid input. Try again")
		} else {
			break
		}
	}
	return i
}

type FrameSizes []webcam.FrameSize

func (slice FrameSizes) Len() int {
	return len(slice)
}

//For sorting purposes
func (slice FrameSizes) Less(i, j int) bool {
	ls := slice[i].MaxWidth * slice[i].MaxHeight
	rs := slice[j].MaxWidth * slice[j].MaxHeight
	return ls < rs
}

//For sorting purposes
func (slice FrameSizes) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
