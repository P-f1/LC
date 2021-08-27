/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package yamldeployer

import (
	b64 "encoding/base64"
	"errors"
	"fmt"

	"strings"
	"sync"

	"git.tibco.com/git/product/ipaas/wi-contrib.git/connection/generic"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/P-f1/LC/labs-devops/lib/k8s"
	"github.com/P-f1/LC/labs-devops/lib/util"
)

var log = logger.GetLogger("tibco-k8s-yaml-deployer")

var initialized bool = false

const (
	cClusterKey     = "Cluster"
	iCommand 		= "Command"
	iKind			= "Kind"
	iName 			= "Name"
	iYAMLDescriptor = "YAMLDescriptor"
	oResponse 		= "Response"
	oErrorMessage 	= "ErrorMessage"
)

type YMALDeployActivity struct {
	metadata *activity.Metadata
	managers map[string]*k8s.DeployManager
	mux      sync.Mutex
}

func NewActivity(metadata *activity.Metadata) activity.Activity {
	aYMALDeployActivity := &YMALDeployActivity{
		metadata: metadata,
		managers: make(map[string]*k8s.DeployManager),
	}

	return aYMALDeployActivity
}

func (a *YMALDeployActivity) Metadata() *activity.Metadata {
	return a.metadata
}

func (a *YMALDeployActivity) Eval(context activity.Context) (done bool, err error) {
	log.Info("[YMALDeployActivity:Eval] entering ........ ")

	command, ok := context.GetInput(iCommand).(string)
	if !ok {
		return false, errors.New("Invalid YAML Descriptor ... ")
	}

	kind, ok := context.GetInput(iKind).(string)
	if !ok {
		return false, errors.New("Invalid YAML Descriptor ... ")
	}

	name, ok := context.GetInput(iName).(string)
	if !ok {
		return false, errors.New("Invalid YAML Descriptor ... ")
	}

	yaml, ok := context.GetInput(iYAMLDescriptor).(string)
	if !ok {
		return false, errors.New("Invalid YAML Descriptor ... ")
	}

	manager, err := a.getManager(context)
	if nil != err {
		log.Info("[YMALDeployActivity:Eval] Error : ", err)
		return false, nil
	}

	var status map[string]interface{}
	switch command {
		case "create" : {
			status, err = manager.Create(kind, yaml)
		}
		case "delete" : {
			status, err = manager.Delete(kind, name)
		}
		case "update" : {
			status, err = manager.Update(kind, yaml)
		}
		case "list" : {
			status, err = manager.List(kind)
		}
		default : {
			err = errors.New(fmt.Sprintf("Illegal command : %s", command))
		}
	}

	if outputErr := prepareOutput(context, status, err); nil != outputErr {
		return false, outputErr
	}

	log.Info("[YMALDeployActivity:Eval] done ........ ")
	return true, nil
}

func prepareOutput(context activity.Context, deployments map[string]interface{}, err error) error {
	log.Debug("[YMALDeployActivity:Eval] deployments : ", deployments)

	deploymentArray := make([]interface{}, 0)
	var errorMessage string
	if nil != err {
		errorMessage = err.Error()
	} else {
		if nil != deployments {
			for _, deployment := range deployments {
				deploymentArray = append(deploymentArray, deployment)
			}
			errorMessage = ""
		} else {
			return fmt.Errorf("unknown error")
		}
	}

	response := map[string]interface{}{
		"responses": deploymentArray,
	}
	context.SetOutput("Response", response)
	context.SetOutput("ErrorMessage", errorMessage)

	return nil
}

func (a *YMALDeployActivity) getManager(context activity.Context) (*k8s.DeployManager, error) {
	log.Info("[YMALDeployActivity:getManager] entering ........ ")

	myId := util.ActivityId(context)
	manager := a.managers[myId]
	if nil == manager {
		a.mux.Lock()
		defer a.mux.Unlock()
		manager = a.managers[myId]
		if nil == manager {
			clusterObj, exist := context.GetSetting(cClusterKey)
			if !exist {
				return nil, activity.NewError("Cluster is not configured", "Deployer-4002", nil)
			}

			genericConn, err := generic.NewConnection(clusterObj)
			if err != nil {
				return nil, err
			}

			var namespace string
			var kubeconfigPath string
			var template []byte
			for name, value := range genericConn.Settings() {
				switch name {
				case "template":
					modelcontent, _ := data.CoerceToObject(value)
					template, _ = b64.StdEncoding.DecodeString(strings.Split(modelcontent["content"].(string), ",")[1])
				case "namespace", "Namespace":
					namespace = value.(string)
				case "kubeconfigPath":
					kubeconfigPath = value.(string)
				}
			}

			if nil == template {
				return nil, activity.NewError("Unable to get model string", "GRAPHBUILDER-4004", nil)
			}

			manager, err = k8s.NewDeployManager(kubeconfigPath, namespace, template)
			if err != nil {
				return nil, err
			}

			a.managers[myId] = manager
		}
	}

	log.Info("[YMALDeployActivity:getManager] done ........ ")
	return manager, nil
}
