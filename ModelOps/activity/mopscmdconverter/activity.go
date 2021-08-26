/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package mopscmdconverter

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/P-f1/LC/flogo-lib/core/activity"
	"github.com/P-f1/LC/flogo-lib/core/data"
	"github.com/P-f1/LC/flogo-lib/logger"
	"github.com/P-f1/LC/labs-flogo-lib/util"
)

var log = logger.GetLogger("tibco-model-ops-cmdconverter")

var initialized bool = false

const (
	iCommand        = "Command"
	iEndpointIP     = "EndpointIp"
	iEndpointPort   = "EndpointPort"
	iDeploymentGpID = "DeploymentGpID"
	iDataFlow       = "DataFlow"
	iComponents     = "Components"
	iSystem         = "System"
	oDeployGpID     = "DeployGpID"
	oDeployplan     = "Deployplan"
	oDeployments    = "Deployments"
	CMD_Deploy      = "deploy"
	CMD_Update      = "update"
	CMD_ListDeploys = "list"
	CMD_Undeploy    = "undeploy"
)

type ModelOpsCMDConverterActivity struct {
	metadata     *activity.Metadata
	mux          sync.Mutex
	endpointPort int64
}

func NewActivity(metadata *activity.Metadata) activity.Activity {
	aModelOpsCMDConverterActivity := &ModelOpsCMDConverterActivity{
		metadata:     metadata,
		endpointPort: 10100,
	}

	return aModelOpsCMDConverterActivity
}

func (a *ModelOpsCMDConverterActivity) Metadata() *activity.Metadata {
	return a.metadata
}

func (a *ModelOpsCMDConverterActivity) Eval(context activity.Context) (done bool, err error) {

	log.Info("[ModelOpsCMDConverterActivity:Eval] entering ........ ")

	command, ok := context.GetInput(iCommand).(string)
	if !ok {
		return false, errors.New("Invalid command ... ")
	}

	endpointIP, ok := context.GetInput(iEndpointIP).(string)
	if !ok {
		endpointIP = "Invalid endpointIP"
	}

	//endpointPort, ok := context.GetInput(iEndpointPort).(int64)
	//if !ok {
	//	return false, errors.New("Invalid endpointPort ... ")
	//}

	deploymentGpID, ok := context.GetInput(iDeploymentGpID).(string)
	if !ok {
		return false, errors.New("Invalid deploymentGpID ... ")
	}

	components := context.GetInput(iComponents).(*data.ComplexObject).Value
	system := context.GetInput(iSystem).(*data.ComplexObject).Value

	log.Info("command : ", command)
	log.Info("deploymentGpID : ", deploymentGpID)
	log.Info("components : ", components)

	context.SetOutput(iCommand, command)
	context.SetOutput(oDeployGpID, deploymentGpID)
	if nil != components {
		dataFlow := context.GetInput(iDataFlow).(*data.ComplexObject).Value
		log.Info("[ModelOpsCMDConverterActivity:Eval] dataFlow = ", dataFlow)
		if nil != dataFlow {
			deployplan := map[string]interface{}{
				"Strategy": "parallel",
				"Pipeline": deploymentPlan(downstreamComponents(dataFlow.([]interface{}))),
			}
			log.Info("[ModelOpsCMDConverterActivity:Eval] deployplan = ", deployplan)
			context.SetOutput(oDeployplan, &data.ComplexObject{Metadata: "Deployplan", Value: deployplan})
		}

		deployments, _ := a.buildDeployments(
			deploymentGpID,
			command,
			endpointIP,
			downstreamComponents(dataFlow.([]interface{})),
			system,
			components.([]interface{}))

		log.Info("[ModelOpsCMDConverterActivity:Eval] deployments = ", deployments)

		context.SetOutput(oDeployments, &data.ComplexObject{Metadata: "Deployments", Value: deployments})

		log.Info("[ModelOpsCMDConverterActivity:Eval] Exit ........ ")
	}

	return true, nil
}

func downstreamComponents(dataFlow []interface{}) map[string][]string {
	downstreamComponentsMap := make(map[string][]string)
	for _, flow := range dataFlow {
		me := flow.(map[string]interface{})["Upstream"].(string)
		downstream := flow.(map[string]interface{})["Downstream"].(string)
		/* register downstream component for me */
		downstreamComponents := downstreamComponentsMap[me]
		if nil == downstreamComponents {
			downstreamComponents = make([]string, 0)
		}
		downstreamComponentsMap[me] = append(downstreamComponents, downstream)

		/* also create an entry in the downstreamComponentsMap
		for my downstream component */
		if nil == downstreamComponentsMap[downstream] {
			downstreamComponentsMap[downstream] = make([]string, 0)
		}
	}
	return downstreamComponentsMap
}

func deploymentPlan(downstreamComponentsMap map[string][]string) []map[string]interface{} {
	plan := make([]map[string]interface{}, 0)

	groupID := 0
	for 0 != len(downstreamComponentsMap) {
		group := make([]map[string]interface{}, 0)
		for key, value := range downstreamComponentsMap {
			if 0 == len(value) {
				group = append(group, map[string]interface{}{"Name": key, "Type": ""})
				delete(downstreamComponentsMap, key)
			}
		}

		for key, value := range downstreamComponentsMap {
			for _, cmp := range group {
				value = deleteFromSlice(value, cmp["Name"].(string))
				downstreamComponentsMap[key] = value
			}
		}

		plan = append(plan, map[string]interface{}{
			"GroupID":     fmt.Sprintf("%d", groupID),
			"Deployments": group,
		})
		groupID++
	}

	return plan
}

func (a *ModelOpsCMDConverterActivity) buildDeployments(
	deploymentGpID string,
	command string,
	endpointIP string,
	downstreamComponentsMap map[string][]string,
	system interface{},
	components []interface{}) ([]map[string]interface{}, error) {

	volumes := make([]map[string]interface{}, 0)
	if nil != system {
		volumeArray := system.(map[string]interface{})["Volume"]
		if nil != volumeArray {
			for _, volume := range volumeArray.([]interface{}) {
				volumes = append(volumes, map[string]interface{}{
					"Name": volume.(map[string]interface{})["Name"],
					"HostPath": map[string]interface{}{
						"Path": volume.(map[string]interface{})["Value"],
					},
				})
			}
		}
	}

	deployments := make([]map[string]interface{}, 0)
	for _, component := range components {
		componentObj := component.(map[string]interface{})
		componentName := componentObj["Name"].(string)
		componentType := componentObj["Type"].(string)
		componentRuntime := componentObj["Runtime"].(string)
		containerName := componentName
		dockerImage := componentObj["DockerImage"].(string)
		replicas, err := data.CoerceToLong(componentObj["Replicas"])
		if nil != err {
			log.Warn("Error setting replicas : ", err.Error(), ", will be set to 1 .")
			replicas = 1
		}

		env := make([]map[string]interface{}, 0)
		if "flogo" == strings.ToLower(componentRuntime) {
			env = append(env, map[string]interface{}{
				"Name":  "FLOGO_APP_PROPS_ENV",
				"Value": "auto",
			})
		}
		if nil != componentObj["Properties"] {
			for _, property := range componentObj["Properties"].([]interface{}) {
				propertyObj := property.(map[string]interface{})
				propertyName := propertyObj["Name"].(string)
				if strings.HasPrefix(propertyName, ".") {
					propertyName = fmt.Sprintf("%s%s", componentName, propertyName)
				}

				if "flogo" == strings.ToLower(componentRuntime) {
					propertyName = strings.Replace(propertyName, ".", "_", -1)
				}

				env = append(env, map[string]interface{}{
					"Name":  propertyName,
					"Value": propertyObj["Value"],
				})
			}
		}
		if nil != downstreamComponentsMap {
			downstreamHosts := make([]string, 0)
			for _, downstreamHost := range downstreamComponentsMap[componentName] {
				downstreamHosts = append(downstreamHosts, fmt.Sprintf("%s-ip-service", downstreamHost))
			}
			downstreamComponents, _ := json.Marshal(downstreamHosts)
			log.Info("(buildDeployments) componentName = ", componentName)
			log.Info("(buildDeployments) downstreamComponentsMap = ", downstreamComponentsMap)
			log.Info("(buildDeployments) downstreamComponents = ", string(downstreamComponents))
			if nil != downstreamComponents {
				env = append(env, map[string]interface{}{
					"Name":  "pipecoupler_downstreamHosts",
					"Value": string(downstreamComponents),
				})
			}

			env = append(env, map[string]interface{}{
				"Name":  "pipecoupler_port",
				"Value": "9997",
			})
		}

		volumeMounts := make([]map[string]interface{}, 0)
		if nil != componentObj["Volumes"] {
			for _, volume := range componentObj["Volumes"].([]interface{}) {
				volumeMounts = append(volumeMounts, map[string]interface{}{
					"MountPath": volume.(map[string]interface{})["MountPoint"],
					"Name":      volume.(map[string]interface{})["Name"],
				})
			}
		}

		isEndpoint := false
		containerPort := int64(9997)
		if "Source" == componentType || "EndPoint" == componentType {
			isEndpoint = true
			containerPort = a.endpointPort + util.GetSN()
			env = append(env, map[string]interface{}{
				"Name":  "System_Port",
				"Value": fmt.Sprintf("%d", containerPort),
			})
			env = append(env, map[string]interface{}{
				"Name":  "System_ExternalEndpointIP",
				"Value": fmt.Sprintf("%s", endpointIP),
			})
		}

		deployment := map[string]interface{}{
			"Command":    command,
			"Deployment": strings.ToLower(fmt.Sprintf("%s.%s.%s", deploymentGpID, componentType, componentName)),
			"IsEndpoint": isEndpoint,
			"Replicas":   replicas,
			"Containers": []map[string]interface{}{
				map[string]interface{}{
					"Name":  containerName,
					"Image": dockerImage,
					"Ports": map[string]interface{}{
						"ContainerPort": containerPort,
					},
					"VolumeMounts": volumeMounts,
					"Env":          env,
				},
			},
			"Volumes": volumes,
		}
		deployments = append(deployments, deployment)
	}

	log.Info("deployments ============>", deployments)

	return deployments, nil
}

func deleteFromSlice(slice []string, targetElement string) []string {
	found := false
	length := len(slice)
	for index, element := range slice {
		if element == targetElement {
			found = true
			if index < length-1 {
				slice[index] = slice[length-1]
			}
			break
		}
	}

	if found {
		slice[length-1] = ""
		slice = slice[:length-1]
	}

	return slice
}
