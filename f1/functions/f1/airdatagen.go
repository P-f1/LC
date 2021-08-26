package f1

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/expression/function"
)

func init() {
	function.Register(&fnAirDataGen{})
}

type fnAirDataGen struct {
}

func (fnAirDataGen) Name() string {
	return "airdatagen"
}

func (fnAirDataGen) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{data.TypeString, data.TypeObject}, false
}

/*
{
	"gateway" : [
		{
			"type" : ["d1", "d2", "d3"],
			"name" : ["r1", "r2", "r3"],
			"value" : [100.0, 200.0]
		},
		{
			"type" : ["d4", "d5", "d6"],
			"name" : ["r4", "r5", "r6"],
			"value" : [5000.0, 8000.0]
		}
	]
}
{\"gateway\":[{"type\":[\"d1\", \"d2\", \"d3\"],\"name\":[\"r1\", \"r2\", \"r3\"],\"value\":[100.0, 200.0]},{\"type\":[\"d4\", \"d5\", \"d6\"],\"name\":[\"r4\", \"r5\", \"r6\"],\"value\":[5000.0, 8000.0]}]}
*/
func (fnAirDataGen) Eval(params ...interface{}) (interface{}, error) {
	gateway := params[0].(string)
	log.Info("(fnAirDataGen.Eval) in gateway ========>", gateway)
	log.Info("(fnAirDataGen.Eval) in gateway ========>", params[1])
	config := params[1].(map[string]interface{})[gateway].([]interface{})
	now := time.Now().UnixNano()
	r := rand.New(rand.NewSource(now))

	device := config[r.Int()%len(config)].(map[string]interface{})
	dTypes := device["type"].([]interface{})
	dNames := device["name"].([]interface{})
	dValues := device["value"].([]interface{})
	dValueMax := dValues[0].(float64)
	dValueMin := dValues[1].(float64)
	reading := map[string]interface{}{
		"id":        strconv.FormatInt(now, 10),
		"origin":    now,
		"device":    dTypes[r.Int()%len(dTypes)].(string),
		"name":      dNames[r.Int()%len(dTypes)].(string),
		"value":     fmt.Sprintf("%f", (r.Float64()*(dValueMax-dValueMin) + dValueMin)),
		"valueType": "float64",
		"mediaType": "data",
	}

	log.Info("(fnAirDataGen.Eval) out dataArray ========>", reading)
	return reading, nil
}
