package f1

import (
	"fmt"
	"testing"
)

var air2f1 = &fnAir2F1Properties{}

func TestDockerFunc_Air2F1Properties(t *testing.T) {
	deployType := "docker"
	prefix := "services.MQTT_Postgres.environment"
	f1Properties := []interface{}{
		map[string]interface{}{
			"Group": "main",
			"Value": []interface{}{
				map[string]interface{}{
					"Name":  "services.MQTT_Postgres.environment[4]",
					"Value": "Mqtt_IoTMQTT_Broker_URL=tcp://a0056dbbadb2f11e99e4c067e42b309c-335014729.us-west-2.elb.amazonaws.com:1883",
				},
			},
		},
	}

	airProperties := []interface{}{
		map[string]interface{}{
			"Name":  "Mqtt_IoTMQTT_Broker_URL",
			"Value": "tcp://192.168.1.152:1883",
		},
		map[string]interface{}{
			"Name":  "Mqtt_IoTMQTT_Password",
			"Value": "SECRET:SzHvHYjKw3yPY0qYP766RSyvczQlGbTRXB0=",
		},
	}

	newf1, _ := air2f1.Eval(deployType, prefix, f1Properties, airProperties)
	fmt.Println("\n\nResult -> ", newf1)
}

func TestK8SFunc_Air2F1Properties(t *testing.T) {
	deployType := "k8s"
	prefix := "spec.template.spec.containers[0].env"
	f1Properties := []interface{}{
		map[string]interface{}{
			"Group": "main",
			"Value": []interface{}{
				map[string]interface{}{
					"Name":  "spec.template.spec.containers[0].env[4].name",
					"Value": "Mqtt_IoTMQTT_Broker_URL",
				},
				map[string]interface{}{
					"Name":  "spec.template.spec.containers[0].env[4].value",
					"Value": "tcp://a0056dbbadb2f11e99e4c067e42b309c-335014729.us-west-2.elb.amazonaws.com:1883",
				},
			},
		},
	}
	airProperties := []interface{}{
		map[string]interface{}{
			"Name":  "Mqtt_IoTMQTT_Broker_URL",
			"Value": "tcp://192.168.1.152:1883",
		},
		map[string]interface{}{
			"Name":  "Mqtt_IoTMQTT_Password",
			"Value": "SECRET:SzHvHYjKw3yPY0qYP766RSyvczQlGbTRXB0=",
		},
	}

	newf1, _ := air2f1.Eval(deployType, prefix, f1Properties, airProperties)
	fmt.Println("\n\nResult -> ", newf1)
}
