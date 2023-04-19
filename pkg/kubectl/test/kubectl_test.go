package cmd

import (
	"bytes"
	"minik8s/pkg/kubectl/cmd"
	"testing"
)

func TestApply(t *testing.T) {
	/* usable only when api-server is on */
	actual := new(bytes.Buffer)
	cmd.RootCmd.SetOut(actual)
	cmd.RootCmd.SetErr(actual)
	cmd.RootCmd.SetArgs([]string{"apply", "-f", "D:\\mini-k8s\\pkg\\kubectl\\test\\test.yaml"})
	cmd.RootCmd.Execute()

	//expected := "This-is-command-a1"
	//fmt.Print(actual.String())
	//
	//assert.Equal(t, actual.String(), expected, "actual is not expected")
}
