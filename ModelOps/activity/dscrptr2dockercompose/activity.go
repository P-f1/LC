/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package dscrptr2dockercompose

import (
	"encoding/json"
	//	"errors"
	"fmt"
	"strings"
	"sync"

	//"github.com/google/uuid"
	yaml "gopkg.in/yaml.v2"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/P-f1/LC/labs-flogo-lib/util"
)

var log = logger.GetLogger("tibco-model-ops-cmdconverter")

var initialized bool = false

const (
	iNetwork        = "Network"
	iEndpointIP     = "EndpointIP"
	iEndpointPort   = "EndpointPort"
	iDeploymentGpID = "DeploymentGpID"
	iDataFlow       = "DataFlow"
	iComponents     = "Components"
	iSystem         = "System"
	oDeployGpID     = "DeployGpID"
	oDockerCompose  = "DockerCompose"
)

type ToDockerComposeActivity struct {
	metadata     *activity.Metadata
	mux          sync.Mutex
	endpointPort int64
}

func NewActivity(metadata *activity.Metadata) activity.Activity {
	aToDockerComposeActivity := &ToDockerComposeActivity{
		metadata:     metadata,
		endpointPort: 10100,
	}

	return aToDockerComposeActivity
}

func (a *ToDockerComposeActivity) Metadata() *activity.Metadata {
	return a.metadata
}

func (a *ToDockerComposeActivity) Eval(context activity.Context) (done bool, err error) {

	log.Info("[ToDockerComposeActivity:Eval] entering ........ ")

	deploymentGpID := context.GetInput(iDeploymentGpID)
	endpointIP := context.GetInput(iEndpointIP)
	endpointPort := context.GetInput(iEndpointIP)
	network := context.GetInput(iNetwork)

	components := context.GetInput(iComponents).(*data.ComplexObject).Value
	system := context.GetInput(iSystem).(*data.ComplexObject).Value

	log.Info("deploymentGpID : ", deploymentGpID)
	log.Info("components : ", components)

	context.SetOutput(oDeployGpID, deploymentGpID)

	if nil != components {
		dataFlow := context.GetInput(iDataFlow).(*data.ComplexObject).Value
		log.Info("[ToDockerComposeActivity:Eval] dataFlow = ", dataFlow)

		deployment := make(map[interface{}]interface{})
		deployment["version"] = "3.7"
		services, _ := a.buildDeployments(
			endpointIP,
			endpointPort,
			deploymentGpID,
			downstreamComponents(dataFlow.([]interface{}), deploymentGpID),
			system,
			components.([]interface{}))

		deployment["services"] = services

		if nil != network {
			deployment["networks"] = map[interface{}]interface{}{
				"default": map[interface{}]interface{}{
					//		"external": map[interface{}]interface{}{
					"name": network,
					//		},
				},
			}
		}

		dockerComposeBytes, err := yaml.Marshal(&deployment)
		if nil != err {
			log.Errorf("error: %v", err)
		}

		dockerCompose := string(dockerComposeBytes)
		log.Info("[ToDockerComposeActivity:Eval] dockerCompose = ", dockerCompose)

		context.SetOutput(oDockerCompose, dockerCompose)

		log.Info("[ToDockerComposeActivity:Eval] Exit ........ ")
	}

	return true, nil
}

func (a *ToDockerComposeActivity) buildDeployments(
	endpointIP interface{},
	endpointPort interface{},
	deploymentGpID interface{},
	downstreamComponentsMap map[string][]string,
	system interface{},
	components []interface{}) (map[string]map[string]interface{}, error) {

	volumeDef := make(map[interface{}]interface{}, 0)
	if nil != system {
		volumeArray := system.(map[string]interface{})["Volume"]
		if nil != volumeArray {
			for _, volume := range volumeArray.([]interface{}) {
				volumeDef[volume.(map[string]interface{})["Name"]] = volume.(map[string]interface{})["Value"]
			}
		}
	}

	deployments := make(map[string]map[string]interface{})
	for _, component := range components {
		componentObj := component.(map[string]interface{})
		componentName := componentObj["Name"].(string)
		componentType := componentObj["Type"].(string)
		componentRuntime := componentObj["Runtime"].(string)
		containerName := componentName
		if nil != deploymentGpID {
			containerName = longName(componentName, deploymentGpID)
		}
		dockerImage := componentObj["DockerImage"].(string)

		env := make([]interface{}, 0)
		if "flogo" == strings.ToLower(componentRuntime) {
			env = append(env, "FLOGO_APP_PROPS_ENV=auto")
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

				env = append(env, fmt.Sprintf("%s=%s", propertyName, propertyObj["Value"]))
			}
		}
		dependsOn := make([]interface{}, 0)
		if nil != downstreamComponentsMap {
			downstreamHosts := make([]string, 0)
			for _, downstreamHost := range downstreamComponentsMap[containerName] {
				downstreamHosts = append(downstreamHosts, downstreamHost)
				dependsOn = append(dependsOn, downstreamHost)
			}
			downstreamComponents, _ := json.Marshal(downstreamHosts)
			log.Info("(buildDeployments) componentName = ", componentName)
			log.Info("(buildDeployments) downstreamComponentsMap = ", downstreamComponentsMap)
			log.Info("(buildDeployments) downstreamComponents = ", string(downstreamComponents))
			if nil != downstreamComponents {
				env = append(env, fmt.Sprintf("%s=%s", "pipecoupler_downstreamHosts", string(downstreamComponents)))
				env = append(env, "pipecoupler_port=9997")
			}
		}

		volumes := make([]interface{}, 0)
		if nil != componentObj["Volumes"] {
			for _, volume := range componentObj["Volumes"].([]interface{}) {
				vmap := volume.(map[string]interface{})
				volumes = append(volumes, fmt.Sprintf("%s:%s", volumeDef[vmap["Name"]], vmap["MountPoint"]))
			}
		}

		deployment := map[string]interface{}{
			"image":          dockerImage,
			"container_name": containerName,
			"volumes":        volumes,
		}

		if 0 < len(dependsOn) {
			deployment["depends_on"] = dependsOn
		}

		if "Source" == componentType || "EndPoint" == componentType {
			containerPort := a.endpointPort + util.GetSN()
			deployment["ports"] = []interface{}{
				fmt.Sprintf("%d:%d", containerPort, containerPort),
			}

			if nil != endpointIP {
				env = append(env, fmt.Sprintf("System_ExternalEndpointIP=%s", endpointIP.(string)))
			}
			env = append(env, fmt.Sprintf("System_EndpointComponent=%s", containerName))
			env = append(env, fmt.Sprintf("System_Port=%d", containerPort))
		}
		deployment["environment"] = env
		deployments[containerName] = deployment
	}

	return deployments, nil
}

func downstreamComponents(
	dataFlow []interface{},
	deploymentGpID interface{}) map[string][]string {
	downstreamComponentsMap := make(map[string][]string)
	for _, flow := range dataFlow {
		me := flow.(map[string]interface{})["Upstream"].(string)
		downstream := flow.(map[string]interface{})["Downstream"].(string)
		if nil != deploymentGpID {
			me = longName(me, deploymentGpID)
			downstream = longName(downstream, deploymentGpID)
		}

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

func longName(componentName string, deploymentGpID interface{}) string {
	return fmt.Sprintf("%s-%s", componentName, deploymentGpID.(string))
}
