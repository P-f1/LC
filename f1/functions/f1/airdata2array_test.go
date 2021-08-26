package f1

import (
	"fmt"
	"testing"
)

func TestFnAirData2Array(t *testing.T) {
	f2 := &fnAirData2Array{}

	reading := map[string]interface{}{
		"id":        "0001",
		"origin":    1234567890,
		"device":    "camera",
		"name":      "backyard",
		"value":     "FDSGFAGFADFDERYET%W#$TQ$TQ$",
		"valueType": "binary",
		"mediaType": "image/jpeg",
	}
	/*
		enriched := []interface{}{
			map[string]interface{}{
				"producer": "producer01",
				"consumer": "consumer23",
				"name":     "base64",
				"value":    "RTQEQ^N%B$Y^#$^N^",
				"type":     "string",
			},
			map[string]interface{}{
				"producer": "producer01",
				"consumer": "consumer23",
				"name":     "inference",
				"value":    "{ \"Result\" : \"car\" }",
				"type":     "string",
			},
		}
	*/
	newf1, _ := f2.Eval("g0", reading, nil)
	fmt.Println("\n\nResult -> ", newf1)
}
