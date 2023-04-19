package pod

import (
	"fmt"
	"minik8s/pkg/apiobject"
	"strconv"
	"strings"
)

func parseMemory(m string) (uint64, error) {
	loc := strings.Index(m, "Mi")
	if loc == -1 {
		res, err := strconv.ParseUint(m, 10, 64)
		return res, err
	}
	res, err := strconv.ParseUint(m[:loc], 10, 64)
	return res * 1024 * 1024, err
}

func parseCmd(cmd []string, args []string) []string {
	if cmd == nil && args == nil {
		return []string{}
	}
	if cmd == nil {
		//invalid
		return []string{}
	}
	res := make([]string, len(cmd)+len(args))
	copy(res, cmd)
	if args != nil {
		res = append(cmd, args...)
	}
	return res
}

func parseEnv(envs []apiobject.Env) []string {
	res := make([]string, len(envs), len(envs))
	for i, env := range envs {
		res[i] = fmt.Sprintf("%s=%s", env.Name, env.Value)
	}
	return res
}
