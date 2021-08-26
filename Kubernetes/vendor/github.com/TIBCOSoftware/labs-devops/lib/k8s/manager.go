package k8s

/*
Command : ""
DeployGPID : ""
Deployplan : {
	Strategy : "",
	Pipeline : [
		{
			GroupID : "",
			Deployments : [
				{
					Name : ""
					Type : ""
				}
			]
		}
	]
},
Deployments : [
	{
		Command : "",
		Deployment : "",
		Replicas : 1,
		Containers : [
			{
				Name : "",
				Image : "",
				Ports : {
					ContainerPort : ""
				},
				Env : [
					{
						Name : "",
						Value : ""
					}
				]
			}
		]
	}
]
*/

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/TIBCOSoftware/flogo-lib/logger"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var client dynamic.Interface = nil
var clientset *kubernetes.Clientset

var log = logger.GetLogger("tibco-k8s")

//-==========================-//
//   Define DeployManager
//-==========================-//

func NewDeployManager(kubeconfigPath string, namespace string, template []byte) (*DeployManager, error) {

	if nil == client {
		kubeconfig := flag.String(
			"kubeconfig",
			filepath.Join(kubeconfigPath, "config"),
			"(optional) absolute path to the kubeconfig file",
		)

		flag.Parse()

		config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			return nil, err
		}

		client, err = dynamic.NewForConfig(config)
		if err != nil {
			return nil, err
		}
		logger.Info("Client created, Client : ", client)

		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			return nil, err
		}
		logger.Info("Clientset created, Clientset : ", clientset)
	}

	cluster, err := NewCluster(kubeconfigPath, namespace)
	manager := &DeployManager{
		template:       template,
		defaultCluster: cluster,
		clusters:       make(map[string]*Cluster),
	}
	return manager, err
}

type DeployManager struct {
	template       []byte
	defaultCluster *Cluster
	clusters       map[string]*Cluster
}

func (this *DeployManager) GetCluster(namespace interface{}) (*Cluster, err) {
	if nil == namespace || namespace == this.defaultCluster.Name() {
		return this.defaultCluster, nil
	} else if nil != this.clusters[namespace.(string)] {
		return this.clusters[namespace.(string)], nil
	}

	cluster, err := NewCluster(kubeconfigPath, namespace.(string))
	if nil != err {
		return nil, err
	}
	clusters[namespace.(string)] = cluster
	return cluster, nil
}

/* Simple API */
func (this *DeployManager) Create(kind string, yaml string) (map[string]interface{}, error) {
	log.Info("[DeployManager:Create] enter, kind : ", kind)
	defer log.Info("[DeployManager:Create] done kind : ", kind)

	result, err := this.defaultCluster.CreateByYAML(kind, yaml)

	return result, err
}

func (this *DeployManager) Update(kind string, yaml string) (map[string]interface{}, error) {
	log.Info("[DeployManager:Update] enter, kind : ", kind)
	defer log.Info("[DeployManager:Update] done kind : ", kind)

	result, err := this.defaultCluster.UpdateByYAML(kind, yaml)

	return result, err
}

func (this *DeployManager) Delete(kind string, name string) (map[string]interface{}, error) {
	log.Info("[DeployManager:Delete] enter, kind : ", kind, ", name : ", name)
	defer log.Info("[DeployManager:Delete] done kind : ", kind, ", name : ", name)

	result, err := this.defaultCluster.Delete(kind, name)

	return result, err
}

func (this *DeployManager) List(kind string) (map[string]interface{}, error) {
	log.Info("[DeployManager:list] kind : ", kind)
	defer log.Info("[DeployManager:list] done ...")

	result, _ := this.defaultCluster.List(kind)

	return result, nil
}

func (this *DeployManager) ListDeploys() (map[string]interface{}, error) {
	return ListDeploys(nil), nil
}

func (this *DeployManager) ListDeploys(namespace interface{}) (map[string]interface{}, error) {
	log.Info("[DeployManager:listDeploys] namespace : ", this.cluster.Name())

	deployments, _ := this.GetCluster(namespace).ListDeployments()

	log.Info("[DeployManager:listDeploys] done ...")
	return deployments, nil
}

func (this *DeployManager) UndeployAll(namespace interface{}) (map[string]interface{}, error) {
	return UndeployAll(), nil
}

func (this *DeployManager) UndeployAll(namespace interface{}) (map[string]interface{}, error) {
	log.Info("[DeployManager:undeployAll] namespace : ", this.cluster.Name())

	deployments, _ := this.GetCluster(namespace).DeleteDeployments()

	this.GetCluster(namespace).DeleteServices()

	log.Info("[DeployManager:undeployAll] namespace : ", this.cluster.Name())
	return deployments, nil
}

/* K8s pipeline API */

func (this *DeployManager) GroupDeploy(namespace interface{}, gpID string, plan map[string]interface{}, deployments []map[string]interface{}) (map[string]interface{}, error) {
	log.Info("[DeployManager:GroupDeploy] Creat : id = ", gpID, ", plan = ", plan, ", deployments = ", deployments)
	status := make(map[string]interface{})
	deploymentMap := make(map[string]map[string]interface{})
	for _, value := range deployments {
		componentName := value["Deployment"].(string)
		componentName = componentName[strings.LastIndex(componentName, ".")+1:]
		deploymentMap[componentName] = value
	}

	log.Info("[DeployManager:GroupDeploy] Creat : deploymentMap = ", deploymentMap)
	for _, group := range plan["Pipeline"].([]map[string]interface{}) {
		log.Info("[DeployManager:GroupDeploy] group = ", group)
		log.Info("[DeployManager:GroupDeploy] groupID = ", group["GroupID"])
		for _, deployment := range group["Deployments"].([]map[string]interface{}) {
			log.Info("[DeployManager:GroupDeploy] deployment = ", deployment)
			deploymentName := deployment["Name"].(string)
			deploy := deploymentMap[deploymentName]
			log.Info("[DeployManager:GroupDeploy] deploy = ", deploy)
			isEndpoint := deploy["IsEndpoint"].(bool)
			replicas := deploy["Replicas"].(int64)
			container := deploy["Containers"].([]map[string]interface{})[0]
			volumes := deploy["Volumes"].([]map[string]interface{})

			result, err := this.Deploy(namespace, deploymentName, isEndpoint, replicas, volumes, container)
			status[deploymentName] = result
			if nil != err {
				log.Errorf("[DeployManager:GroupDeploy] error = %v, status = %v", err, status)
				return status, err
			}
		}
	}
	log.Info("[DeployManager:GroupDeploy] exit ......")
	return status, nil
}

func (this *DeployManager) Deploy(
	namespace interface{},
	deployment string,
	isEndpoint bool,
	replicas int64,
	volumes []map[string]interface{},
	container map[string]interface{}) (map[string]interface{}, error) {
	log.Info("[DeployManager:Deploy] Creat deployment : ", deployment)

	port := container["Ports"].(map[string]interface{})["ContainerPort"].(int64)

	serviceType := "ClusterIP"
	if isEndpoint {
		serviceType = "LoadBalancer"
	}
	result01, err := this.GetCluster(namespace).CreateService(fmt.Sprintf("%s-ip-service", deployment), serviceType, deployment, port, port)
	log.Info("[DeployManager:Deploy] service created : ", result01)
	result02, err := this.GetCluster(namespace).CreateDeployment(deployment, this.template, replicas, volumes, container)
	log.Info("[DeployManager:Deploy] deployment created : ", result02)

	return result02, err
}

func (this *DeployManager) UpdateSet(namespace, deployment string, replicas int64, volumes []map[string]interface{}, container map[string]interface{}) (map[string]interface{}, error) {
	log.Info("[DeployManager:update] update deployment : ", deployment)

	result, err := this.GetCluster(namespace).UpdateDeployment(deployment, replicas, volumes, container)

	log.Info("[DeployManager:update] deployment updated : ", result)
	return result, err
}

func (this *DeployManager) Ping(namespace interface{}, deployment string) (map[string]interface{}, error) {
	log.Info("[DeployManager:ping] deployment : ", deployment)

	result, getErr := this.GetCluster(namespace).Namespace().Get(deployment, metav1.GetOptions{})
	if getErr != nil {
		return nil, fmt.Errorf("failed to find deployment: %v", getErr)
	}

	deployments := map[string]interface{}{
		deployment: result.Object,
	}

	log.Info("[DeployManager:ping] deployment : ", deployment)
	return deployments, nil
}

func (this *DeployManager) Undeploy(namespace interface{}, deployment string) (map[string]interface{}, error) {
	log.Info("[DeployManager:undeploy] deployment : ", deployment)

	result01, err := this.GetCluster(namespace).DeleteService(fmt.Sprintf("%s-ip-service", deployment))
	log.Info("[DeployManager:Undeploy] service deleted : ", result01)
	result02, err := this.GetCluster(namespace).UndeployDeployment(deployment)
	log.Info("[DeployManager:Undeploy] deployment deleted : ", result01)

	log.Info("[DeployManager:undeploy] deployment : ", deployment)
	return result02, err
}
