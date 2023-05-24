package workflow

import (
	"github.com/tidwall/gjson"
	"minik8s/pkg/apiobject"
	"minik8s/pkg/serverless/activator"
	"testing"

	"bou.ke/monkey"
)

func TestParseParams(t *testing.T) {
	// test1: only integer, nothing need to ignore
	params := []byte(`{"x": 4, "y": 5}`)
	inputPath := "$.x,$.y"
	result, err := ParseParams(params, inputPath)
	if err != nil {
		t.Errorf("ParseParams failed, error: %s", err)
	}
	t.Logf("result: %s", string(result))

	// test2: only integer, but need to ignore
	params = []byte(`{"x": 4, "y": 5}`)
	inputPath = "$.x"
	result, err = ParseParams(params, inputPath)
	if err != nil {
		t.Errorf("ParseParams failed, error: %s", err)
	}
	t.Logf("result: %s", string(result))

	// test3: integer and string, nothing need to ignore
	params = []byte(`{"x": 4, "str": "hello"}`)
	inputPath = "$.x,$.str"
	result, err = ParseParams(params, inputPath)
	if err != nil {
		t.Errorf("ParseParams failed, error: %s", err)
	}
	t.Logf("result: %s", string(result))

	// test4: integer and string, but need to ignore
	params = []byte(`{"x": 4, "str": "hello"}`)
	inputPath = "$.x"
	result, err = ParseParams(params, inputPath)
	if err != nil {
		t.Errorf("ParseParams failed, error: %s", err)
	}
	t.Logf("result: %s", string(result))
}


func TestHasField(t *testing.T) {
	value := 5
	chElem := apiobject.ChoiceItem {
		Variable: "$.z",
		NumericGreaterThan: &value,
		Next: "PrintSum",
	}
	if !HasField(chElem, "NumericGreaterThan") {
		t.Errorf("HasField failed, error: %s, %d", "NumericGreaterThan", chElem.NumericGreaterThan)
	}
	if HasField(chElem, "NumericLessThan") {
		t.Errorf("HasField failed, error: %s, %d", "NumericLessThan", chElem.NumericLessThan)
	}
}


func TestExecuteChoice(t *testing.T) {
	value1 := 1
	value2 := 2
	choice := apiobject.ChoiceState {
			Type: "Choice",
			Choices: []apiobject.ChoiceItem {
				{
					Variable: "$.foo",
					NumericEquals: &value1,
					Next: "FirstMatchState",
				},
				{
					Variable: "$.foo",
					NumericEquals: &value2,
					Next: "SecondMatchState",
				},
			},
			Default: "DefaultState",
	}

	params := []byte(`{"foo": 1}`)
	next, err := ExecuteChoice(choice, params)
	if err != nil {
		t.Errorf("ExecuteChoice failed, error: %s", err)
	}
	t.Logf("next: %s", next)

	params = []byte(`{"foo": 2}`)
	next, err = ExecuteChoice(choice, params)
	if err != nil {
		t.Errorf("ExecuteChoice failed, error: %s", err)
	}


	// the second test
	value := 5
	choice = apiobject.ChoiceState {
		Type: "Choice",
		Choices: []apiobject.ChoiceItem {
			{
				Variable: "$.z",
				NumericGreaterThan: &value,
				Next: "PrintSum",
			},
			{
				Variable: "$.z",
				NumericLessThan: &value,
				Next: "GetDiff",
			},
		},
		Default: "PrintError",
	}

	params = []byte(`{"z": 6}`)
	next, err = ExecuteChoice(choice, params)
	if err != nil {
		t.Errorf("ExecuteChoice failed, error: %s", err)
	}
	if next != "PrintSum" {
		t.Errorf("ExecuteChoice failed, error: %s", next)
	}

	params = []byte(`{"z": 4}`)
	next, err = ExecuteChoice(choice, params)
	if err != nil {
		t.Errorf("ExecuteChoice failed, error: %s", err)
	}
	if next != "GetDiff" {
		t.Errorf("ExecuteChoice failed, error: %s", next)
	}

	params = []byte(`{"z": 5}`)
	next, err = ExecuteChoice(choice, params)
	if err != nil {
		t.Errorf("ExecuteChoice failed, error: %s", err)
	}
	if next != "PrintError" {
		t.Errorf("ExecuteChoice failed, error: %s", next)
	}

	// the third test, check string
	strValue := "hello"
	choice = apiobject.ChoiceState {
		Type: "Choice",
		Choices: []apiobject.ChoiceItem {
			{
				Variable: "$.str",
				StringEquals: &strValue,
				Next: "PrintSum",
			},
			{
				Variable: "$.str",
				StringGreaterThan: &strValue,
				Next: "PrintSum",
			},
			{
				Variable: "$.str",
				StringLessThan: &strValue,
				Next: "GetDiff",
			},
		},
		Default: "PrintError",
	}


	params = []byte(`{"str": "hello"}`)
	next, err = ExecuteChoice(choice, params)
	if err != nil {
		t.Errorf("ExecuteChoice failed, error: %s", err)
	}
	if next != "PrintSum" {
		t.Errorf("ExecuteChoice failed, error: %s", next)
	}

	params = []byte(`{"str": "world"}`)
	next, err = ExecuteChoice(choice, params)
	if err != nil {
		t.Errorf("ExecuteChoice failed, error: %s", err)
	}
	if next != "PrintSum" {
		t.Errorf("ExecuteChoice failed, error: %s", next)
	}

	params = []byte(`{"str": "hell"}`)
	next, err = ExecuteChoice(choice, params)
	if err != nil {
		t.Errorf("ExecuteChoice failed, error: %s", err)
	}
	if next != "GetDiff" {
		t.Errorf("ExecuteChoice failed, error: %s", next)
	}

	params = []byte(`{"str": "world hello"}`)
	next, err = ExecuteChoice(choice, params)
	if err != nil {
		t.Errorf("ExecuteChoice failed, error: %s", err)
	}
	if next != "PrintSum" {
		t.Errorf("ExecuteChoice failed, error: %s", next)
	}
}

func GenerateWorkflow() apiobject.WorkFlow {
	value := 5
	example := apiobject.WorkFlow {
		Kind : "Workflow",
		APIVersion: "v1",
		Comment: "An example of basic workflow.",
		StartAt: "getsum",
		States: map[string]apiobject.State {
			"getsum": apiobject.TaskState {
				Type: "Task",
				InputPath: "$.x,$.y",
				Next: "judgesum",
			},
			"judgesum": apiobject.ChoiceState {
				Type: "Choice",
				Choices: []apiobject.ChoiceItem {
					{
						Variable: "$.z",
						NumericGreaterThan: &value,
						Next: "printsum",
					},
					{
						Variable: "$.z",
						NumericLessThan: &value,
						Next: "getdiff",
					},
				},
				Default: "printerror",
			},
			"printsum": apiobject.TaskState {
				Type: "Task",
				InputPath: "$.z",
				ResultPath: "$.str",
				End: true,
			},
			"getdiff": apiobject.TaskState {
				Type: "Task",
				InputPath: "$.x,$.y,$.z",
				Next: "printdiff",
			},
			"printdiff": apiobject.TaskState {
				Type: "Task",
				InputPath: "$.z",
				ResultPath: "$.str",
				End: true,
			},
			"printerror": apiobject.FailState {
				Type: "Fail",
				Error: "DefaultStateError",
				Cause: "No Matches!",
			},
		},
	}
	return example
}

func TestExecuteWorkFlow(t *testing.T) {
	monkey.Patch(activator.TriggerFunc, func(string, []byte)([]byte, error) {
		return []byte(`{"z":3, "x":4, "y":5, "str": "hello world"}`), nil
	})

	workflow := GenerateWorkflow()
	params := []byte(`{"x": 4, "y": 5}`)
	result, err := ExecuteWorkFlow(&workflow, params)
	if err != nil {
		t.Errorf("ExecuteWorkFlow failed, error: %s", err)
	}
	t.Logf("result: %s", string(result))
}

func TestGetParam(t *testing.T) {
	data := `"{"z": 9, "x": 5, "y": 4}"`
	path := "$.x"
	result := gjson.Get(data, path[2:])
	if !result.Exists() {
		t.Errorf("GetParam failed, error: %s", "result is not exist")
	}
}