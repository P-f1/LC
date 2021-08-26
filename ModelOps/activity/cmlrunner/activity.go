/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package cmlrunner

import (
	ctx "context"
	"fmt"
	"sync"

	_ "github.com/project-flogo/catalystml-flogo/action"
	_ "github.com/project-flogo/catalystml-flogo/operations/cleaning"
	_ "github.com/project-flogo/catalystml-flogo/operations/image_processing"
	_ "github.com/project-flogo/catalystml-flogo/operations/nlp"
	_ "github.com/project-flogo/catalystml-flogo/operations/string_processing"

	"github.com/project-flogo/core/action"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/TIBCOSoftware/labs-flogo-lib/util"
)

var log = logger.GetLogger("tibco-cml-pipeline")

var initialized bool = false

const (
	CML         = "CML"
	ISURL       = "IsRef"
	InputSchema = "InputSchema"
	Input       = "Input"
	Output      = "Output"
)

type CMLRunnerActivity struct {
	metadata         *activity.Metadata
	activityToModel  map[string]string
	pipelineRegistry map[string]*Pipeline
	mux              sync.Mutex
}

func NewActivity(metadata *activity.Metadata) activity.Activity {
	aCMLRunnerActivity := &CMLRunnerActivity{
		metadata:         metadata,
		activityToModel:  make(map[string]string),
		pipelineRegistry: make(map[string]*Pipeline),
	}

	return aCMLRunnerActivity
}

func (a *CMLRunnerActivity) Metadata() *activity.Metadata {
	return a.metadata
}

func (a *CMLRunnerActivity) Eval(context activity.Context) (done bool, err error) {

	log.Info("[CMLRunnerActivity:Eval] entering ........ ")

	pipeline, err := a.getPipeline(context)

	if nil != err {
		return false, err
	}

	input := context.GetInput(Input).(*data.ComplexObject).Value
	log.Info("[CMLRunnerActivity:Eval] Input : ", input)

	out, _ := pipeline.action.(action.SyncAction).Run(ctx.Background(), input.(map[string]interface{}))

	context.SetOutput(Output, out)

	log.Info("[CMLRunnerActivity:Eval] Output : ", out)

	log.Info("[CMLRunnerActivity:Eval] Exit ........ ")

	return true, nil
}

func (a *CMLRunnerActivity) getPipeline(context activity.Context) (*Pipeline, error) {
	myId := util.ActivityId(context)
	pipeline := a.pipelineRegistry[a.activityToModel[myId]]

	if nil == pipeline {
		a.mux.Lock()
		defer a.mux.Unlock()
		pipeline = a.pipelineRegistry[a.activityToModel[myId]]
		if nil == pipeline {
			factory := action.GetFactory("github.com/project-flogo/catalystml-flogo/action")
			isURL, ok := context.GetSetting(ISURL)
			if !ok {
				log.Warn("(CMLRunnerActivity.getPipeline) isURL not set, will set to false.")
				isURL = false
			}

			cml, exist := context.GetSetting(CML)
			if exist {
				var act action.Action
				var err error
				if !isURL.(bool) {
					log.Info("(CMLRunnerActivity.getPipeline) processing (non-url mode) CML = ", cml)
					act, err = factory.New(&action.Config{Settings: map[string]interface{}{"catalystMl": cml.(string)}})
				} else {
					log.Info("(CMLRunnerActivity.getPipeline) processing (url mode) CML = ", cml)
					act, err = factory.New(&action.Config{Settings: map[string]interface{}{"catalystMlURI": cml.(string)}})
				}
				if nil != err {
					return nil, err
				}
				log.Info("(CMLRunnerActivity.getPipeline) act = ", act)
				pipeline = &Pipeline{
					inputDef: a.buildInputData(myId, context),
					action:   act,
				}
				a.pipelineRegistry[myId] = pipeline
			} else {
				return nil, fmt.Errorf("(CMLRunnerActivity.getPipeline) CML not set!!")
			}
		}
	}

	return pipeline, nil
}

func (a *CMLRunnerActivity) buildInputData(myId string, context activity.Context) map[string]*Field {
	inputData := make(map[string]*Field)
	inputFieldnames, _ := context.GetSetting("InputSchema")
	log.Info("Processing handlers :InputSchema = ", inputFieldnames)

	for _, inputFieldname := range inputFieldnames.([]interface{}) {
		inputFieldnameInfo := inputFieldname.(map[string]interface{})
		attribute := &Field{}
		attribute.SetName(inputFieldnameInfo["FieldName"].(string))
		attribute.SetType(inputFieldnameInfo["Type"].(string))
		attribute.SetOptional(nil != inputFieldnameInfo["Optional"] && "no" == inputFieldnameInfo["Optional"].(string))
		if nil != inputFieldnameInfo["Default"] && "" != inputFieldnameInfo["Default"].(string) {
			//attribute.SetDValue()
		}
		inputData[attribute.GetName()] = attribute
	}
	return inputData
}

type Field struct {
	name     string
	dValue   interface{}
	dataType string
	optional bool
}

func (this *Field) SetName(name string) {
	this.name = name
}

func (this *Field) GetName() string {
	return this.name
}

func (this *Field) SetDValue(dValue string) {
	this.dValue = dValue
}

func (this *Field) GetDValue() interface{} {
	return this.dValue
}

func (this *Field) SetType(dataType string) {
	this.dataType = dataType
}

func (this *Field) GetType() string {
	return this.dataType
}

func (this *Field) SetOptional(optional bool) {
	this.optional = optional
}

func (this *Field) IsOptional() bool {
	return this.optional
}

type Pipeline struct {
	inputDef map[string]*Field
	action   action.Action
}
