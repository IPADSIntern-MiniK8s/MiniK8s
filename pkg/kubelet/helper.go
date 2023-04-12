package kubelet

import (
	"os/exec"
	"strings"
)

var runtimePath, _ = exec.LookPath("nerdctl")

func Ctl(args ...string) (string, error) {
	res, err := exec.Command(runtimePath, append([]string{"-n", "minik8s"}, args...)...).CombinedOutput()
	return string(res), err
}

func CheckCmd(containerName string, args []string, answer string) bool {
	output, _ := Ctl(append([]string{"exec", containerName}, args...)...)
	return strings.Index(output, answer) > -1
}
