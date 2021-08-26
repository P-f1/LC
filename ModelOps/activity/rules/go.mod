module github.com/TIBCOSoftware/ModelOps/activity/rules

go 1.12

require (
	github.com/Shopify/sarama v1.28.0
	github.com/TIBCOSoftware/flogo-lib v0.5.8
	github.com/TIBCOSoftware/labs-flogo-lib v0.0.0-00010101000000-000000000000
	github.com/eclipse/paho.mqtt.golang v1.3.2
	github.com/edgexfoundry/app-functions-sdk-go v1.3.1
	github.com/edgexfoundry/go-mod-core-contracts v0.1.149
	github.com/project-flogo/rules v0.0.0-00010101000000-000000000000
)

replace github.com/TIBCOSoftware/labs-flogo-lib => ../../../../labs-flogo-lib

replace github.com/project-flogo/rules => /Users/steven/Applications/flogo/2.9.0_f1/2.9/lib/ext/src/github.com/project-flogo/rules
