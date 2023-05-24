package apiobject

import (
	"encoding/json"

	"github.com/tidwall/gjson"
)

// example:
// {
//   "Comment": "An example of the Amazon States Language using a choice state.",
//   "StartAt": "FirstState",
//   "States": {
//     "FirstState": {
//       "Type": "Task",
//       "Resource": "arn:aws:lambda:us-east-1:123456789012:function:FUNCTION_NAME",
//       "Next": "ChoiceState"
//     },
//     "ChoiceState": {
//       "Type" : "Choice",
//       "Choices": [
//         {
//           "Variable": "$.foo",
//           "NumericEquals": 1,
//           "Next": "FirstMatchState"
//         },
//         {
//           "Variable": "$.foo",
//           "NumericEquals": 2,
//           "Next": "SecondMatchState"
//         }
//       ],
//       "Default": "DefaultState"
//     },
//     "FirstMatchState": {
//       "Type" : "Task",
//       "Resource": "arn:aws:lambda:us-east-1:123456789012:function:OnFirstMatch",
//       "Next": "NextState"
//     },

//     "SecondMatchState": {
//       "Type" : "Task",
//       "Resource": "arn:aws:lambda:us-east-1:123456789012:function:OnSecondMatch",
//       "Next": "NextState"
//     },

//     "DefaultState": {
//       "Type": "Fail",
//       "Error": "DefaultStateError",
//       "Cause": "No Matches!"
//     },

//     "NextState": {
//       "Type": "Task",
//       "InputPath": "$.orderId, $.customer"
//    	 "ResultPath": "$.myResult"
//       "End": true
//     }
//   }
// }

type State interface {
}

type TaskState struct {
	Type StateType `json:"type"`
	InputPath string `json:"inputPath,omitempty"`
	ResultPath string `json:"outputPath,omitempty"`
	Next string `json:"next,omitempty"`
	End bool `json:"end,omitempty"`
}

type FailState struct {
	Type StateType `json:"type"`
	Error string `json:"error"`
	Cause string `json:"cause"`
}


type ChoiceState struct {
	Type StateType `json:"type"`
	Choices []ChoiceItem `json:"choices"`
	Default string `json:"default,omitempty"`
}


type ChoiceItem struct {
	Variable 			string `json:"variable"`
	NumericEquals 		*int `json:"NumericEquals,omitempty"`
	NumericNotEquals 	*int `json:"NumericNotEquals,omitempty"`
	NumericLessThan 	*int `json:"NumericLessThan,omitempty"`
	NumericGreaterThan 	*int `json:"NumericGreaterThan,omitempty"`
	StringEquals 		*string `json:"StringEquals,omitempty"`
	StringNotEquals		*string `json:"StringNotEquals,omitempty"`
	StringLessThan 		*string `json:"StringLessThan,omitempty"`
	StringGreaterThan 	*string `json:"StringGreaterThan,omitempty"`
	Next 				string `json:"next"`
}


type StateType string

const (
	Task StateType = "Task"
	Choice StateType = "Choice"
	Parallel StateType = "Parallel"
	Wait StateType = "Wait"
	Fail StateType = "Fail"
	Succeed StateType = "Succeed"
)

type WorkFlow struct {
	Kind 		string `json:"kind,omitempty"`
	APIVersion 	string `json:"apiVersion,omitempty"`

	Name 		string `json:"name"`
	Status 		VersionLabel `json:"status,omitempty"`
	StartAt 	string `json:"startAt"`

	States 		map[string]State `json:"states"`

	Comment 	string `json:"comment,omitempty"`
}


func (w *WorkFlow) MarshalJSON() ([]byte, error) {
	type Alias WorkFlow
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(w),
	})
}


func (w *WorkFlow) UnMarshalJSON(data []byte) error {
	// type Alias WorkFlow
	// aux := &struct {
	// 	*Alias
	// }{
	// 	Alias: (*Alias)(w),
	// }
	
	// if err := json.Unmarshal(data, &aux); err != nil {
	// 	return err
	// }
	w.Kind = "Workflow"
	w.APIVersion = gjson.Get(string(data), "apiVersion").String()
	w.Name = gjson.Get(string(data), "name").String()
	status := gjson.Get(string(data), "status")
	if status.Exists() {
		w.Status = VersionLabel(status.String())
	}
	w.StartAt = gjson.Get(string(data), "startAt").String()
	comment := gjson.Get(string(data), "comment")
	if comment.Exists() {
		w.Comment = comment.String()
	}
	states := gjson.Get(string(data), "states")
	if states.Exists() {
		w.States = make(map[string]State)
		states.ForEach(func(key, value gjson.Result) bool {
			stateType := gjson.Get(value.String(), "type").String()
			switch stateType {
			case "Task":
				var taskState TaskState
				err := json.Unmarshal([]byte(value.String()), &taskState)
				if err != nil {
					return false
				}
				w.States[key.String()] = taskState
			case "Choice":
				var choiceState ChoiceState
				err := json.Unmarshal([]byte(value.String()), &choiceState)
				if err != nil {
					return false
				}
				w.States[key.String()] = choiceState
			case "Fail":
				var failState FailState
				err := json.Unmarshal([]byte(value.String()), &failState)
				if err != nil {
					return false
				}
			}
			return true
		})
	}

	return nil
}