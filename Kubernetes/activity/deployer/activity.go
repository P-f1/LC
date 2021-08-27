/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package deployer

import (
	b64 "encoding/base64"
	"encoding/json"
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

var log = logger.GetLogger("tibco-activity-k8s-deployer")

var initialized bool = false

const (
	cContextKey     = "Context"
	iCommand        = "Command"
	iNamespace      = "Namespace"
	iDeployGpID     = "DeployGpID"
	iDeployplan     = "Deployplan"
	iDeployments    = "Deployments"
	CMD_Deploy      = "deploy"
	CMD_Update      = "update"
	CMD_Ping        = "ping"
	CMD_ListDeploys = "list"
	CMD_Undeploy    = "undeploy"
	CMD_UndeployAll = "undeployAll"
	CMD_Test        = "test"
)

type KubernetesDeployActivity struct {
	metadata *activity.Metadata
	managers map[string]*k8s.DeployManager
	mux      sync.Mutex
}

func NewActivity(metadata *activity.Metadata) activity.Activity {
	aKubernetesDeployActivity := &KubernetesDeployActivity{
		metadata: metadata,
		managers: make(map[string]*k8s.DeployManager),
	}

	return aKubernetesDeployActivity
}

func (a *KubernetesDeployActivity) Metadata() *activity.Metadata {
	return a.metadata
}

func (a *KubernetesDeployActivity) Eval(context activity.Context) (done bool, err error) {
	log.Info("[KubernetesDeployActivity:Eval] entering ........ ")

	command, ok := context.GetInput(iCommand).(string)
	if !ok {
		return false, errors.New("Invalid command ... ")
	}

	namespace, _ := context.GetInput(iNamespace)

	var deployGpID string
	var deployplan map[string]interface{}
	var deployments []map[string]interface{}
	if CMD_UndeployAll != command && CMD_ListDeploys != command {
		deployGpID, ok = context.GetInput(iDeployGpID).(string)
		if !ok {
			log.Warn("Invalid DeployGpID ... ")
		}

		deployplan, ok = context.GetInput(iDeployplan).(*data.ComplexObject).Value.(map[string]interface{})
		if !ok {
			log.Warn("No deploy plane ... ")
		}

		if CMD_Undeploy != command {
			deployments = context.GetInput(iDeployments).(*data.ComplexObject).Value.([]map[string]interface{})
		}
	}

	manager, err := a.getManager(context)
	if nil != err {
		log.Info("[KubernetesDeployActivity:Eval] Error : ", err)
		return false, nil
	}

	log.Info("[KubernetesDeployActivity:Eval] command : ", command)
	log.Info("[KubernetesDeployActivity:Eval] deployGpID : ", deployGpID)
	log.Info("[KubernetesDeployActivity:Eval] deployplan : ", deployplan)
	log.Info("[KubernetesDeployActivity:Eval] deployments : ", deployments)

	var status map[string]interface{}
	switch command {
	case CMD_Deploy:
		status, err = manager.GroupDeploy(namespace, deployGpID, deployplan, deployments)
	case CMD_Update:
		//status, err = a.update(deployment, replicas, containers, cluster)
	case CMD_Ping:
		//status, err = a.ping(deployment, cluster)
	case CMD_ListDeploys:
		status, err = manager.ListDeploys(namespace)
	case CMD_Undeploy:
		//status, err = a.undeploy(deployment, cluster)
	case CMD_UndeployAll:
		status, err = manager.UndeployAll(namespace)
	}

	if outputErr := prepareOutput(context, status, err); nil != outputErr {
		return false, outputErr
	}

	log.Info("[KubernetesDeployActivity:Eval] done ........ ")
	return true, nil
}

func prepareOutput(context activity.Context, deployments map[string]interface{}, err error) error {
	log.Debug("[KubernetesDeployActivity:Eval] deployments : ", deployments)

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

	jsonDeployments, _ := json.Marshal(
		map[string]interface{}{
			"response": deploymentArray,
		},
	)
	context.SetOutput("JSONDeployments", string(jsonDeployments))
	context.SetOutput("ErrorMessage", errorMessage)

	return nil
}

func (a *KubernetesDeployActivity) getManager(context activity.Context) (*k8s.DeployManager, error) {
	log.Info("[KubernetesDeployActivity:getManager] entering ........ ")

	myId := util.ActivityId(context)
	manager := a.managers[myId]
	if nil == manager {
		a.mux.Lock()
		defer a.mux.Unlock()
		manager = a.managers[myId]
		if nil == manager {
			contextObj, exist := context.GetSetting(cContextKey)
			if !exist {
				return nil, activity.NewError("Context is not configured", "Deployer-4002", nil)
			}

			genericConn, err := generic.NewConnection(contextObj)
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

	log.Info("[KubernetesDeployActivity:getManager] done ........ ")
	return manager, nil
}
