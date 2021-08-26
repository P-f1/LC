/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package tgdbupsert

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"tgdb"
	f "tgdb/factory"

	"git.tibco.com/git/product/ipaas/wi-contrib.git/connection/generic"
	"github.com/P-f1/LC/flogo-lib/core/activity"
	"github.com/P-f1/LC/flogo-lib/core/data"
	"github.com/P-f1/LC/flogo-lib/logger"
	"github.com/P-f1/LC/labs-graphbuilder-lib/dbservice"
	"github.com/P-f1/LC/labs-graphbuilder-lib/dbservice/factory"
	"github.com/P-f1/LC/labs-graphbuilder-lib/util"
)

const (
	Connection = "tgdbConnection"
	KeepAlive  = "keepAlive"
)

var log = logger.GetLogger("tibco-activity-tgdbupsert")

type TGDBUpsertActivity struct {
	metadata          *activity.Metadata
	activityToService map[string]dbservice.UpsertService
	mux               sync.Mutex
}

func NewActivity(metadata *activity.Metadata) activity.Activity {
	return &TGDBUpsertActivity{
		metadata:          metadata,
		activityToService: make(map[string]dbservice.UpsertService),
	}
}

func (a *TGDBUpsertActivity) Metadata() *activity.Metadata {
	return a.metadata
}

func (a *TGDBUpsertActivity) Eval(context activity.Context) (done bool, err error) {

	log.Info("(TGDBUpsertActivity) entering ......")
	defer log.Info("(TGDBUpsertActivity) exit ......")

	tgdbService, err := a.getTGDBService(context)
	if nil != err {
		return false, err
	}

	iInputData := context.GetInput("Graph")
	if nil == iInputData {
		return false, errors.New("Illegal nil graph data")
	}

	inputData, ok := iInputData.(map[string]interface{})
	if !ok {
		return false, errors.New("Illegal graph data type, should be map[string]interface{}.")
	}

	iGraph := inputData["graph"]
	if nil == iGraph {
		return false, errors.New("Illegal nil graph content")
	}

	graph, ok := iGraph.(map[string]interface{})
	if !ok {
		return false, errors.New("Illegal graph content, should be map[string]interface{}.")
	}

	//connect()
	err = tgdbService.UpsertGraph(nil, graph)
	if nil != err {
		return false, err
	}

	return true, nil
}

func connect() {
	url := "tcp://192.168.1.152:8222/{dbName=housedb}"
	user := "napoleon"
	passwd := "bonaparte"
	memberName := "Napoleon Bonaparte"

	cf := f.GetConnectionFactory()
	conn, err := cf.CreateAdminConnection(url, user, passwd, nil)
	if err != nil {
		fmt.Printf("connection error: %s, %s\n", err.GetErrorCode(), err.GetErrorMsg())
		os.Exit(1)
	}
	conn.Connect()
	defer conn.Disconnect()

	gof, err := conn.GetGraphObjectFactory()
	if err != nil {
		fmt.Printf("graph object factory error: %s, %s\n", err.GetErrorCode(), err.GetErrorMsg())
		os.Exit(1)
	}
	if gmd, err := conn.GetGraphMetadata(true); err == nil {
		fmt.Printf("graph metadata: %v\n", gmd)
	}

	key, err := gof.CreateCompositeKey("houseMemberType")
	key.SetOrCreateAttribute("memberName", memberName)
	fmt.Printf("search house member: %s\n", memberName)
	member, err := conn.GetEntity(key, nil)
	if err != nil {
		fmt.Printf("Failed to fetch member: %v\n", err)
	}
	if member != nil {
		if attrs, err := member.GetAttributes(); err == nil {
			for _, v := range attrs {
				fmt.Printf("Member attribute %s => %v\n", v.GetName(), v.GetValue())
			}
			if node, ok := member.(tgdb.TGNode); ok {
				edges := node.GetEdges()
				fmt.Printf("check relationships: %d\n", len(edges))
				for _, edge := range edges {
					n := edge.GetVertices()
					fmt.Printf("relationship '%v': %v -> %v\n", edge.GetAttribute("relType").GetValue(),
						n[0].GetAttribute("memberName").GetValue(), n[1].GetAttribute("memberName").GetValue())
				}
			}
		}
	}
}

func (a *TGDBUpsertActivity) getTGDBService(context activity.Context) (dbservice.UpsertService, error) {
	myId := util.ActivityId(context)

	tgdbService := a.activityToService[myId]
	//tgdb.GetFactory().GetService(a.activityToConnector[myId])
	if nil == tgdbService {
		a.mux.Lock()
		defer a.mux.Unlock()
		tgdbService = a.activityToService[myId]
		//tgdb.GetFactory().GetService(a.activityToConnector[myId])
		if nil == tgdbService {
			log.Info("Initializing TGDB Service start ...")
			defer log.Info("Initializing TGDB Service done ...")

			connection, exist := context.GetSetting(Connection)
			if !exist {
				return nil, activity.NewError("TGDB connection is not configured", "TGDB-UPSERT-4001", nil)
			}

			genericConn, err := generic.NewConnection(connection)
			if err != nil {
				return nil, err
			}

			var connectorName string
			var apiVersion string
			properties := make(map[string]interface{})
			for name, value := range genericConn.Settings() {
				switch name {
				case "url":
					properties["url"], _ = data.CoerceToString(value)
				case "user":
					properties["user"], _ = data.CoerceToString(value)
				case "password":
					properties["password"], _ = data.CoerceToString(value)
				case KeepAlive:
					properties[KeepAlive], _ = data.CoerceToBoolean(value)
				case "name":
					connectorName, _ = data.CoerceToString(value)
				case "apiVersion":
					apiVersion, _ = data.CoerceToString(value)
				}
			}

			allowEmptyStringKey, exist := context.GetSetting("allowEmptyStringKey")
			if exist {
				properties["allowEmptyStringKey"] = allowEmptyStringKey
			} else {
				log.Warn("allowEmptyStringKey configuration is not configured, will make type defininated implicit!")
			}

			log.Info("(getTGDBService) - properties = ", properties)

			tgdbService, err = factory.GetFactory(dbservice.TGDB, apiVersion).CreateUpsertService(connectorName, properties)
			a.activityToService[myId] = tgdbService
			if nil != err {
				return nil, err
			}
		}
	}

	return tgdbService, nil
}
