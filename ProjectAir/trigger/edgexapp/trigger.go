/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package edgexapp

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"

	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
)

const (
	cPort      = "Port"
	cPath      = "Path"
	serviceKey = "RspControllerEventHandlerApp"
)

//-============================================-//
//   Entry point register Trigger & factory
//-============================================-//

var triggerMd = trigger.NewMetadata(&Settings{}, &HandlerSettings{}, &Output{})

func init() {
	_ = trigger.Register(&EdgeXApp{}, &Factory{})
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

	return &EdgeXApp{settings: settings}, nil
}

//-=========================-//
//      Define Trigger
//-=========================-//

var logger log.Logger

type EdgeXApp struct {
	metadata *trigger.Metadata
	config   *trigger.Config
	mux      sync.Mutex

	settings *Settings
	handlers []trigger.Handler
}

// implements trigger.Initializable.Initialize
func (this *EdgeXApp) Initialize(ctx trigger.InitContext) error {
	this.handlers = ctx.GetHandlers()
	logger = ctx.Logger()

	return nil
}

// implements ext.Trigger.Start
func (this *EdgeXApp) Start() error {
	logger.Info("(Start) Processing handlers")
	for _, handler := range this.handlers {
		handlerSetting := &HandlerSettings{}
		err := metadata.MapToStruct(handler.Settings(), handlerSetting, true)
		if err != nil {
			return err
		}

		go func() {
			err = this.register()
			if err != nil {
				logger.Error("ListenAndServe: ", err)
			}
		}()
		logger.Info("(Start) Started path = ", handlerSetting.Path, ", port = ", this.settings.Port)
	}
	logger.Info("(Start) Now started")

	return nil
}

// implements ext.Trigger.Stop
func (this *EdgeXApp) Stop() error {
	return nil
}

func (this *EdgeXApp) register() error {

	edgexSdk := &appsdk.AppFunctionsSDK{ServiceKey: serviceKey}
	if err := edgexSdk.Initialize(); err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("SDK initialization failed: %v\n", err))
		return err
	}

	appSettings := edgexSdk.ApplicationSettings()
	if appSettings == nil {
		edgexSdk.LoggingClient.Error("No application settings found")
		return errors.New("No application settings found")
	}

	deviceNames, ok := appSettings["DeviceNames"]
	if !ok {
		edgexSdk.LoggingClient.Error("DeviceNames application setting not found")
		return errors.New("DeviceNames application setting not found")
	}
	deviceNamesList := []string{deviceNames}
	edgexSdk.LoggingClient.Info(fmt.Sprintf("Running the application functions for %v devices...", deviceNamesList))

	valueDescriptor, ok := appSettings["ValueDescriptorToFilter"]
	if !ok {
		edgexSdk.LoggingClient.Error("ValueDescriptorToFilter application setting not found")
		return errors.New("ValueDescriptorToFilter application setting not found")
	}
	valueDescriptorList := []string{valueDescriptor}

	edgexSdk.SetFunctionsPipeline(
		transforms.NewFilter(deviceNamesList).FilterByDeviceName,
		transforms.NewFilter(valueDescriptorList).FilterByValueDescriptor,
		this.ProcessEdgeXEvents,
	)
	edgexSdk.MakeItRun()
	return nil
}

func (t *EdgeXApp) ProcessEdgeXEvents(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	t.mux.Lock()
	defer t.mux.Unlock()
	outputData := &Output{}
	outputData.Event = params[0].(map[string]interface{})

	logger.Info("(EdgeXApp.ProcessEdgeXEvents) - edgexcontext : ", params)
	logger.Debug("(EdgeXApp.ProcessEdgeXEvents) - params : ", params)

	_, err := t.handlers[0].Handle(context.Background(), outputData)

	if nil != err {
		logger.Errorf("Run action for handler [%s] failed for reason [%s] message lost", t.handlers[0], err)
	}
	logger.Infof("(FileContentHandler.HandleContent) - Trigger done for content")
	return true, nil
}
