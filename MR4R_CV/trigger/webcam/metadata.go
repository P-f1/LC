package webcam

import (
	"errors"
)

type Settings struct {
}

type HandlerSettings struct {
	DeviceID string `md:"DeviceID,required"`
}

type Output struct {
	VideoFrame        string `md:"VideoFrame"`
	OutputFrameWidth  int    `md:"OutputFrameWidth"`
	OutputFrameHeight int    `md:"OutputFrameHeight"`
}

func (this *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"VideoFrame":        this.VideoFrame,
		"OutputFrameWidth":  this.OutputFrameWidth,
		"OutputFrameHeight": this.OutputFrameHeight,
	}
}

func (this *Output) FromMap(values map[string]interface{}) error {

	var ok bool
	this.VideoFrame, ok = values["VideoFrame"].(string)
	if !ok {
		return errors.New("Invalid VideoFrame data!")
	}
	this.OutputFrameWidth, ok = values["OutputFrameWidth"].(int)
	if !ok {
		return errors.New("Invalid OutputFrameWidth data!")
	}
	this.OutputFrameHeight, ok = values["OutputFrameHeight"].(int)
	if !ok {
		return errors.New("Invalid OutputFrameHeight data!")
	}

	return nil
}
