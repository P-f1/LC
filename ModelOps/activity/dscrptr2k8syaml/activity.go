/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package dscrptr2k8syaml

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	yaml "gopkg.in/yaml.v2"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/logger"
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
	oNamespace      = "Namespace"
	oDeployplan     = "Deployplan"
	oDeployments    = "Deployments"
	CMD_Deploy      = "deploy"
	CMD_Update      = "update"
	CMD_ListDeploys = "list"
	CMD_Undeploy    = "undeploy"
)

type Descriptor2K8sYamlActivity struct {
	metadata     *activity.Metadata
	mux          sync.Mutex
	endpointPort int64
}

func NewActivity(metadata *activity.Metadata) activity.Activity {
	aDescriptor2K8sYamlActivity := &Descriptor2K8sYamlActivity{
		metadata:     metadata,
		endpointPort: 10100,
	}

	return aDescriptor2K8sYamlActivity
}

func (a *Descriptor2K8sYamlActivity) Metadata() *activity.Metadata {
	return a.metadata
}

func (a *Descriptor2K8sYamlActivity) Eval(context activity.Context) (done bool, err error) {

	log.Info("[Descriptor2K8sYamlActivity:Eval] entering ........ ")

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
	namespace := deploymentGpID

	components := context.GetInput(iComponents).(*data.ComplexObject).Value
	system := context.GetInput(iSystem).(*data.ComplexObject).Value

	log.Info("command : ", command)
	log.Info("namespace : ", namespace)
	log.Info("components : ", components)

	context.SetOutput(iCommand, command)
	context.SetOutput(oNamespace, namespace)
	if nil != components {
		dataFlow := context.GetInput(iDataFlow).(*data.ComplexObject).Value
		log.Info("[Descriptor2K8sYamlActivity:Eval] dataFlow = ", dataFlow)
		if nil != dataFlow {
			deployplan := map[string]interface{}{
				"Strategy": "parallel",
				"Pipeline": deploymentPlan(downstreamComponents(dataFlow.([]interface{}))),
			}
			log.Info("[Descriptor2K8sYamlActivity:Eval] deployplan = ", deployplan)
			context.SetOutput(oDeployplan, &data.ComplexObject{Metadata: "Deployplan", Value: deployplan})
		}

		deployments, _ := a.buildComponentDscptrs(
			namespace,
			command,
			endpointIP,
			downstreamComponents(dataFlow.([]interface{})),
			system,
			components.([]interface{}))

		log.Info("[Descriptor2K8sYamlActivity:Eval] deployments = ", deployments)

		context.SetOutput(oDeployments, &data.ComplexObject{Metadata: "Deployments", Value: deployments})

		log.Info("[Descriptor2K8sYamlActivity:Eval] Exit ........ ")
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

func (a *Descriptor2K8sYamlActivity) buildComponentDscptrs(
	namespace string,
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
					"name": volume.(map[string]interface{})["Name"],
					"hostPath": map[string]interface{}{
						"path": volume.(map[string]interface{})["Value"],
					},
				})
			}
		}
	}

	componentDscptrs := make([]map[string]interface{}, 0)

	namespaceObj := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata": map[string]interface{}{
			"name": namespace,
		},
	}

	namespaceBytes, _ := yaml.Marshal(namespaceObj)
	componentDscptrs = append(componentDscptrs, map[string]interface{}{
		"Name":       namespace,
		"Descriptor": string(namespaceBytes),
	})

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
				"name":  "FLOGO_APP_PROPS_ENV",
				"value": "auto",
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
					"name":  propertyName,
					"value": propertyObj["Value"],
				})
			}
		}
		if nil != downstreamComponentsMap {
			downstreamHosts := make([]string, 0)
			for _, downstreamHost := range downstreamComponentsMap[componentName] {
				downstreamHosts = append(downstreamHosts, fmt.Sprintf("%s-ip-service", downstreamHost))
			}
			downstreamComponents, _ := json.Marshal(downstreamHosts)
			log.Info("(buildComponentDscptrs) componentName = ", componentName)
			log.Info("(buildComponentDscptrs) downstreamComponentsMap = ", downstreamComponentsMap)
			log.Info("(buildComponentDscptrs) downstreamComponents = ", string(downstreamComponents))
			if nil != downstreamComponents {
				env = append(env, map[string]interface{}{
					"name":  "pipecoupler_downstreamHosts",
					"value": string(downstreamComponents),
				})
			}

			env = append(env, map[string]interface{}{
				"name":  "pipecoupler_port",
				"value": "9997",
			})
		}

		volumeMounts := make([]map[string]interface{}, 0)
		if nil != componentObj["Volumes"] {
			for _, volume := range componentObj["Volumes"].([]interface{}) {
				volumeMounts = append(volumeMounts, map[string]interface{}{
					"mountPath": volume.(map[string]interface{})["MountPoint"],
					"name":      volume.(map[string]interface{})["Name"],
				})
			}
		}

		serviceType := "ClusterIP"
		containerPort := int64(9997)
		if "Source" == componentType || "EndPoint" == componentType {
			containerPort = a.endpointPort + util.GetSN()
			env = append(env, map[string]interface{}{
				"name":  "System_Port",
				"value": fmt.Sprintf("%d", containerPort),
			})
			env = append(env, map[string]interface{}{
				"name":  "System_ExternalEndpointIP",
				"value": fmt.Sprintf("%s", endpointIP),
			})
			serviceType = "LoadBalancer"
		}

		deploymentObj := map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": componentName,
			},
			"spec": map[string]interface{}{
				"replicas": replicas,
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"component": componentName,
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"component": componentName,
						},
					},
					"spec": map[string]interface{}{
						"containers": []map[string]interface{}{
							map[string]interface{}{
								"name":  containerName,
								"image": dockerImage,
								"ports": []map[string]interface{}{
									map[string]interface{}{
										"containerPort": containerPort,
									},
								},
								"volumeMounts": volumeMounts,
								"env":          env,
							},
						},
						"volumes": volumes,
					},
				},
			},
		}
		deploymentByte, _ := yaml.Marshal(deploymentObj)

		serviceObj := map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name": fmt.Sprintf("%s-ip-service", componentName),
			},
			"spec": map[string]interface{}{
				"type": serviceType,
				"selector": map[string]interface{}{
					"component": componentName,
				},
				"ports": []interface{}{
					map[string]interface{}{
						"port":       containerPort,
						"targetPort": containerPort,
					},
				},
			},
		}
		serviceByte, _ := yaml.Marshal(serviceObj)

		componentDscptrs = append(componentDscptrs, map[string]interface{}{
			"Name":       componentName,
			"Descriptor": string(deploymentByte),
		})
		componentDscptrs = append(componentDscptrs, map[string]interface{}{
			"Name":       fmt.Sprintf("%s-ip-service", componentName),
			"Descriptor": string(serviceByte),
		})
	}

	log.Info("ComponentDscptr ============>", componentDscptrs)

	return componentDscptrs, nil
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
