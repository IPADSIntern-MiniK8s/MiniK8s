
package apiobject

import (
	"encoding/json"
	"testing"
)

func TestWorkflow(t *testing.T) {
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
						NumericEquals: 1,
						Next: "FirstMatchState",
					},
					{
						Variable: "$.foo",
						NumericEquals: 2,
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