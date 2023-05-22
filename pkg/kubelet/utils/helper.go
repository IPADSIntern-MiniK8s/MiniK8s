package utils

import (
	"fmt"
	"github.com/containerd/containerd"
	"os/exec"
	"strings"
)

var runtimePath, _ = exec.LookPath("nerdctl")

func Ctl(namespace string, args ...string) (string, error) {
	//fmt.Println(append([]string{"-n", namespace}, args...))
	res, err := exec.Command(runtimePath, append([]string{"-n", namespace}, args...)...).CombinedOutput()
	return string(res), err
}

func CheckCmd(namespace string, containerName string, args []string, answer string) bool {
	output, _ := Ctl(namespace, append([]string{"exec", containerName}, args...)...)
	return strings.Index(output, answer) > -1
}

func GetInfo(namespace, containerName, fields string) (string, error) {
	output, err := Ctl(namespace, "inspect", "-f", fmt.Sprintf("{{%s}}", fields), containerName)
	if err != nil {
		return "", err
	}
	//remove the last \n
	return output[:len(output)-1], nil
}

func NewClient(namespace string) (*containerd.Client, error) {
	return containerd.New("/run/containerd/containerd.sock", containerd.WithDefaultNamespace(namespace))
}
