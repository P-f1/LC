/*
 * Copyright © 2019. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package vstreamendpoint

import (
	"errors"
	"reflect"
	"sync"

	"github.com/P-f1/LC/flogo-lib/core/activity"
	"github.com/P-f1/LC/flogo-lib/core/data"
	"github.com/P-f1/LC/flogo-lib/logger"
	"github.com/P-f1/LC/labs-mr4r/lib/internet/streamserver"
	"github.com/P-f1/LC/labs-mr4r/lib/util"
)

var log = logger.GetLogger("activity-mr4r-vsendpoint")

const (
	cConnection     = "Connection"
	cConnectionName = "name"
	input           = "Video"
	streamId        = "StreamId"
)

type VideoEndPoint struct {
	metadata            *activity.Metadata
	activityToConnector map[string]string
	mux                 sync.Mutex
}

func NewActivity(metadata *activity.Metadata) activity.Activity {
	aCSVParserActivity := &VideoEndPoint{
		metadata:            metadata,
		activityToConnector: make(map[string]string),
	}
	return aCSVParserActivity
}

func (a *VideoEndPoint) Metadata() *activity.Metadata {
	return a.metadata
}

func (a *VideoEndPoint) Eval(ctx activity.Context) (done bool, err error) {

	broker, err := a.getVideoBroker(ctx)

	if nil != err {
		return false, err
	}

	streamId, validId := ctx.GetInput(streamId).(string)
	if !validId {
		log.Warn("Invalid stream id, expecting string type but get : ", reflect.TypeOf(ctx.GetInput("StreamId")).String())
		streamId = "*"
	}

	data, ok := ctx.GetInput(input).(string)
	if !ok {
		log.Warn("Invalid data, expecting string but get : ", ctx.GetInput("Video"))
	}

	if "" != data {
		bytes := []byte(data)
		log.Debug("Serialized data : ", bytes)

		broker.SendData(streamId, bytes)
	}

	return true, nil
}

func (a *VideoEndPoint) getVideoBroker(context activity.Context) (*streamserver.Server, error) {
	myId := util.ActivityId(context)

	videoBroker := streamserver.GetFactory().GetServer(a.activityToConnector[myId])
	if nil == videoBroker {
		log.Info("Look up Video data broker start ...")
		connection, exist := context.GetSetting(cConnection)
		if !exist {
			return nil, activity.NewError("Video connection is not configured", "MR4R-Video-4001", nil)
		}

		connectionInfo, _ := data.CoerceToObject(connection)
		if connectionInfo == nil {
			return nil, activity.NewError("Video connection not able to be parsed", "MR4R-Video-4002", nil)
		}

		var connectorName string
		connectionSettings, _ := connectionInfo["settings"].([]interface{})
		if connectionSettings != nil {
			for _, v := range connectionSettings {
				setting, _ := data.CoerceToObject(v)
				if setting != nil {
					if setting["name"] == cConnectionName {
						connectorName, _ = data.CoerceToString(setting["value"])
					}
				}
			}
			videoBroker = streamserver.GetFactory().GetServer(connectorName)
			if nil == videoBroker {
				return nil, errors.New("Broker not found, connection id = " + connectorName)
			}
			a.activityToConnector[myId] = connectorName
		}
		log.Info("Look up Video data broker end ...")
	}

	return videoBroker, nil
}
