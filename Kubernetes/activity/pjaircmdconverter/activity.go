/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package pjaircmdconverter

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/logger"
)

var log = logger.GetLogger("tibco-activity-projair-cmdconverter")

var initialized bool = false

const (
	iCommand        = "Command"
	iApplicationID  = "ApplicationID"
	iContainerName  = "ContainerName"
	iDockerImage    = "DockerImage"
	iReplicas       = "Replicas"
	iComponents     = "Components"
	CMD_Deploy      = "deploy"
	CMD_Update      = "update"
	CMD_ListDeploys = "list"
	CMD_Undeploy    = "undeploy"
)

type ProjAirCMDConverterActivity struct {
	metadata *activity.Metadata
	mux      sync.Mutex
}

func NewActivity(metadata *activity.Metadata) activity.Activity {
	aProjAirCMDConverterActivity := &ProjAirCMDConverterActivity{
		metadata: metadata,
	}

	return aProjAirCMDConverterActivity
}

func (a *ProjAirCMDConverterActivity) Metadata() *activity.Metadata {
	return a.metadata
}

func (a *ProjAirCMDConverterActivity) Eval(context activity.Context) (done bool, err error) {

	log.Info("[ProjAirCMDConverterActivity:Eval] entering ........ ")

	command, ok := context.GetInput(iCommand).(string)
	if !ok {
		return false, errors.New("Invalid command ... ")
	}

	applicationID, ok := context.GetInput(iApplicationID).(string)
	if !ok {
		return false, errors.New("Invalid applicationID ... ")
	}

	containerName, ok := context.GetInput(iContainerName).(string)
	if !ok {
		containerName = "air-pipeline"
	}

	dockerImage, ok := context.GetInput(iDockerImage).(string)
	if !ok {
		dockerImage = "bigoyang/dump-env"
	}

	replicas, ok := context.GetInput(iReplicas).(int64)
	if !ok {
		replicas = 1
	}

	components := context.GetInput(iComponents).(*data.ComplexObject).Value

	log.Info("command : ", command)
	log.Info("applicationID : ", applicationID)
	log.Info("components : ", components)

	var deploymentName string
	env := make([]map[string]interface{}, 0)
	env = append(env, map[string]interface{}{
		"Name":  "FLOGO_APP_PROPS_ENV",
		"Value": "auto",
	})
	if nil != components {
		componentArray, ok := components.([]interface{})
		if ok {
			var protocol string
			var datastore string
			for _, component := range componentArray {
				componentObj := component.(map[string]interface{})
				componentName := componentObj["Name"].(string)
				componentType := componentObj["Type"].(string)
				switch componentType {
				case "protocol":
					protocol = componentName
				case "datastore":
					datastore = componentName
				}
				if nil != componentObj["Properties"] {
					for _, property := range componentObj["Properties"].([]interface{}) {
						propertyObj := property.(map[string]interface{})
						propertyName := propertyObj["Name"].(string)
						if strings.HasPrefix(propertyName, ".") {
							propertyName = fmt.Sprintf("%s%s", componentName, propertyName)
						}

						env = append(env, map[string]interface{}{
							"Name":  strings.Replace(propertyName, ".", "_", -1),
							"Value": propertyObj["Value"],
						})
					}
				}
			}
			deploymentName = strings.ToLower(fmt.Sprintf("%s.%s.%s", protocol, datastore, applicationID))
		}
	}
	deploment := map[string]interface{}{
		"Command":    command,
		"Deployment": deploymentName,
		"Replicas":   replicas,
		"Containers": map[string]interface{}{
			"Name":  containerName,
			"Image": dockerImage,
			"Ports": map[string]interface{}{
				"ContainerPort": nil,
			},
			"Env": env,
		},
	}

	log.Info("[ProjAirCMDConverterActivity:Eval] deploment = ", deploment)

	complexObj := &data.ComplexObject{Metadata: "DeployDescriptor", Value: deploment}
	context.SetOutput("DeployDescriptor", complexObj)
	log.Info("[ProjAirCMDConverterActivity:Eval] Exit ........ ")

	return true, nil
}
