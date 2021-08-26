/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package pipelinebuilder

import (
	b64 "encoding/base64"
	"encoding/json"

	"errors"
	"io/ioutil"

	"fmt"
	"strings"
	"sync"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/logger"
)

var log = logger.GetLogger("tibco-model-ops-pipelinebuilder")

var initialized bool = false

const (
	sTemplateFolder                = "TemplateFolder"
	iApplicationPipelineDescriptor = "ApplicationPipelineDescriptor"
	oFlogoApplicationDescriptor    = "FlogoApplicationDescriptor"
)

type PipelineBuilderActivity struct {
	metadata  *activity.Metadata
	mux       sync.Mutex
	templates map[string]*FlogoTemplateLibrary
}

func NewActivity(metadata *activity.Metadata) activity.Activity {
	aPipelineBuilderActivity := &PipelineBuilderActivity{
		metadata:  metadata,
		templates: make(map[string]*FlogoTemplateLibrary),
	}

	return aPipelineBuilderActivity
}

func (a *PipelineBuilderActivity) Metadata() *activity.Metadata {
	return a.metadata
}

func (a *PipelineBuilderActivity) Eval(context activity.Context) (done bool, err error) {

	log.Info("[PipelineBuilderActivity:Eval] entering ........ ")

	templateLibrary, err := a.getTemplateLibrary(context)
	if err != nil {
		return false, err
	}

	applicationPipelineDescriptorStr, ok := context.GetInput(iApplicationPipelineDescriptor).(string)
	if !ok {
		return false, errors.New("Invalid command ... ")
	}
	log.Info("[PipelineBuilderActivity:Eval]  Pipeline Descriptor : ", applicationPipelineDescriptorStr)

	var applicationPipelineDescriptor map[string]interface{}
	json.Unmarshal([]byte(applicationPipelineDescriptorStr), &applicationPipelineDescriptor)
	if nil != err {
		return true, nil
	}

	elementMap := make(map[string]*Element)
	contributes := NewContributes()
	flogoPropertyExtractor := NewFlogoPropertyExtractor()

	/* application name */
	elementMap["root.name"] = &Element{
		name:     "name",
		dValue:   applicationPipelineDescriptor["name"].(string),
		dataType: "string",
	}

	/* populate triggers */
	flogoTrigger := templateLibrary.getTrigger(applicationPipelineDescriptor["trigger"].(map[string]interface{}))
	flogoPropertyExtractor.extract(flogoTrigger["descriptor"].(map[string]interface{}))
	elementMap["root.triggers[]"] = &Element{
		name:     "triggers",
		dValue:   []interface{}{flogoTrigger["descriptor"]},
		dataType: "[]interface {}",
	}
	contributes.add(flogoTrigger["contribute"].(map[string]interface{}))

	activities := applicationPipelineDescriptor["activities"].([]interface{})
	/* populate links tasks */
	links := make([]interface{}, 0)
	tasks := make([]interface{}, 0)
	var previousFlogoActivityDesc map[string]interface{}
	for index, activity := range activities {
		flogoActivity := templateLibrary.getActivity(previousFlogoActivityDesc, activity.(map[string]interface{}))
		flogoActivityDesc := flogoActivity["descriptor"].(map[string]interface{})
		log.Info("[PipelineBuilderActivity:Eval]  index = ", index, ", activity = ", activity)
		log.Info("[PipelineBuilderActivity:Eval]  flogoActivity id = ", flogoActivityDesc["id"], ", previousFlogoActivity id = ", previousFlogoActivityDesc["id"])
		if 0 != index {
			links = append(links, map[string]interface{}{
				"id":   index,
				"from": previousFlogoActivityDesc["id"],
				"to":   flogoActivityDesc["id"],
				"type": "default",
			})
		}
		tasks = append(tasks, flogoActivityDesc)
		contributes.add(flogoActivity["contribute"].(map[string]interface{}))
		flogoPropertyExtractor.extract(flogoActivityDesc)
		previousFlogoActivityDesc = flogoActivityDesc
	}
	elementMap["root.resources[0].data.links[]"] = &Element{
		name:     "links",
		dValue:   links,
		dataType: "[]interface {}",
	}
	elementMap["root.resources[0].data.tasks[]"] = &Element{
		name:     "tasks",
		dValue:   tasks,
		dataType: "[]interface {}",
	}

	/* populate properties */
	elementMap["root.properties[]"] = &Element{
		name:     "properties",
		dValue:   flogoPropertyExtractor.GetData(),
		dataType: "[]interface {}",
	}

	/* populate contributes */
	/*
		[
			{"ref":"git.tibco.com/git/product/ipaas/wi-mqtt.git/Mqtt","s3location":"Tibco/Mqtt"},
			{"ref":"github.com/TIBCOSoftware/ModelOps","s3location":"{USERID}/ModelOps"},
			{"ref":"github.com/TIBCOSoftware/GraphBuilder_Tools","s3location":"{USERID}/GraphBuilder_Tools"}
		]
	*/
	elementMap["root.contrib"] = &Element{
		name:     "contrib",
		dValue:   contributes.getString(),
		dataType: "string",
	}

	builder := NewFlogoAppBuilder(elementMap)
	newPipeline := builder.Build(builder, templateLibrary.pipeline)

	fmt.Println("newPipeline = ", newPipeline)

	jsondata, err := json.Marshal(newPipeline)
	if nil != err {
		fmt.Println("Unable to serialize object, reason : ", err.Error())
		return true, err
	}

	fmt.Println(string(jsondata))
	context.SetOutput(oFlogoApplicationDescriptor, string(jsondata))

	log.Info("[PipelineBuilderActivity:Eval] Exit ........ ")

	return true, nil
}

func (a *PipelineBuilderActivity) BuildFlogoApp(
	applicationPipelineDescriptor map[string]interface{},
	templateLibrary *FlogoTemplateLibrary,
) (string, error) {

	log.Info("[PipelineBuilderActivity:BuildFlogoApp] entering ........ ")

	elementMap := make(map[string]*Element)
	contributes := NewContributes()
	flogoPropertyExtractor := NewFlogoPropertyExtractor()

	/* application name */
	elementMap["root.name"] = &Element{
		name:     "name",
		dValue:   applicationPipelineDescriptor["name"].(string),
		dataType: "string",
	}

	/* populate triggers */
	flogoTrigger := templateLibrary.getTrigger(applicationPipelineDescriptor["trigger"].(map[string]interface{}))
	flogoPropertyExtractor.extract(flogoTrigger["descriptor"].(map[string]interface{}))
	elementMap["root.triggers[]"] = &Element{
		name:     "triggers",
		dValue:   []interface{}{flogoTrigger["descriptor"]},
		dataType: "[]interface {}",
	}
	contributes.add(flogoTrigger["contribute"].(map[string]interface{}))

	activities := applicationPipelineDescriptor["activities"].([]interface{})
	/* populate links tasks */
	links := make([]interface{}, 0)
	tasks := make([]interface{}, 0)
	var previousFlogoActivityDesc map[string]interface{}
	for index, activity := range activities {
		flogoActivity := templateLibrary.getActivity(previousFlogoActivityDesc, activity.(map[string]interface{}))
		flogoActivityDesc := flogoActivity["descriptor"].(map[string]interface{})
		log.Info("[PipelineBuilderActivity:Eval]  index = ", index, ", activity = ", activity)
		log.Info("[PipelineBuilderActivity:Eval]  flogoActivity id = ", flogoActivityDesc["id"], ", previousFlogoActivity id = ", previousFlogoActivityDesc["id"])
		if 0 != index {
			links = append(links, map[string]interface{}{
				"id":   index,
				"from": previousFlogoActivityDesc["id"],
				"to":   flogoActivityDesc["id"],
				"type": "default",
			})
		}
		tasks = append(tasks, flogoActivityDesc)
		contributes.add(flogoActivity["contribute"].(map[string]interface{}))
		flogoPropertyExtractor.extract(flogoActivityDesc)
		previousFlogoActivityDesc = flogoActivityDesc
	}
	elementMap["root.resources[0].data.links[]"] = &Element{
		name:     "links",
		dValue:   links,
		dataType: "[]interface {}",
	}
	elementMap["root.resources[0].data.tasks[]"] = &Element{
		name:     "tasks",
		dValue:   tasks,
		dataType: "[]interface {}",
	}

	/* populate properties */
	elementMap["root.properties[]"] = &Element{
		name:     "properties",
		dValue:   flogoPropertyExtractor.GetData(),
		dataType: "[]interface {}",
	}

	/* populate contributes */
	/*
		[
			{"ref":"git.tibco.com/git/product/ipaas/wi-mqtt.git/Mqtt","s3location":"Tibco/Mqtt"},
			{"ref":"github.com/TIBCOSoftware/ModelOps","s3location":"{USERID}/ModelOps"},
			{"ref":"github.com/TIBCOSoftware/GraphBuilder_Tools","s3location":"{USERID}/GraphBuilder_Tools"}
		]
	*/
	elementMap["root.contrib"] = &Element{
		name:     "contrib",
		dValue:   contributes.getString(),
		dataType: "string",
	}

	builder := NewFlogoAppBuilder(elementMap)
	newPipeline := builder.Build(builder, templateLibrary.pipeline)

	fmt.Println("newPipeline = ", newPipeline)

	jsondata, err := json.Marshal(newPipeline)
	if nil != err {
		fmt.Println("Unable to serialize object, reason : ", err.Error())
		return "", err
	}
	log.Info("[PipelineBuilderActivity:Eval] Exit ........ ")

	return string(jsondata), nil
}

//func (a *PipelineBuilderActivity) BuildDeployDescriptor(context activity.Context) (map[string]interface{}, err error) {
//	return nil, nil
//}

func (a *PipelineBuilderActivity) getTemplateLibrary(ctx activity.Context) (*FlogoTemplateLibrary, error) {

	log.Info("[PipelineBuilderActivity:getTemplate] entering ........ ")
	defer log.Info("[PipelineBuilderActivity:getTemplate] exit ........ ")

	myId := ActivityId(ctx)
	templateLib := a.templates[myId]

	var err error
	if nil == templateLib {
		a.mux.Lock()
		defer a.mux.Unlock()
		templateLib = a.templates[myId]
		if nil == templateLib {
			templateFolderSetting, exist := ctx.GetSetting(sTemplateFolder)
			if !exist {
				return nil, activity.NewError("Template is not configured", "PipelineBuilder-4002", nil)
			}
			templateFolder := templateFolderSetting.(string)
			templateLib, err = NewFlogoTemplateLibrary(templateFolder)
			if nil != err {
				return nil, err
			}
			a.templates[myId] = templateLib
		}
	}
	return a.templates[myId], nil
}

func NewFlogoTemplateLibrary(templateFolder string) (*FlogoTemplateLibrary, error) {
	pipeline, err := buildTemplate(fmt.Sprintf("%s/pipeline.json", templateFolder))
	if nil != err {
		return nil, err
	}
	triggers := buildTemplates(fmt.Sprintf("%s/triggers", templateFolder))
	activities := buildTemplates(fmt.Sprintf("%s/activities", templateFolder))

	templateLib := &FlogoTemplateLibrary{
		pipeline:   pipeline,
		triggers:   triggers,
		activities: activities,
	}
	return templateLib, nil
}

type FlogoTemplateLibrary struct {
	pipeline   map[string]interface{}
	triggers   map[string]interface{}
	activities map[string]interface{}
}

func (this *FlogoTemplateLibrary) getPipeline() map[string]interface{} {
	return deepCopy(this.pipeline)
}

func (this *FlogoTemplateLibrary) getTrigger(descriptor map[string]interface{}) map[string]interface{} {
	name := descriptor["name"].(string)
	triggerCopy := deepCopy(this.triggers[name].(map[string]interface{}))

	builder := NewFlogoTriggerBuilder()
	triggerCopy = builder.Build(builder, triggerCopy).(map[string]interface{})
	return triggerCopy
}

func (this *FlogoTemplateLibrary) getActivity(previousActivityDesc, activityDesc map[string]interface{}) map[string]interface{} {
	log.Info("[FlogoTemplateLibrary:getActivity] entering ........ ")
	name := activityDesc["name"].(string)
	activityCopy := deepCopy(this.activities[name].(map[string]interface{}))

	builder := NewFlogoActivityBuilder(previousActivityDesc)
	activityCopy = builder.Build(builder, activityCopy).(map[string]interface{})
	return activityCopy
}

func ActivityId(ctx activity.Context) string {
	return fmt.Sprintf("%s_%s", ctx.FlowDetails().Name(), ctx.TaskName())
}

func deepCopy(a map[string]interface{}) map[string]interface{} {
	var b map[string]interface{}
	byt, _ := json.Marshal(a)
	json.Unmarshal(byt, &b)
	return b
}

func buildTemplates(folder string) map[string]interface{} {
	templates := make(map[string]interface{})
	files, _ := ioutil.ReadDir(folder)
	for index := range files {
		template, err := buildTemplate(fmt.Sprintf("%s/%s", folder, files[index].Name()))

		if nil != err {
			continue
		}
		filename := files[index].Name()
		name := filename[:strings.LastIndex(filename, ".")]
		templates[name] = template
	}

	return templates
}

func buildTemplate(filename string) (map[string]interface{}, error) {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("File reading error", err)
		return nil, err
	}

	var result map[string]interface{}
	json.Unmarshal(fileContent, &result)

	if nil != err {
		log.Error(err)
	}

	log.Info("[PipelineBuilderActivity:buildTemplate] FlogoTemplate : filename = ", filename, ", template = ", result)
	return result, nil
}

func NewContributes() Contributes {
	return Contributes{contributeMap: make(map[string]map[string]interface{})}
}

type Contributes struct {
	contributeMap map[string]map[string]interface{}
}

func (this *Contributes) add(contribute map[string]interface{}) {
	this.contributeMap[contribute["ref"].(string)] = contribute
}

func (this *Contributes) getString() string {
	contributeArray := make([]map[string]interface{}, 0)
	for _, contribute := range this.contributeMap {
		contributeArray = append(contributeArray, contribute)
	}
	contributeArrayBytes, _ := json.Marshal(contributeArray)
	contributeArrayString := b64.URLEncoding.EncodeToString(contributeArrayBytes)
	return contributeArrayString
}
