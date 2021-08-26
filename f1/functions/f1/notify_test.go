package f1

import (
	"fmt"
	"testing"
)

func TestFnAirNotify(t *testing.T) {
	f2 := &fnNotify{}

	reading := map[string]interface{}{
		"device": "device-pos-rest",
		"id":     "a6b1dec8-43e3-4f36-aa78-be9dedbbc215",
		"name":   "basket-open",
		"origin": 1619451567637760300,
		"value":  "basket_id:abc-012345-def,customer_id:joe5,employee_id:mary1,event_time:1.619451564791e+12,lane_id:1,", "valueType": "String",
	}

	enriched := []interface{}{
		map[string]interface{}{
			"producer": "Inference.REST",
			"name":     "Inferred",
			"value":    "{\"Data\":\"ERROR : Item not scanned!\"}\n",
			"type":     "string",
		},
	}

	match := []interface{}{
		map[string]interface{}{
			"type":  "contains",
			"value": "ERROR",
		},
		map[string]interface{}{
			"type":  "contains",
			"value": "WARN",
		},
	}

	newf1, _ := f2.Eval("g0", reading, enriched, "@Inference.REST..Inferred@", match, "NOTIFIER")
	fmt.Println("\n\nResult -> ", newf1)
}
