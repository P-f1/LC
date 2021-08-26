package edgexapp

import (
	"github.com/project-flogo/core/data/coerce"
)

type Settings struct {
	Port string `md:"Port"`
}

type HandlerSettings struct {
	Path string `md:"Path"`
}

type Output struct {
	Event map[string]interface{} `md:"Event"`
}

func (this *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Event": this.Event,
	}
}

func (this *Output) FromMap(values map[string]interface{}) error {

	var err error
	this.Event, err = coerce.ToObject(values["Event"])
	if err != nil {
		return err
	}

	return nil
}
