package container

import (
	"os/exec"
)

var runtimePath, _ = exec.LookPath("nerdctl")

func ctl(args ...string) (string, error) {
	res, err := exec.Command(runtimePath, append([]string{"-n", "minik8s"}, args...)...).CombinedOutput()
	return string(res), err
}
