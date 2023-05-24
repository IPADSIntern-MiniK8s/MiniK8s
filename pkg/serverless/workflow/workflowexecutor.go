package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/serverless/activator"
	"minik8s/utils"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/tidwall/gjson"
)

var epsilon = 1e-9

// CheckNode check the node is valid or not by checking the function
func CheckNode(nodeName string) bool {
	funcUrl := "http://localhost:8080/api/v1/functions/" + nodeName
	_, err :=  utils.SendRequest("GET", nil, funcUrl)
	if err != nil {
		return false
	}
	return true
}



// ParseParams fileter the params by the inputPath or resultPath
func ParseParams(params []byte, inputPath string) ([]byte, error) {
	wanted := strings.Split(inputPath, ",")

	filterdParams := make(map[string]interface{})
	for _, elem := range wanted {
		name := elem[2:]
		value := gjson.Get(string(params), name)
		if !value.Exists() {
			log.Error("[ParseParams] the params is not valid, the name is: ", name, ", the params is: ", string(params))
			return nil, errors.New("the params is not valid")
		}
		filterdParams[name] = value.Value()
	}

	jsonData, err := json.Marshal(filterdParams)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

// HasField check the field is in the struct or not
func HasField(obj interface{}, fieldName string) bool {
	t := reflect.ValueOf(obj)
	value := t.FieldByName(fieldName)
	return value.Kind() == reflect.Ptr && !value.IsNil()
}


// isNumeric check whether the variable's type is numeric
func isNumeric(variable interface{}) bool {
	switch variable.(type) {
		// actually, if use gjson to get the value, the type is float64 default
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64,	complex64, complex128:
		return true
	default:
		return false
	}
}

// isString check whether the variable's type is string
func isString(variable interface{}) bool {
	switch variable.(type) {
	case string:
		return true
	default:
		return false
	}
}


// replaceSingleQuotesWithDoubleQuotes: replace the single quotes with double quotes
func replaceSingleQuotesWithDoubleQuotes(str string) string {
	// the default string in dict is single quotes, need to replace it with double quotes
	bytes := []byte(str)
	for i := 0; i < len(bytes); i++ {
		if bytes[i] == '\'' {
			bytes[i] = '"'
		}
	}
	return string(bytes)
}

func ExecuteWorkFlow(workflow *apiobject.WorkFlow, params []byte) ([]byte, error) {
	// traverse the workflow
	startNode := workflow.StartAt
	if startNode == "" {
		return nil, errors.New("workflow start node is empty")
	}

	currentNode, ok := workflow.States[startNode]
	currentName := startNode
	if !ok {
		log.Error("[Execute]")
		return nil, errors.New("workflow start node is not valid")
	}

	for {
		prevName := currentName
		err := error(nil)
		switch currentNode.(type) {
		case apiobject.TaskState: {
			params, err = ExecuteTask(currentNode.(apiobject.TaskState), currentName, params)
			if err != nil {
				return nil, err
			}
			if currentNode.(apiobject.TaskState).End {
				return params, nil
			}
			if currentNode.(apiobject.TaskState).Next == "" {
				return nil, errors.New("the next node is empty")
			}
			currentName = currentNode.(apiobject.TaskState).Next
		}
		case apiobject.ChoiceState: {
			currentName, err = ExecuteChoice(currentNode.(apiobject.ChoiceState), params)
			if err != nil {
				return nil, err
			}
		}
		case apiobject.FailState: {
			result := ExecuteFail(currentNode.(apiobject.FailState))
			return []byte(result), nil
		}
		default: {
			return nil, errors.New("the workflow is not valid")
		}
		}

		currentNode = workflow.States[currentName]

		// don't allow loop now
		if currentNode == nil || prevName == currentName {
			break
		}
	}
	return nil, errors.New("the workflow is not valid")
}





// ExecuteTask execute the task node and return the result
func ExecuteTask(task apiobject.TaskState, functionName string, params []byte) ([]byte, error) {
	// get the function name
	if functionName == "" {
		return nil, errors.New("task resource is empty")
	}

	// check the function is valid or not
	if !CheckNode(functionName) {
		return nil, errors.New("function is not valid")
	}

	// try to trigger the function
	// if the InputPath is not empty, need to parse the params to abstract the input
	inputParams := params
	err := error(nil)
	if task.InputPath != "" {
		inputParams, err = ParseParams(params, task.InputPath)
		if err != nil {
			return nil, err
		}
	}

	result, err := activator.TriggerFunc(functionName, inputParams)
	if err != nil {
		return nil, err
	}
	
	// python's dict is single quotes, need to replace it with double quotes
	paramsStr := string(result)
	paramsStr = replaceSingleQuotesWithDoubleQuotes(paramsStr)
	result = []byte(paramsStr)

	// if the ResultPath is not empty, need to parse the result to abstract the output
	if task.ResultPath != "" {
		result, err = ParseParams(result, task.ResultPath)

		if err != nil {
			return nil, err
		}
	}
	return result, nil
}


// ExecuteChoice TODO: need to 'AND' and 'OR' later
func ExecuteChoice(choice apiobject.ChoiceState, params []byte) (string, error) {
	// get the val of the variable
	for _, chElem := range(choice.Choices) {
		variable := gjson.Get(string(params), chElem.Variable[2:])
		if !variable.Exists() {
			return "", errors.New("the variable is not exist")
		}
		value := variable.Value()
		log.Info("[ExecuteChoice] value: ", value)
		// judge the condition
		if HasField(chElem, "NumericEquals") {
			val, ok := value.(float64)
			if ok && math.Abs(float64(*chElem.NumericEquals) - val) < epsilon {
				log.Info("[ExecuteChoice] return")
				return chElem.Next, nil
			}
		} else if HasField(chElem, "StringEquals") {
			if isString(value) && *chElem.StringEquals == value {
				return chElem.Next, nil
			}
		} else if HasField(chElem, "NumericNotEquals") {
			val, ok := value.(float64)
			if ok && math.Abs(float64(*chElem.NumericNotEquals) - val) > epsilon {
				return chElem.Next, nil
			}
		} else if HasField(chElem, "StringNotEquals") {
			
			if isString(value) && *chElem.StringNotEquals != value {
				return chElem.Next, nil
			}
		} else if HasField(chElem, "NumericLessThan") {
			val, ok := value.(float64)
			if ok && float64(*chElem.NumericLessThan) > val {
				return chElem.Next, nil
			}
		} else if HasField(chElem, "StringLessThan") {
			val, ok := value.(string)
			if ok && *chElem.StringLessThan > val {
				return chElem.Next, nil
			}
		} else if HasField(chElem, "NumericGreaterThan") {
			val, ok := value.(float64)
			if ok && float64(*chElem.NumericGreaterThan) < val {
				return chElem.Next, nil
			}
		} else if HasField(chElem, "StringGreaterThan") {
			val, ok := value.(string)
			if ok && *chElem.StringGreaterThan < val {
				return chElem.Next, nil
			}
		}
	}
	log.Info("[ExecuteChoice] default: ", choice.Default)
	if choice.Default != "" {
		return choice.Default, nil
	}
	return "", errors.New("the choice is not valid")
}


func ExecuteFail(fail apiobject.FailState) string {
	result := fmt.Sprintf("Fail: %s, Cause: %s", fail.Error, fail.Cause)
	return result
}
