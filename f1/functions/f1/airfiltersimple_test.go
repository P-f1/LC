package f1

import (
	"fmt"
	"testing"

	"github.com/project-flogo/core/data/expression/function"
	"github.com/stretchr/testify/assert"
)

func TestFnAirFilterSimple_Eval(t *testing.T) {
	filter := &fnAirFilterSimple{}
	j2o := &fnJson2Object{}
	reading := map[string]interface{}{
		"id":        "0001",
		"origin":    "1234567890",
		"device":    "d1",
		"name":      "r1",
		"value":     "200",
		"valueType": "int",
		"mediaType": "",
	}
	conditions, err01 := function.Eval(j2o, "[{\"name\":\"r2\", \"device\":\"d1\"}, {\"name\":\"r1\", \"device\":\"d2\"}]")
	fmt.Println("#### ", conditions)
	assert.Nil(t, err01)
	v, err02 := function.Eval(filter, reading, conditions)
	assert.Nil(t, err02)
	fmt.Println("#### ", v)
}

func TestFnAirFirstTrue_Eval(t *testing.T) {
	filter := &fnAirFirstTrue{}
	j2o := &fnJson2Object{}
	reading := map[string]interface{}{
		"id":        "0001",
		"origin":    "1234567890",
		"device":    "d2",
		"name":      "r1",
		"value":     "200",
		"valueType": "int",
		"mediaType": "",
	}
	conditions, err01 := function.Eval(j2o, "[{\"name\":\"r2\", \"device\":\"d1\"}, {\"name\":\"r1\", \"device\":\"d2\"}]")
	fmt.Println("#### ", conditions)
	assert.Nil(t, err01)
	v, err02 := function.Eval(filter, reading, conditions)
	assert.Nil(t, err02)
	fmt.Println("#### ", v)
}
