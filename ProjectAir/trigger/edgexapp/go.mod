module github.com/TIBCOSoftware/labs-modelops-contrib/air/trigger/edgexapp

go 1.13

replace github.com/TIBCOSoftware/labs-modelops-contrib/air/trigger/edgexapp/eventhandler => ./edgexapp/eventhandler

require (
	github.com/edgexfoundry/app-functions-sdk-go v1.2.0
	github.com/edgexfoundry/go-mod-core-contracts v0.1.58
	github.com/project-flogo/core v1.3.0
	github.com/stretchr/testify v1.5.1
)
