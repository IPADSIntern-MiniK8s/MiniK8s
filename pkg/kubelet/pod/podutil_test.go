package pod

import (
	"minik8s/pkg/apiobject"
	"testing"
)

func TestParseMemory(t *testing.T) {
	res, err := parseMemory("100Mi")
	if err != nil || res != 100*1024*1024 {
		t.Fatalf("test parseMemory error")
	}
	res, err = parseMemory("10000")
	if err != nil || res != 10000 {
		t.Fatalf("test parseMemory failed")
	}
}

func TestParseCmd(t *testing.T) {
	res := parseCmd([]string{"/bin/bash"}, []string{"-c", "echo Hello Kubernetes!"})
	if len(res) != 3 || res[0] != "/bin/bash" || res[1] != "-c" || res[2] != "echo Hello Kubernetes!" {
		t.Fatalf("test parsecmd failed")
	}
}

func TestParseEnv(t *testing.T) {
	res := parseEnv([]apiobject.Env{
		{"a", "b"}, {"c", "d"},
	})
	if len(res) != 2 || res[0] != "a=b" || res[1] != "c=d" {
		t.Fatalf("test parseenv failed")
	}
}
