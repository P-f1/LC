package f1

import (
	"github.com/P-f1/LC/labs-flogo-lib/util"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/expression/function"
)

func init() {
	function.Register(&fnCombineProperties{})
}

type fnGetProperty struct {
}

func (fnGetProperty) Name() string {
	return "getproperty"
}

func (fnGetProperty) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{data.TypeArray, data.TypeString}, false
}

func (fnGetProperty) Eval(params ...interface{}) (interface{}, error) {
	if nil != params[0] && nil != params[1] {
		for _, prop := range params[0].([]interface{}) {
			targetName := params[1].(string)
			name := util.GetPropertyElementAsString("Name", prop)
			if name == targetName {
				return util.GetPropertyElement("Value", prop), nil
			}
		}
	}

	return nil, nil
}
