
package apiobject

import (
	"encoding/json"
	"testing"
)

func TestWorkflow(t *testing.T) {
	value1 := 1
	value2 := 2
	example := WorkFlow {
		APIVersion: "v1",
		Comment: "An example of the Amazon States Language using a choice state.",
		StartAt: "FirstState",
		States: map[string]State {
			"FirstState": TaskState {
				Type: "Task",
				InputPath: "$.orderId, $.customer",
				ResultPath: "$.myResult",
				Next: "ChoiceState",
			},
			"ChoiceState": ChoiceState {
				Type: "Choice",
				Choices: []ChoiceItem {
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
			},
			"FirstMatchState": TaskState {
				Type: "Task",
				InputPath: "$.orderId, $.customer",
				ResultPath: "$.myResult",
				Next: "NextState",
			},
			"SecondMatchState": TaskState {
				Type: "Task",
				InputPath: "$.orderId, $.customer",
				ResultPath: "$.myResult",
				Next: "NextState",
			},
			"DefaultState": FailState {
				Type: "Fail",
				Error: "DefaultStateError",
				Cause: "No Matches!",
			},
			"NextState": TaskState {
				Type: "Task",
				InputPath: "$.orderId, $.customer",
				ResultPath: "$.myResult",
				End: true,
			},
		},
	}

	workflowJson, err := json.MarshalIndent(example, "", "    ")
	if err != nil {
		t.Errorf("GenerateWorkflow failed, error marshalling replicas: %s", err)
	}
	t.Logf("workflow: %s", workflowJson)
}


func TestWorkflowExample(t *testing.T) {
	value := 5
	example := WorkFlow {
		Kind : "Workflow",
		APIVersion: "v1",
		Comment: "An example of basic workflow.",
		StartAt: "getsum",
		States: map[string]State {
			"getsum": TaskState {
				Type: "Task",
				InputPath: "$.x,$.y",
				Next: "judgesum",
			},
			"judgesum": ChoiceState {
				Type: "Choice",
				Choices: []ChoiceItem {
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
			"printsum": TaskState {
				Type: "Task",
				InputPath: "$.z",
				ResultPath: "$.str",
				End: true,
			},
			"getdiff": TaskState {
				Type: "Task",
				InputPath: "$.x,$.y,$.z",
				Next: "printdiff",
			},
			"printdiff": TaskState {
				Type: "Task",
				InputPath: "$.z",
				ResultPath: "$.str",
				End: true,
			},
			"printerror": FailState {
				Type: "Fail",
				Error: "DefaultStateError",
				Cause: "No Matches!",
			},
		},
	}
	workflowJson, err := json.MarshalIndent(example, "", "    ")
	if err != nil {
		t.Errorf("GenerateWorkflow failed, error marshalling replicas: %s", err)
	}
	t.Logf("workflow: %s", workflowJson)

	workflowJson, err = example.MarshalJSON()
	// t.Logf(string(workflowJson))
	if err != nil {
		t.Errorf("GenerateWorkflow failed, error marshalling replicas: %s", err)
	}

	workflow := WorkFlow{}
	err = workflow.UnMarshalJSON(workflowJson)
	if err != nil {
		t.Errorf("GenerateWorkflow failed, error marshalling replicas: %s", err)
	}
}


func TestMarshal(t *testing.T) {
	rawData := []byte(`{"kind":"Workflow","apiVersion":"v1","name":"workflow-exp","status":"update","startAt":"getsum","states":{"getdiff":{"inputPath":"$.x,$.y,$.z","next":"printdiff","type":"Task"},"getsum":{"inputPath":"$.x,$.y","next":"judgesum","type":"Task"},"judgesum":{"choices":[{"NumericGreaterThan":5,"next":"printsum","variable":"$.z"},{"NumericLessThan":5,"next":"getdiff","variable":"$.z"}],"default":"printerror","type":"Choice"},"printdiff":{"end":true,"inputPath":"$.z","outputPath":"$.str","type":"Task"},"printerror":{"cause":"No Matches!","error":"DefaultStateError","type":"Fail"},"printsum":{"end":true,"inputPath":"$.z","outputPath":"$.str","type":"Task"}},"comment":"An example of basic workflow."}`)
	workflow := WorkFlow{}
	err := workflow.UnMarshalJSON(rawData)
	if err != nil {
		t.Errorf("GenerateWorkflow failed, error marshalling replicas: %s", err)
	}
	workflowJson, err := workflow.MarshalJSON()
	if err != nil {
		t.Errorf("GenerateWorkflow failed, error marshalling replicas: %s", err)
	}
	t.Logf("workflow: %s", workflowJson)
}