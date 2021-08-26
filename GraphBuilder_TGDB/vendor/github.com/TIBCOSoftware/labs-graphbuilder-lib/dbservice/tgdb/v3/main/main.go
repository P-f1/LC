package main

import (
	"fmt"
	"os"

	"tgdb"
	"tgdb/factory"
)

func main() {
	url := "tcp://127.0.0.1:8222/{dbName=housedb}"
	user := "napoleon"
	passwd := "bonaparte"
	memberName := "Napoleon Bonaparte"

	cf := factory.GetConnectionFactory()
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
