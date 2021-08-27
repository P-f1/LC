package cmltfengine

import (
	"errors"
	"plugin"
	"sync"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/logger"
)

// activityLog is the default logger for the Log Activity
var log = logger.GetLogger("flogo-cml-tf")

const (
	sModel     = "model"
	sFramework = "framework"
	sSigDef    = "sigDefName"
	sTag       = "tag"
	sPlugin    = "plugin"

	iInput    = "Input"
	iFeatures = "features"
	oOutput   = "Output"
)

func NewActivity(metadata *activity.Metadata) activity.Activity {
	return &CMLTFEngineActivity{metadata: metadata}
}

// CMLTFEngineActivity is an Activity that is used to log a message to the console
// inputs : {message, flowInfo}
// outputs: none
type CMLTFEngineActivity struct {
	metadata *activity.Metadata
	tfengine TFEngine
	mux      sync.Mutex
}

// Metadata returns the activity's metadata
func (a *CMLTFEngineActivity) Metadata() *activity.Metadata {
	return a.metadata
}

// Eval implements api.Activity.Eval - Logs the Message
func (a *CMLTFEngineActivity) Eval(context activity.Context) (done bool, err error) {

	tfengine, err := a.getTFEngine(context)
	if nil != err {
		return true, err
	}

	input := context.GetInput(iInput).(*data.ComplexObject).Value.(map[string]interface{})
	features := input["DataIn"].([]interface{})

	/*
		var estInputsB = make(map[string]interface{})
		estInputsB["one"] = 0.140586
		estInputsB["two"] = 0.140586
		estInputsB["three"] = 0.140586
		estInputsB["label"] = 0.

		var featuresB []interface{}
		featuresB = append(featuresB, map[string]interface{}{
			"name": "inputs",
			"data": estInputsB,
		})

		[{"name" : "inputs", "data" : [{"one", 0.140586}, {"two, 0.140586}, {"three", 0.140586},{"label", 0.}]}]

	*/

	framework, ok := context.GetSetting(sFramework)
	if !ok {
		log.Warn("(CMLPipelineActivity.getPipeline) framework not set !")
	}
	model, ok := context.GetSetting(sModel)
	if !ok {
		log.Warn("(CMLPipelineActivity.getPipeline) model not set !")
	}
	sigDef, ok := context.GetSetting(sSigDef)
	if !ok {
		log.Warn("(CMLPipelineActivity.getPipeline) sigDef not set !")
	}
	tag, ok := context.GetSetting(sTag)
	if !ok {
		log.Warn("(CMLPipelineActivity.getPipeline) tag not set !")
	}

	result := tfengine.Execute(
		framework.(string),
		model.(string),
		sigDef.(string),
		tag.(string),
		features,
	)

	context.SetOutput(oOutput, result)

	return true, nil
}

func (a *CMLTFEngineActivity) getTFEngine(context activity.Context) (TFEngine, error) {
	tfengine := a.tfengine

	if nil == tfengine {
		a.mux.Lock()
		defer a.mux.Unlock()
		tfengine = tfengine
		if nil == tfengine {

			enginePlugin, ok := context.GetSetting(sPlugin)
			if !ok {
				log.Error("(CMLPipelineActivity.getPipeline) engine plugin not set !")
				return nil, errors.New("engine plugin not set !")
			}

			// load module
			// 1. open the so file to load the symbols
			plug, err := plugin.Open(enginePlugin.(string))
			if err != nil {
				log.Errorf("(CMLPipelineActivity.getPipeline) engine plugin %s can not be opened !", enginePlugin.(string))
				return nil, err
			}

			// 2. look up a symbol (an exported function or variable)
			// in this case, variable Greeter
			symEng, err := plug.Lookup("TF")
			if err != nil {
				log.Error("(CMLPipelineActivity.getPipeline) TF lookup fail !")
				return nil, err
			}

			// 3. Assert that loaded symbol is of a desired type
			// in this case interface type TF (defined above)
			tfengine, ok = symEng.(TFEngine)
			if !ok {
				log.Error("(CMLPipelineActivity.getPipeline) unexpected type from module symbol !")
				return nil, errors.New("unexpected type from module symbol")
			}
		}
	}
	return tfengine, nil
}

type TFEngine interface {
	Execute(
		framework string,
		model string,
		sigDef string,
		tag string,
		features []interface{}) interface{}
}
