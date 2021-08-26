/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package pipelinebuilder

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

/* GOLangObjectHandler : for extract properties */
const (
	PROPERTY_PREFIX = "=$property[\""
	PROPERTY_SUFFIX = "\"]"
)

func NewFlogoPropertyExtractor() FlogoPropertyExtractor {
	return FlogoPropertyExtractor{
		properties: make(map[string]interface{}),
	}
}

type FlogoPropertyExtractor struct {
	properties map[string]interface{}
}

func (this FlogoPropertyExtractor) extract(element map[string]interface{}) {
	objectWalker := NewGOLangObjectWalker(this)
	objectWalker.Start(element)
}

func (this FlogoPropertyExtractor) reset() {
	this.properties = make(map[string]interface{})
}

func (this FlogoPropertyExtractor) GetData() []map[string]interface{} {
	propertyArr := make([]map[string]interface{}, len(this.properties))
	index := 0
	for _, property := range this.properties {
		propertyArr[index] = map[string]interface{}{
			"name":  property,
			"type":  "string",
			"value": "None",
		}
		index++
	}
	return propertyArr
}

func (this FlogoPropertyExtractor) HandleElements(namespace ElementId, element interface{}, dataType interface{}) interface{} {
	elementIds := namespace.GetId()
	log.Debug("Handle : id = ", elementIds, ", element = ", element, ", type = ", dataType)
	if "string" == dataType && strings.HasPrefix(element.(string), PROPERTY_PREFIX) {
		this.properties[element.(string)] = element.(string)[len(PROPERTY_PREFIX) : len(element.(string))-len(PROPERTY_SUFFIX)]
	}

	return nil
}

/* GOLangObjectHandler : for build app */
type FlogoBuilder struct {
}

func (this FlogoBuilder) Build(handler GOLangObjectHandler, jsonData interface{}) interface{} {
	objectWalker := NewGOLangObjectWalker(handler)
	return objectWalker.Start(jsonData)
}

func (this FlogoBuilder) GetData() []map[string]interface{} {
	log.Info("(FlogoBuilder.GetData) Should be overrided!!")
	return nil
}

func (this FlogoBuilder) HandleElements(namespace ElementId, element interface{}, dataType interface{}) interface{} {
	log.Info("(FlogoBuilder.HandleElements) Should be overrided!!")
	return nil
}

/* Build trigger */

func NewFlogoTriggerBuilder() *FlogoTriggerBuilder {
	return &FlogoTriggerBuilder{}
}

type FlogoTriggerBuilder struct {
	FlogoBuilder
}

func (this FlogoTriggerBuilder) HandleElements(namespace ElementId, element interface{}, dataType interface{}) interface{} {
	return nil
}

/* Build activity */

func NewFlogoActivityBuilder(upstreamComponent map[string]interface{}) *FlogoActivityBuilder {
	return &FlogoActivityBuilder{
		upstreamComponent: upstreamComponent,
	}
}

const (
	UPSTREAM_DATA_PREFIX = "=$activity["
	FLOW_DATA_PREFIX     = "=$flow."
)

type FlogoActivityBuilder struct {
	FlogoBuilder
	upstreamComponent map[string]interface{}
}

func (this FlogoActivityBuilder) HandleElements(namespace ElementId, element interface{}, dataType interface{}) interface{} {
	elementIds := namespace.GetId()
	log.Info("(FlogoActivityBuilder.HandleElements) Handle : id = ", elementIds, ", element = ", element, ", type = ", dataType)
	if "string" == dataType {
		if nil != this.upstreamComponent && strings.HasPrefix(element.(string), UPSTREAM_DATA_PREFIX) {
			suffix := element.(string)[strings.Index(element.(string), "]")]
			upstreamComponentName := this.upstreamComponent["name"].(string)
			return fmt.Sprintf("%s%s%s", UPSTREAM_DATA_PREFIX, upstreamComponentName, string(suffix))
		} else if strings.HasPrefix(element.(string), FLOW_DATA_PREFIX) {

		}
	}
	return nil
}

/* Build pipeline */

func NewFlogoAppBuilder(arrtibuteMap map[string]*Element) *FlogoAppBuilder {
	return &FlogoAppBuilder{
		arrtibuteMap: arrtibuteMap,
	}
}

type FlogoAppBuilder struct {
	FlogoBuilder
	arrtibuteMap map[string]*Element
}

func (this FlogoAppBuilder) HandleElements(namespace ElementId, element interface{}, dataType interface{}) interface{} {
	elementIds := namespace.GetId()
	log.Debug("Handle : id = ", elementIds, ", element = ", element, ", type = ", dataType)
	for _, elementId := range elementIds {
		elementDef := this.arrtibuteMap[elementId]
		log.Debug("map = ", this.arrtibuteMap, ", key = ", elementId, ", value = ", elementDef)
		if nil != elementDef {
			//elementName := elementDef.GetName()
			//elementDataType := elementDef.GetType()
			//fmt.Println(
			//	"Name : ", elementName,
			//	", Defined Type : ", elementDataType,
			//	", Original Type : ", reflect.TypeOf(element).String(),
			//	", Value (before) : ", element,
			//)
			return elementDef.GetDValue()
		}
	}
	return nil
}

/* GOLang Object Processing Framework */

type Scope struct {
	index    int
	maxIndex int
	name     string
	array    bool
}

type ElementId struct {
	namespace []Scope
	name      interface{}
}

func (this *ElementId) GetIndex() int {
	return this.namespace[len(this.namespace)-1].index
}

func (this *ElementId) GetId() []string {
	ids := make([]string, 1)

	var buffer bytes.Buffer
	arrayElement := false
	for i := range this.namespace {
		if !arrayElement {
			if 0 != i {
				buffer.WriteString(".")
			}
			buffer.WriteString(this.namespace[i].name)
		} else {
			arrayElement = false
		}

		if this.namespace[i].array {
			buffer.WriteString("[")
			if -1 < this.namespace[i].index {
				buffer.WriteString(strconv.Itoa(this.namespace[i].index))
			}
			buffer.WriteString("]")
			arrayElement = true
		}
	}
	if nil != this.name {
		buffer.WriteString(".")
		buffer.WriteString(this.name.(string))
	}
	ids[0] = buffer.String()
	return ids
}

func (this *ElementId) SetName(name string) {
	this.name = name
}

func (this *ElementId) updateIndex(index int, maxIndex int) {
	log.Debug("   Before updateIndex : ", this.namespace, ", index : ", index)
	this.namespace[len(this.namespace)-1].index = index
	this.namespace[len(this.namespace)-1].maxIndex = maxIndex
	log.Debug("   After updateIndex : ", this.namespace, ", index : ", index)
}

func (this *ElementId) enterScope(scopename string, isArray bool) {
	log.Debug("Before enterScope : ", this.namespace, ", ID = ", this.GetId()) //, ", index : ", this.namespace[len(this.namespace)-1].index)
	this.name = nil
	this.namespace = append(this.namespace, Scope{name: scopename, array: isArray, index: -1, maxIndex: -1})
	log.Debug("After enterScope : ", this.namespace, ", ID = ", this.GetId()) //, ", index : ", this.namespace[len(this.namespace)-1].index)
}

func (this *ElementId) leaveScope(scopename string, isArray bool) {
	log.Debug("Before leaveScope : ", this.namespace, ", ID = ", this.GetId()) //, ", index : ", this.namespace[len(this.namespace)-1].index)
	this.namespace = this.namespace[:len(this.namespace)-1]
	this.name = nil
	log.Debug("After leaveScope : ", this.namespace, ", ID = ", this.GetId()) //, ", index : ", this.namespace[len(this.namespace)-1].index)
}

type Element struct {
	name     string
	dValue   interface{}
	dataType string
}

func (this *Element) SetName(name string) {
	this.name = name
}

func (this *Element) GetName() string {
	return this.name
}

func (this *Element) SetDValue(dValue string) {
	this.dValue = dValue
}

func (this *Element) GetDValue() interface{} {
	return this.dValue
}

func (this *Element) SetType(dataType string) {
	this.dataType = dataType
}

func (this *Element) GetType() string {
	return this.dataType
}

type GOLangObjectHandler interface {
	HandleElements(namespace ElementId, element interface{}, dataType interface{}) interface{}
	GetData() []map[string]interface{}
}

type GOLangObjectWalker struct {
	ElementId
	currentLevel int
	jsonHandler  GOLangObjectHandler
}

func NewGOLangObjectWalker(jsonHandler GOLangObjectHandler) GOLangObjectWalker {
	GOLangObjectWalker := GOLangObjectWalker{
		currentLevel: 0,
		jsonHandler:  jsonHandler}
	GOLangObjectWalker.ElementId = ElementId{
		namespace: make([]Scope, 0),
	}

	return GOLangObjectWalker
}

func (this *GOLangObjectWalker) Start(jsonData interface{}) interface{} {
	this.walk("root", jsonData)
	return jsonData
}

func (this *GOLangObjectWalker) walk(name string, data interface{}) interface{} {

	var modifiedData interface{}
	switch data.(type) {
	case []interface{}:
		{
			this.startArray(name)
			modifiedData = this.jsonHandler.HandleElements(this.ElementId, data, "[]interface{}")
			if nil == modifiedData {
				dataArray := data.([]interface{})
				maxIndex := len(dataArray) - 1
				for index, subdata := range dataArray {
					this.updateIndex(index, maxIndex)
					this.walk(name, subdata)
				}
				this.updateIndex(-1, -1)
			}
			this.endArray(name)
			break
		}
	case map[string]interface{}:
		{
			this.startObject(name)
			modifiedData = this.jsonHandler.HandleElements(this.ElementId, data, "map[string]interface{}")
			if nil == modifiedData {
				modifiedMap := make(map[string]interface{})
				dataMap := data.(map[string]interface{})
				for subname, subdata := range dataMap {
					modifiedSubdata := this.walk(subname, subdata)
					if nil != modifiedSubdata {
						modifiedMap[subname] = modifiedSubdata
					}
				}
				for subname, subdata := range modifiedMap {
					dataMap[subname] = subdata
				}
			}
			this.endObject(name)
			break
		}
	default:
		{
			this.ElementId.SetName(name)
			modifiedData = this.jsonHandler.HandleElements(this.ElementId, data, reflect.TypeOf(data).String())
		}
	}
	return modifiedData
}

func (this *GOLangObjectWalker) startArray(name string) {
	log.Debug("Start Array before scope -> ", name, ", ", this.namespace)
	this.ElementId.enterScope(name, true)
	log.Debug("Start Array after scope -> ", name, ", ", this.namespace)
}

func (this *GOLangObjectWalker) endArray(name string) {
	log.Debug("End Array -> ", name)
	this.ElementId.leaveScope(name, true)
}

func (this *GOLangObjectWalker) startObject(name string) {
	this.ElementId.enterScope(name, false)
	log.Debug("Start Object -> ", name, ", ", this.namespace)
}

func (this *GOLangObjectWalker) endObject(name string) {
	log.Debug("End Object -> ", name)
	this.ElementId.leaveScope(name, false)
}
