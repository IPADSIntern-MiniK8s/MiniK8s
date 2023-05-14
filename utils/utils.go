package utils

/* ========== Resource Function ========== */

func IsLabelEqual(a map[string]string, b map[string]string) bool {
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

/* ========== Time Function ========== */

func WaitForever() {
	<-make(chan struct{})
}
