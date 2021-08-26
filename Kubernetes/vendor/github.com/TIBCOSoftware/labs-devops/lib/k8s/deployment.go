package k8s

import (
	"fmt"

	"github.com/TIBCOSoftware/labs-devops/lib/util"

	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/retry"
)

//-====================-//
//   Define Deployment
//-====================-//

func NewDeploymentByYAML(ymalObj map[interface{}]interface{}) (*Deployment, error) {
	deployment := &Deployment{
		_template: ymalObj,
	}
	return deployment, nil
}

func NewDeployment(name string, template []byte) (*Deployment, error) {
	var rootObject interface{}
	err := yaml.Unmarshal(template, &rootObject)
	if err != nil {
		return nil, err
	}

	deployment := &Deployment{
		_name:     name,
		_template: rootObject.(map[interface{}]interface{}),
	}
	return deployment, nil
}

type Deployment struct {
	_name     string
	_template map[interface{}]interface{}
}

func (this *Deployment) GetName() string {
	return this._name
}

func (this *Deployment) Create(cluster *Cluster) (map[string]interface{}, error) {
	log.Info("[Deployment:Create] Creat deployment : ", this._name)

	result, err := cluster.Namespace().Create(
		&unstructured.Unstructured{
			Object: this.createDeploymentObj(),
		},
		metav1.CreateOptions{},
	)
	if err != nil {
		return map[string]interface{}{this._name: result}, err
	}

	log.Info("[Deployment:Create] deployment created : ", result.GetName())
	return map[string]interface{}{this._name: result.Object}, nil
}

func (this *Deployment) UpdateSimple(cluster *Cluster) (map[string]interface{}, error) {
	name := this._template["metadata"].(map[interface{}]interface{})["name"].(string)
	log.Info("[Deployment:UpdateSimple] update deployment updated, name = ", name)

	var result *unstructured.Unstructured
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var err error
		/* Check if exist */
		_, err = cluster.Namespace().Get(name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get latest version of Deployment: %v", err)
		}
		/* Update with new value */
		result, err = cluster.Namespace().Update(
			&unstructured.Unstructured{
				Object: this.createDeploymentObj(),
			},
			metav1.UpdateOptions{},
		)

		return err
	})

	if retryErr != nil {
		log.Errorf("update failed: %v", retryErr)
		return nil, retryErr
	}

	log.Info("[Deployment:UpdateSimple] deployment updated : ", result.GetName())
	return map[string]interface{}{result.GetName(): result.Object}, nil
}

func (this *Deployment) Build(replicas int64, volumes []map[string]interface{}, container map[string]interface{}) (map[string]interface{}, error) {
	return this.Update(replicas, this.createDeploymentObj(), volumes, container)
}

func (this *Deployment) Update(
	replicas int64,
	deploymentObj map[string]interface{},
	volumes []map[string]interface{},
	container map[string]interface{}) (map[string]interface{}, error) {

	/* For containers */
	containers, found, err := unstructured.NestedSlice(deploymentObj, "spec", "template", "spec", "containers")
	if err != nil || !found || containers == nil {
		return nil, fmt.Errorf("deployment containers not found or error in spec: %v", err)
	}

	targetContainer := containers[0].(map[string]interface{})

	log.Info("container ============>", container["Ports"])

	/* ports */
	ports := map[string]interface{}{
		"containerPort": container["Ports"].(map[string]interface{})["ContainerPort"],
	}
	if err := unstructured.SetNestedSlice(targetContainer, []interface{}{ports}, "ports"); err != nil {
		return nil, err
	}

	/* volumeMounts */
	volumeMounts, found, err := unstructured.NestedSlice(targetContainer, "volumeMounts")
	if err != nil || !found || volumeMounts == nil {
		volumeMounts = make([]interface{}, 0)
	}
	for _, volumeMount := range container["VolumeMounts"].([]map[string]interface{}) {
		log.Info("volumeMount : ", volumeMount)
		if nil != volumeMount["MountPath"] {
			volumeMounts = append(volumeMounts, map[string]interface{}{
				"name":      volumeMount["Name"].(string),
				"mountPath": volumeMount["MountPath"],
			})
		}
	}
	if err := unstructured.SetNestedSlice(targetContainer, volumeMounts, "volumeMounts"); err != nil {
		return nil, err
	}

	/* env */
	env, found, err := unstructured.NestedSlice(targetContainer, "env")
	if err != nil || !found || env == nil {
		env = make([]interface{}, 0)
	}
	for _, envEntry := range container["Env"].([]map[string]interface{}) {
		log.Info("envEntry : ", envEntry)
		if nil != envEntry["Value"] {
			env = append(env, map[string]interface{}{
				"name":  envEntry["Name"].(string),
				"value": envEntry["Value"],
			})
		}
	}
	if err := unstructured.SetNestedSlice(targetContainer, env, "env"); err != nil {
		return nil, err
	}

	if err := unstructured.SetNestedField(targetContainer, container["Image"], "image"); err != nil {
		return nil, err
	}

	if err := unstructured.SetNestedField(targetContainer, container["Name"], "name"); err != nil {
		return nil, err
	}

	if err := unstructured.SetNestedField(deploymentObj, containers, "spec", "template", "spec", "containers"); err != nil {
		return nil, err
	}

	/* For volumes */
	volumesObj, found, err := unstructured.NestedSlice(deploymentObj, "spec", "template", "spec", "volumes")
	if err != nil || !found || containers == nil {
		volumesObj = make([]interface{}, 0)
	}

	for _, volume := range volumes {
		log.Info("volume : ", volume)
		volumesObj = append(volumesObj, map[string]interface{}{
			"name": volume["Name"],
			"hostPath": map[string]interface{}{
				"path": volume["HostPath"].(map[string]interface{})["Path"],
			},
		})
	}

	if err := unstructured.SetNestedField(deploymentObj, volumesObj, "spec", "template", "spec", "volumes"); err != nil {
		return nil, err
	}

	/* For replicas */
	log.Info("(UpdateDeployment) Replicas = ", replicas)
	if err := unstructured.SetNestedField(deploymentObj, replicas, "spec", "replicas"); err != nil {
		return nil, err
	}

	/* For selector */
	if err := unstructured.SetNestedField(deploymentObj, container["Name"], "spec", "selector", "matchLabels", "component"); err != nil {
		return nil, err
	}
	if err := unstructured.SetNestedField(deploymentObj, container["Name"], "spec", "template", "metadata", "labels", "component"); err != nil {
		return nil, err
	}

	/* For deployment name */
	if err := unstructured.SetNestedField(deploymentObj, this._name, "metadata", "name"); err != nil {
		return nil, err
	}

	log.Info("deploymentObj ============>", deploymentObj)

	return deploymentObj, nil
}

func (this *Deployment) createDeploymentObj() map[string]interface{} {
	newDeploymentObj := make(map[string]interface{})
	util.CopyMap(this._template, newDeploymentObj)
	return newDeploymentObj
}
