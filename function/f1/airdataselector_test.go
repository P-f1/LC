package f1

import (
	"fmt"
	"testing"
)

func TestFnAirDataSelector(t *testing.T) {
	f2 := &fnAirDataSelector{}

	reading := map[string]interface{}{
		"id":        "0001",
		"origin":    1234567890,
		"device":    "camera",
		"name":      "backyard",
		"value":     "FDSGFAGFADFDERYET%W#$TQ$TQ$",
		"valueType": "binary",
		"mediaType": "image/jpeg",
	}

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
			"name":     "Inferred",
			"value":    "{ \"Result\" : \"car\" }",
			"type":     "string",
		},
		map[string]interface{}{
			"producer": "Inference.REST",
			"name":     "Inferred",
			"value":    "{ \"Result\" : \"car\" }",
			"type":     "string",
		},
	}
	//	newf1, _ := f2.Eval("g0", reading, enriched, "{\"data\": \"@f1..value@\", \"extrafield\":\"abd\"}")
	//newf1, _ := f2.Eval("g0", reading, enriched, "@f1..value@")
	//newf1, _ := f2.Eval("g0", reading, enriched, "@producer01.consumer23.Inferred@")

	newf1, _ := f2.Eval("g0", reading, enriched, "@Inference.REST..Inferred@")
	fmt.Println("\n\nResult -> ", newf1)
}
