/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package K8smonitor

import (
	"context"
	b64 "encoding/base64"
	"strings"
	"sync"

	"github.com/P-f1/LC/flogo-lib/core/activity"
	"github.com/P-f1/LC/flogo-lib/core/data"
	"github.com/P-f1/LC/labs-devops/lib/k8s"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
)

const (
	cClusterKey     = "Cluster"
	cConnectionName = "name"
)

//-============================================-//
//   Entry point register Trigger & factory
//-============================================-//

var triggerMd = trigger.NewMetadata(&Settings{}, &HandlerSettings{}, &Output{})

func init() {
	_ = trigger.Register(&K8sMonitor{}, &Factory{})
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

	return &K8sMonitor{settings: settings}, nil
}

//-=========================-//
//      Define Trigger
//-=========================-//

var logger log.Logger

type K8sMonitor struct {
	metadata *trigger.Metadata
	config   *trigger.Config
	mux      sync.Mutex

	settings *Settings
	handlers []trigger.Handler
	cluster  *k8s.Cluster
}

// implements trigger.Initializable.Initialize
func (this *K8sMonitor) Initialize(ctx trigger.InitContext) error {

	this.handlers = ctx.GetHandlers()
	logger = ctx.Logger()

	return nil
}

// implements ext.Trigger.Start
func (this *K8sMonitor) Start() error {

	logger.Info("Start")
	handlers := this.handlers

	logger.Info("Processing handlers")

	clusterObj, exist := handlers[0].Settings()[cClusterKey]
	if !exist {
		return activity.NewError("Cluster is not configured", "K8sMonitor-4002", nil)
	}

	//Read cluster details
	clusterInfo, _ := data.CoerceToObject(clusterObj)
	if clusterInfo == nil {
		return activity.NewError("Unable extract model", "K8sMonitor-4001", nil)
	}

	var namespace string
	var kubeconfigPath string
	var template []byte
	connectionSettings, _ := clusterInfo["settings"].([]interface{})
	if connectionSettings != nil {
		for _, v := range connectionSettings {
			setting, err := data.CoerceToObject(v)
			if nil != err {
				continue
			}

			if nil != setting {
				if setting["name"] == "template" {
					modelcontent, _ := data.CoerceToObject(setting["value"])
					template, _ = b64.StdEncoding.DecodeString(strings.Split(modelcontent["content"].(string), ",")[1])
				} else if setting["name"] == "name" {
					namespace = setting["value"].(string)
				} else if setting["name"] == "kubeconfigPath" {
					kubeconfigPath = setting["value"].(string)
				}
			}

		}

		if nil == template {
			return activity.NewError("Unable to get model string", "K8sMonitor-4004", nil)
		}

		var err error
		this.cluster, err = k8s.NewCluster(kubeconfigPath, namespace)
		if err != nil {
			return err
		}

		this.cluster.ListenEvent(this)
	}

	return nil
}

// implements ext.Trigger.Stop
func (this *K8sMonitor) Stop() error {
	this.cluster.StopEvent()
	this.cluster = nil
	return nil
}

func (this *K8sMonitor) HandleEvent(event string) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	logger.Info("Got event : ", event)
	outputData := &Output{}
	outputData.JSONEvent = event
	logger.Debug("Send event out : ", outputData)

	_, err := this.handlers[0].Handle(context.Background(), outputData)
	if nil != err {
		logger.Info("Error -> ", err)
	}

	return err
}

func (this *K8sMonitor) NoMoreEvent(resume bool) {
	logger.Info("k8s monitor closed, no more k8s event!")
}
