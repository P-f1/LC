package K8smonitor

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

type Settings struct {
}

type HandlerSettings struct {
	Connection connection.Manager `md:"Connection,required"`
}

type Output struct {
	JSONEvent string `md:"JSONEvent"`
}

func (this *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"JSONEvent": this.JSONEvent,
	}
}

func (this *Output) FromMap(values map[string]interface{}) error {

	var err error
	this.JSONEvent, err = coerce.ToString(values["JSONEvent"])
	if err != nil {
		return err
	}

	return nil
}
