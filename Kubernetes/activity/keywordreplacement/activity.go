/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package keywordreplacement

import (
	b64 "encoding/base64"
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/TIBCOSoftware/labs-graphbuilder-lib/util"
)

// activityLogger is the default logger for the Filter Activity
var log = logger.GetLogger("keyword-replacement")

const (
	Template   = "Template"
	LeftToken  = "LeftToken"
	RightToken = "RightToken"
	Mapping    = "Mapping"
	iVariables = "Variables"
	oDocument  = "Document"
)

// Mapping is an Activity that is used to Filter a message to the console
type MappingActivity struct {
	metadata       *activity.Metadata
	initialized    bool
	keywordMappers map[string]*KeywordMapper
	fieldMappers   map[string]map[string]string
	mux            sync.Mutex
}

// NewActivity creates a new AppActivity
func NewActivity(metadata *activity.Metadata) activity.Activity {
	aCSVParserActivity := &MappingActivity{
		metadata:       metadata,
		keywordMappers: make(map[string]*KeywordMapper),
		fieldMappers:   make(map[string]map[string]string),
	}
	return aCSVParserActivity
}

// Metadata returns the activity's metadata
func (a *MappingActivity) Metadata() *activity.Metadata {
	return a.metadata
}

// Eval implements api.Activity.Eval - Filters the Message
func (a *MappingActivity) Eval(ctx activity.Context) (done bool, err error) {

	keywordMapper, fieldMapper, err := a.getKeywordMapper(ctx)
	if nil != err {
		return false, err
	}
	log.Info("field mapping = ", fieldMapper)

	log.Info("input data = ", ctx.GetInput(iVariables))
	
	tupleIn := ctx.GetInput(iVariables).(*data.ComplexObject).Value.(map[string]interface{})
	keywordMap := make(map[string]interface{})
	for variableKey, value := range tupleIn {
		key := fieldMapper[variableKey]
		if "" != key {
			keywordMap[key] = value
		}
	}
	log.Info("keyword map = ", keywordMap)

	ctx.SetOutput("Document", keywordMapper.replace("", keywordMap))

	return true, nil
}

func (a *MappingActivity) init(context activity.Context) error {

	return nil
}

func (a *MappingActivity) getKeywordMapper(ctx activity.Context) (*KeywordMapper, map[string]string, error) {
	myId := util.ActivityId(ctx)
	mapper := a.keywordMappers[myId]
	fieldMapper := a.fieldMappers[myId]

	if nil == mapper {
		a.mux.Lock()
		defer a.mux.Unlock()
		mapper = a.keywordMappers[myId]
		fieldMapper = a.fieldMappers[myId]
		if nil == mapper {
			fieldMapper = make(map[string]string)
			mappings, _ := ctx.GetSetting(Mapping)
			log.Info("Processing handlers : mappings = ", mappings)
			for _, mapping := range mappings.([]interface{}) {
				mappingInfo := mapping.(map[string]interface{})
				fieldMapper[mappingInfo["Variable"].(string)] = mappingInfo["Keyword"].(string)
				//mappingInfo["Type"].(string)
			}

			templateObj, exist := ctx.GetSetting(Template)
			if !exist {
				return nil, nil, errors.New("Template not defined!")
			}
			encodedTemplate, _ := data.CoerceToObject(templateObj)
			template, _ := b64.StdEncoding.DecodeString(strings.Split(encodedTemplate["content"].(string), ",")[1])
			
			lefttoken, exist := ctx.GetSetting(LeftToken)
			if !exist {
				return nil, nil, errors.New("LeftToken not defined!")
			}
			righttoken, exist := ctx.GetSetting(RightToken)
			if !exist {
				return nil, nil, errors.New("RightToken not defined!")
			}
			mapper = NewKeywordMapper(string(template), lefttoken.(string), righttoken.(string))

			a.keywordMappers[myId] = mapper
			a.fieldMappers[myId] = fieldMapper
		}
	}
	return mapper, fieldMapper, nil
}

type KeywordReplaceHandler struct {
	result     string
	keywordMap map[string]interface{}
}

func (this *KeywordReplaceHandler) setMap(keywordMap map[string]interface{}) {
	this.keywordMap = keywordMap
}

func (this *KeywordReplaceHandler) startToMap() {
	this.result = ""
}

func (this *KeywordReplaceHandler) replace(keyword string) string {
	if nil != this.keywordMap[keyword] {
		return this.keywordMap[keyword].(string)
	}
	return ""
}

func (this *KeywordReplaceHandler) endOfMapping(document string) {
	this.result = document
}

func (this *KeywordReplaceHandler) getResult() string {
	return this.result
}

func NewKeywordMapper(
	template string,
	lefttag string,
	righttag string) *KeywordMapper {
	mapper := KeywordMapper{
		template:     template,
		keywordOnly:  false,
		slefttag:     lefttag,
		srighttag:    righttag,
		slefttaglen:  len(lefttag),
		srighttaglen: len(righttag),
	}
	return &mapper
}

type KeywordMapper struct {
	template     string
	keywordOnly  bool
	slefttag     string
	srighttag    string
	slefttaglen  int
	srighttaglen int
	document     bytes.Buffer
	keyword      bytes.Buffer
	mh           KeywordReplaceHandler
}

func (this *KeywordMapper) replace(template string, keywordMap map[string]interface{}) string {
	if "" == template {
		template = this.template
		if "" == template {
			return ""
		}
	}

	log.Info("[KeywordMapper.replace] template = ", template)

	this.mh.setMap(keywordMap)
	this.document.Reset()
	this.keyword.Reset()

	scope := false
	boundary := false
	skeyword := ""
	svalue := ""

	this.mh.startToMap()
	for i := 0; i < len(template); i++ {
		//log.Infof("template[%d] = ", i, template[i])
		// maybe find a keyword beginning Tag - now isn't in a keyword
		if !scope && template[i] == this.slefttag[0] {
			if this.isATag(i, this.slefttag, template) {
				this.keyword.Reset()
				scope = true
			}
		} else if scope && template[i] == this.srighttag[0] {
			// maybe find a keyword ending Tag - now in a keyword
			if this.isATag(i, this.srighttag, template) {
				i = i + this.srighttaglen - 1
				skeyword = this.keyword.String()[this.slefttaglen:this.keyword.Len()]
				svalue = this.mh.replace(skeyword)
				if "" == svalue {
					svalue = fmt.Sprintf("%s%s%s", this.slefttag, skeyword, this.srighttag)
				}
				//log.Info("value ->", svalue);
				this.document.WriteString(svalue)
				boundary = true
				scope = false
			}
		}

		if !boundary {
			if !scope && !this.keywordOnly {
				this.document.WriteByte(template[i])
			} else {
				this.keyword.WriteByte(template[i])
			}
		} else {
			boundary = false
		}
		
		//log.Info("document = ", this.document)

	}
	this.mh.endOfMapping(this.document.String())
	return this.mh.getResult()
}

func (this *KeywordMapper) isATag(i int, tag string, template string) bool {
	for j := 0; j < len(tag); j++ {
		if tag[j] != template[i+j] {
			return false
		}
	}
	return true
}

/*
  public static void main(String[] argc)
  {
     KeywordMapper ikm = new KeywordMapper();
     //ikm.setMapping("KEY1", "REAL_KEY1");
     //ikm.setMapping("KEY2", "REAL_KEY2");
     //ikm.setMapping("KEY3", "REAL_KEY3");
     //ikm.setLeftTag("%");
     //ikm.setRightTag("%");
     System.out.println("Result ---> " + ikm.replace("parameter1=$KEY1$,parameter2=$KEY2$,parameter3=$KEY3$"));
  }
*/
