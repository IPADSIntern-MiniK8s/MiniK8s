package utils

/* ========== Resource Function ========== */

func IsPodFitSelector(selector map[string]string, pod map[string]string) bool {
	for k, v := range selector {
		if pod[k] != v {
			return false
		}
	}
	return true
}

/* ========== Time Function ========== */

func WaitForever() {
	<-make(chan struct{})
}
