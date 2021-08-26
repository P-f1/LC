package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/project-flogo/ml/activity/inference/framework"
	"github.com/project-flogo/ml/activity/inference/framework/tf"
	"github.com/project-flogo/ml/activity/inference/model"
)

var _ tf.TensorflowModel

// variables needed to persist model between inferences
var tfmodelmap map[string]*model.Model
var modelRunMutex sync.Mutex

var TF TFEngine

type TFEngine struct {
}

func (this TFEngine) Execute(
	fw string,
	modelpath string,
	sigDef string,
	tag string,
	features []interface{}) interface{} {

	tfFramework := framework.Get(fw)
	fmt.Printf("Using Framework %s", tfFramework.FrameworkTyp())
	if tfFramework == nil {
		fmt.Printf("%s framework not registered", fw)
	}
	fmt.Printf("Loaded Framework: " + tfFramework.FrameworkTyp())

	// Defining the flags to be used to load model
	flags := model.ModelFlags{
		Tag:    tag,
		SigDef: sigDef,
	}

	// if modelmap does not exist then make it
	if tfmodelmap == nil {
		tfmodelmap = make(map[string]*model.Model)
		fmt.Printf("Making map of models with keys of 'ModelKey'.")
	}

	// check if this instance of tf model already exists if not load it
	modelKey := modelpath
	fmt.Printf("ModelKey is:", modelKey)
	var err error
	if tfmodelmap[modelKey] == nil {
		tfmodelmap[modelKey], err = model.Load(modelpath, tfFramework, flags)
		if err != nil {
			fmt.Printf("Model loading fail")
		}

	} else {
		fmt.Printf("Model already loaded - skipping loading")
	}

	fmt.Printf("Incoming features: ")
	fmt.Printf("%v", features)

	// Grab the input feature set and parse out the feature labels and values
	inputSample := make(map[string]interface{})
	for _, feat := range features {
		featMap := feat.(map[string]interface{})
		inputName := featMap["name"].(string)
		inmodel := false
		for key := range tfmodelmap[modelKey].Metadata.Inputs.Features {
			if key == inputName {
				inputSample[inputName] = featMap["data"]
				inmodel = true
			}
		}
		if !inmodel {
			fmt.Printf("%s not an input into model", featMap["name"].(string))
		}
	}

	fmt.Printf("Parsing of features completed")

	modelRunMutex.Lock()
	tfmodelmap[modelKey].SetInputs(inputSample)
	output, err := tfmodelmap[modelKey].Run(tfFramework)
	modelRunMutex.Unlock()
	if err != nil {
		fmt.Printf("Error running the ml framework: %s", err)
	}

	fmt.Printf("Model execution completed with result:")
	fmt.Printf("output : %v", output)

	if strings.Contains(tfmodelmap[modelKey].Metadata.Method, "tensorflow/serving/classify") {
		var out = make(map[string]interface{})

		classes := output["classes"].([][]string)[0]
		scores := output["scores"].([][]float32)[0]

		for i := 0; i < len(classes); i++ {
			out[classes[i]] = scores[i]
		}

		fmt.Printf("1 : %v", out)
		return out
	} else {
		fmt.Printf("2 : %v", output)
		return output
	}
}
