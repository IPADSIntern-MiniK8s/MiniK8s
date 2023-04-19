package kubelet

import (
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
