/*
 * Copyright © 2019. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package videoserver

import (
	"context"
	b64 "encoding/base64"
	"strings"
	"sync"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/labs-mr4r/lib/internet/streamserver"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
)

const (
	cConnection     = "Connection"
	cConnectionName = "name"
)

//-============================================-//
//   Entry point register Trigger & factory
//-============================================-//

var triggerMd = trigger.NewMetadata(&Settings{}, &HandlerSettings{}, &Output{})

func init() {
	_ = trigger.Register(&VideoServer{}, &Factory{})
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

	return &VideoServer{settings: settings}, nil
}

//-=========================-//
//      Define Trigger
//-=========================-//

var logger log.Logger

type VideoServer struct {
	metadata *trigger.Metadata
	config   *trigger.Config
	server   *streamserver.Server
	mux      sync.Mutex

	settings *Settings
	handlers []trigger.Handler
}

// implements trigger.Initializable.Initialize
func (this *VideoServer) Initialize(ctx trigger.InitContext) error {

	this.handlers = ctx.GetHandlers()
	logger = ctx.Logger()

	return nil
}

// implements ext.Trigger.Start
func (this *VideoServer) Start() error {

	logger.Info("Start")
	handlers := this.handlers

	logger.Info("Processing handlers")

	connection, exist := handlers[0].Settings()[cConnection]
	if !exist {
		return activity.NewError("Video connection is not configured", "TGDB-Video-4001", nil)
	}

	connectionInfo, _ := data.CoerceToObject(connection)
	if connectionInfo == nil {
		return activity.NewError("Video connection not able to be parsed", "TGDB-Video-4002", nil)
	}

	var serverId string
	properties := make(map[string]interface{})
	connectionSettings, _ := connectionInfo["settings"].([]interface{})
	if connectionSettings != nil {
		for _, v := range connectionSettings {
			setting, err := data.CoerceToObject(v)
			if nil != err {
				continue
			}

			if nil != setting {
				var err error
				if setting["name"] == streamserver.ServerPort {
					properties[streamserver.ServerPort], err = data.CoerceToString(setting["value"])
					if nil != err {
						return err
					}
				} else if setting["name"] == cConnectionName {
					serverId = setting["value"].(string)
				} else if setting["name"] == streamserver.ConnectionPath {
					properties[streamserver.ConnectionPath], err = data.CoerceToString(setting["value"])
					if nil != err {
						return err
					}
				} else if setting["name"] == streamserver.ConnectionTlsEnabled {
					properties[streamserver.ConnectionTlsEnabled], err = data.CoerceToBoolean(setting["value"])
					if nil != err {
						return err
					}
				} else if setting["name"] == streamserver.ConnectionUploadCRT {
					properties[streamserver.ConnectionUploadCRT], err = data.CoerceToBoolean(setting["value"])
					if nil != err {
						return err
					}
				} else if setting["name"] == streamserver.ConnectionTlsCRT {
					tlsCRT, err := data.CoerceToObject(setting["value"])
					if nil != err {
						properties[streamserver.ConnectionTlsCRT], _ = b64.StdEncoding.DecodeString(strings.Split(tlsCRT["content"].(string), ",")[1])
						properties[streamserver.ConnectionTlsCRTPath], _ = tlsCRT["filename"].(string)
					}
				} else if setting["name"] == streamserver.ConnectionTlsKey {
					tlsKey, err := data.CoerceToObject(setting["value"])
					if nil != err {
						properties[streamserver.ConnectionTlsKey], _ = b64.StdEncoding.DecodeString(strings.Split(tlsKey["content"].(string), ",")[1])
						properties[streamserver.ConnectionTlsKeyPath], _ = tlsKey["filename"].(string)
					}
				} else if setting["name"] == streamserver.ConnectionTlsCRTPath {
					if nil != err {
						properties[streamserver.ConnectionTlsCRTPath], err = data.CoerceToString(setting["value"])
					}
				} else if setting["name"] == streamserver.ConnectionTlsKeyPath {
					if nil != err {
						properties[streamserver.ConnectionTlsKeyPath], err = data.CoerceToString(setting["value"])
					}
				}
			}

		}
		logger.Debug(properties)

		var err error
		this.server, err = streamserver.GetFactory().CreateServer(serverId, properties, this)
		if nil != err {
			return err
		}
		logger.Info("Server = ", *this.server)
		go this.server.Start()
	}

	return nil
}

// implements ext.Trigger.Stop
func (this *VideoServer) Stop() error {
	this.server.Stop()
	return nil
}

func (this *VideoServer) ProcessRequest(request string) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	logger.Debug("Got Video Request : ", request)
	outputData := &Output{}
	outputData.Request = request
	logger.Debug("Send Video Request out : ", outputData)

	_, err := this.handlers[0].Handle(context.Background(), outputData)
	if nil != err {
		logger.Info("Error -> ", err)
	}

	return err
}
