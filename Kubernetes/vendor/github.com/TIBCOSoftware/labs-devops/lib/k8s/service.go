package k8s

import (
	"encoding/json"

	"github.com/P-f1/LC/labs-devops/lib/util"
	
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//-====================-//
//   Define Service
//-====================-//

func BuildServiceByYMAL(ymal string) (*corev1.Service, error) {
	var objectRootObj interface{}
	err := yaml.Unmarshal([]byte(ymal), &objectRootObj)
	if err != nil {
		log.Errorf("err : %v\n", err)
	}

	srvMap := make(map[string]interface{})
	util.CopyMap(objectRootObj.(map[interface{}]interface{}), srvMap)

	srvBytes, err := json.Marshal(srvMap)
	if err != nil {
		log.Errorf("err : %v\n", err)
	}
	
	srv := &corev1.Service{}
	err = json.Unmarshal(srvBytes, srv)
	if err != nil {
		log.Errorf("err : %v\n", err)
	}

	log.Info("[Service:BuildService] service built : srv = ", srv)
	return srv, nil
}

func NewService(name string) *Service {
	service := &Service{
		_name: name,
	}
	return service
}

type Service struct {
	_name string
}

func (this *Service) GetName() string {
	return this._name
}

func (this *Service) BuildService(name string, serviceType string, compoment string, port int64, targetPort int64) (*corev1.Service, error) {
	log.Info("[Service:BuildService] Build Service : ", this._name)

	service := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata": map[string]interface{}{
			"name": name,
		},
		"spec": map[string]interface{}{
			"type": serviceType,
			"selector": map[string]interface{}{
				"component": compoment,
			},
			"ports": []interface{}{
				map[string]interface{}{
					"port":       port,
					"targetPort": targetPort,
				},
			},
		},
	}

	srvBytes, err := json.Marshal(service)
	if err != nil {
		log.Errorf("err : %v\n", err)
	}
	
	srv := &corev1.Service{}
	err = json.Unmarshal(srvBytes, srv)
	if err != nil {
		log.Errorf("err : %v\n", err)
	}

	log.Info("[Service:BuildService] service built : ", srv.GetName())
	return srv, nil
}
