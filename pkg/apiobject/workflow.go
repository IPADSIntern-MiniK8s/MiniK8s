package apiobject

import "encoding/json"

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
	NumericEquals 		int `json:"NumericEquals,omitempty"`
	NumericNotEquals 	int `json:"NumericNotEquals,omitempty"`
	NumericLessThan 	int `json:"NumericLessThan,omitempty"`
	NumericGreaterThan 	int `json:"NumericGreaterThan,omitempty"`
	StringEquals 		string `json:"StringEquals,omitempty"`
	StringNotEquals		string `json:"StringNotEquals,omitempty"`
	StringLessThan 		string `json:"StringLessThan,omitempty"`
	StringGreaterThan 	string `json:"StringGreaterThan,omitempty"`
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
	APIVersion 	string `json:"apiVersion,omitempty"`

	Name 		string `json:"name"`
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
	type Alias WorkFlow
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(w),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}