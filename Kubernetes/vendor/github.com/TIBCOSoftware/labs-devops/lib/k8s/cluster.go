package k8s

import (
	"encoding/json"
	"errors"
	"fmt"

	yaml2 "gopkg.in/yaml.v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	tcoredv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/util/retry"

	wproxy "k8s.io/client-go/tools/watch"
)

//-==========================-//
//   Define Virtual Cluster
//-==========================-//

type ClusterEventListerner interface {
	HandleEvent(event string) error
	NoMoreEvent(resume bool)
}

func NewCluster(kubeconfigPath string, namespace string) (*Cluster, error) {
	return &Cluster{
		_namespace:     namespace,
		_deploymentRes: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
	}, nil
}

type Cluster struct {
	_watcher       watch.Interface
	_namespace     string
	_deploymentRes schema.GroupVersionResource
}

func (this *Cluster) Name() string {
	return this._namespace
}

func (this *Cluster) Namespace() dynamic.ResourceInterface {
	return client.Resource(this._deploymentRes).Namespace(this._namespace)
}

func (this *Cluster) ServiceBuilder() tcoredv1.ServiceInterface {
	return clientset.CoreV1().Services(this._namespace)
}

/* For create by YMAL */

func (this *Cluster) CreateByYAML(kind string, yaml string) (map[string]interface{}, error) {
	
	switch kind {
		case "Deployment", "deployment" : {
			var rootObject interface{}
			err := yaml2.Unmarshal([]byte(yaml), &rootObject)
			if err != nil {
				return nil, err
			}

			k8sObject := rootObject.(map[interface{}]interface{})
			deployment, err := NewDeploymentByYAML(k8sObject)
			if err != nil {
				return nil, err
			}
			result, err := deployment.Create(this)
			return result, err
		}
		case "Service", "service" : {
			service, err := BuildServiceByYMAL(yaml)
			if err != nil {
				return nil, err
			}
			
			srvObj, err := this.ServiceBuilder().Create(service)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{"service": srvObj}, nil
		}
	}

	return make(map[string]interface{}), nil
}

func (this *Cluster) UpdateByYAML(kind string, yaml string) (map[string]interface{}, error) {
	
	switch kind {
		case "Deployment", "deployment" : {
			var rootObject interface{}
			err := yaml2.Unmarshal([]byte(yaml), &rootObject)
			if err != nil {
				return nil, err
			}

			k8sObject := rootObject.(map[interface{}]interface{})
			deployment, err := NewDeploymentByYAML(k8sObject)
			if err != nil {
				return nil, err
			}
			result, err := deployment.UpdateSimple(this)
			return result, err
		}
		case "Service", "service" : {
			service, err := BuildServiceByYMAL(yaml)
			if err != nil {
				return nil, err
			}
			
			srvObj, err := this.ServiceBuilder().Update(service)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{"service": srvObj}, nil
		}
	}

	return make(map[string]interface{}), nil
}

func (this *Cluster) Delete(kind string, name string) (map[string]interface{}, error) {
	
	switch kind {
		case "Deployment", "deployment" : {
			if "" != name {
				return this.UndeployDeployment(name)
			} else {
				return this.DeleteDeployments()
			}
		}
		case "Service", "service" : {
			if "" != name {
				return this.DeleteService(name)
			} else {
				return this.DeleteServices()
			}
		}
	}

	return make(map[string]interface{}), errors.New("Illegal kind!")
}

func (this *Cluster) List(kind string) (map[string]interface{}, error) {
	switch kind {
		case "Deployment", "deployment" : {
			return this.ListDeployments()
		}
		case "Service", "service" : {
			return this.ListServices()
		}
	}

	return make(map[string]interface{}), errors.New("Illegal kind!")	
}

/* For Kubernetes service */

func (this *Cluster) ListServices() (map[string]interface{}, error) {
	services, _ := this.ServiceBuilder().List(metav1.ListOptions{})
	servicesMap := make(map[string]interface{})
	for _, service := range services.Items {
		servicesMap[service.GetName()] = service
	}

	return servicesMap, nil
}

func (this *Cluster) DeleteServices() (map[string]interface{}, error) {
	servicesMap, _ := this.ListServices()
	for name, _ := range servicesMap {
		this.ServiceBuilder().Delete(name, &metav1.DeleteOptions{})
	}
	return servicesMap, nil
}

func (this *Cluster) CreateService(name string, serviceType string, component string, port int64, targetPort int64) (map[string]interface{}, error) {
	log.Info("[Cluster:Create] Creat service : ", name)

	service, err := NewService(name).BuildService(name, serviceType, component, port, targetPort)
	if err != nil {
		return nil, err
	}

	srvObj, err := this.ServiceBuilder().Create(service)
	log.Infof("srv : %v , err : %v\n", srvObj, err)

	if err != nil {
		return map[string]interface{}{name: srvObj}, err
	}

	log.Info("[Cluster:Create] service created : ", srvObj.GetName())
	return map[string]interface{}{name: srvObj}, nil
}

func (this *Cluster) DeleteService(name string) (map[string]interface{}, error) {
	log.Info("[Cluster:Delete] Delete service : ", name)

	err := this.ServiceBuilder().Delete(name, &metav1.DeleteOptions{})
	log.Infof("srv : %s , err : %v\n", name, err)

	if err != nil {
		return map[string]interface{}{"Service": name}, err
	}

	log.Info("[Cluster:Delete] service deleted : ", name)
	return map[string]interface{}{"Service": name}, nil
}

/* For Kubernetes deployment */

func (this *Cluster) ListDeployments() (map[string]interface{}, error) {
	log.Info("[KubernetesDeployActivity:listDeploys] namespace : ", this.Name())

	list, err := this.Namespace().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	deployments := make(map[string]interface{})
	for _, d := range list.Items {
		deploymentName := d.Object["metadata"].(map[string]interface{})["name"].(string)
		deployments[deploymentName] = d.Object
		replicas, found, err := unstructured.NestedInt64(d.Object, "spec", "replicas")
		if err != nil || !found {
			fmt.Printf("Replicas not found for deployment %s: error=%s", d.GetName(), err)
			continue
		}
		fmt.Printf(" * %s (%d replicas)\n", d.GetName(), replicas)
	}

	log.Info("[KubernetesDeployActivity:listDeploys] done ...")
	return deployments, nil
}

func (this *Cluster) DeleteDeployments() (map[string]interface{}, error) {
	log.Info("[KubernetesDeployActivity:undeployAll] namespace : ", this.Name())
	deployments, err := this.ListDeployments()
	if nil != err {
		return nil, err
	}
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	for deployment, _ := range deployments {
		if err := this.Namespace().Delete(deployment, deleteOptions); err != nil {
			delete(deployments, deployment)
			continue
		}
	}
	log.Info("[KubernetesDeployActivity:undeployAll] namespace : ", this.Name())
	return deployments, nil
}

func (this *Cluster) CreateDeployment(deploymentName string, template []byte, replicas int64, volumes []map[string]interface{}, container map[string]interface{}) (map[string]interface{}, error) {
	log.Info("[Cluster:Deploy] Creat deployment : ", deploymentName)
	deployment, _ := NewDeployment(deploymentName, template)
	deploymentObj, err := deployment.Build(replicas, volumes, container)
	if nil != err {
		return nil, err
	}
	log.Info("[KubernetesDeployActivity:Deploy] DeploymentObj created : ", deploymentObj)

	result, err := this.Namespace().Create(
		&unstructured.Unstructured{
			Object: deploymentObj,
		},
		metav1.CreateOptions{},
	)
	if err != nil {
		return map[string]interface{}{deployment.GetName(): result}, err
	}

	log.Info("[Cluster:Deploy] deployment created : ", result.GetName())
	return map[string]interface{}{deployment.GetName(): result.Object}, nil
}

func (this *Cluster) UpdateDeployment(deploymentName string, replicas int64, volumes []map[string]interface{}, container map[string]interface{}) (map[string]interface{}, error) {
	log.Info("[Cluster:update] update deployment : ", deploymentName)

	var result *unstructured.Unstructured
	deployment, _ := NewDeployment(deploymentName, nil)
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var err error
		liveInstance, err := this.Namespace().Get(deploymentName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get latest version of Deployment: %v", err)
		}

		liveInstance.Object, err = deployment.Update(replicas, liveInstance.Object, volumes, container)
		if nil != err {
			return err
		}

		result, err = this.Namespace().Update(liveInstance, metav1.UpdateOptions{})

		return err
	})

	if retryErr != nil {
		log.Errorf("update failed: %v", retryErr)
		return nil, retryErr
	}

	log.Info("[Cluster:update] deployment updated : ", result.GetName())
	return map[string]interface{}{deployment.GetName(): result.Object}, nil
}

func (this *Cluster) Ping(deployment string) (map[string]interface{}, error) {
	log.Info("[Cluster:ping] deployment : ", deployment)

	result, getErr := this.Namespace().Get(deployment, metav1.GetOptions{})
	if getErr != nil {
		return nil, fmt.Errorf("failed to find deployment: %v", getErr)
	}

	deployments := map[string]interface{}{
		deployment: result.Object,
	}

	log.Info("[Cluster:ping] deployment : ", deployment)
	return deployments, nil
}

func (this *Cluster) UndeployDeployment(deployment string) (map[string]interface{}, error) {
	log.Info("[Cluster:undeploy] deployment : ", deployment)

	result, getErr := this.Namespace().Get(deployment, metav1.GetOptions{})
	if getErr != nil {
		return nil, fmt.Errorf("failed to find deployment: %v", getErr)
	}

	deployments := map[string]interface{}{
		deployment: result.Object,
	}

	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	if err := this.Namespace().Delete(deployment, deleteOptions); err != nil {
		return nil, err
	}

	log.Info("[KubernetesDeployActivity:undeploy] deployment : ", deployment)
	return deployments, nil
}

func (this *Cluster) StopEvent() {
	this._watcher.Stop()
}

func (this *Cluster) Watch(options metav1.ListOptions) (watch.Interface, error) {
	watcher, err := client.Resource(this._deploymentRes).Namespace(this._namespace).Watch(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return watcher, nil
}

func (this *Cluster) ListenEvent(listener ClusterEventListerner) error {
	log.Info("[Cluster:ListenEvent] namespace : ", this._namespace)

	var err error
	this._watcher, err = wproxy.NewRetryWatcher("0", this)
	if nil != err {
		return err
	}

	go func() {
		for true {
			event, ok := <-this._watcher.ResultChan()
			if !ok || "" == event.Type || nil == event.Object {
				log.Warn("[Cluster:ListenEvent] event.Type = ", event.Type, ", event.Object = ", event.Object)
				log.Warn("\n\n[Cluster:ListenEvent] The watched channel is broken!!\n\n")
				listener.NoMoreEvent(true)
				return
			}
			eventBytes, _ := json.Marshal(event)
			listener.HandleEvent(string(eventBytes))
		}
	}()

	log.Info("[Cluster:ListenEvent] done ...")
	return nil
}
