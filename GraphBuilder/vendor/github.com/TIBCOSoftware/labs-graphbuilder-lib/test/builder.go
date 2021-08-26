// builder
package main

import (
	"fmt"
	"log"

	"github.com/P-f1/LC/labs-graphbuilder-lib/model"
)

func main() {
	fmt.Println("Hello World!")
	modelName := "MeetUp"
	jsonmodel := ""
	nodes := "[map[Group:[map[Name:SpanGlish!! Spanish/English Language Exchange]] Category:[map[Name:language/ethnic identity]] Country:[map[Country_Code:gb]] Meet_Up:[map[Name:Eat out to Help out Scheme. Night out in English & Spanish. Only August]] State_Province:[map[State_Province_Code:<nil> Country_Code:gb]] City:[map[Country_Code:gb City_Name:London State_Province_Code:<nil>]] Address:[map[Street_Address:To Be Confirmed  City_Name:London State_Province_Code:<nil> Country_Code:gb]] Location:[map[UID:TBC Name:TBC]]]]"
	edges := "[map[Contains_Meet_Up:[] in_Country:[] Take_place:[] in_State_Province:[] in_City:[] on:[] Contains_Group:[]]]"
	var err error
	graphModel, err := model.NewGraphModel(modelName, jsonmodel)
	if nil != err {
		log.Fatalln(err)
	}
	graphId := graphModel.GetId()

	graphBuilder := model.NewGraphBuilder()
	deltaGraph := graphBuilder.CreateGraph(graphId, graphModel)
	err = graphBuilder.BuildGraph(
		&deltaGraph,
		graphModel,
		context.GetInput(Nodes).(*data.ComplexObject).Value,
		context.GetInput(Edges).(*data.ComplexObject).Value,
		false,
	)
	if nil != err {
		log.Fatalln(err)
	}

}
