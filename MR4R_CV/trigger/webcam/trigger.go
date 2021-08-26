/*
 * Copyright © 2019. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package webcam

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

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
	logger.Debug("Stopping drone")
	return nil
}

func (this *WebcamVideoReceiver) readFromWebcam() error {
	// open webcam
	var err error
	this.webcam, err = gocv.OpenVideoCapture(this.deviceID)
	if err != nil {
		logger.Errorf("Error opening capture device: %v\n", this.deviceID)
		return err
	}
	defer this.webcam.Close()

	img := gocv.NewMat()
	defer img.Close()

	for {
		logger.Info("(WebcamVideoReceiver.readFromWebcam) begin .............")
		if ok := this.webcam.Read(&img); !ok {
			return errors.New(fmt.Sprintf("Device closed: %v\n", this.deviceID))
		}
		if img.Empty() {
			continue
		}

		this.OnVideoData(img.ToBytes(), img.Cols(), img.Rows())
	}

	return nil
}
