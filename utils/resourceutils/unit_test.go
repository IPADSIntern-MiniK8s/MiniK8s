package resourceutils

import "testing"
func TestUnit(t *testing.T) {
	// test getUnit
	str := "100m"
	unit := GetUnit(str)
	if unit != "m" {
		t.Errorf("getUnit error, expect m, get %s", unit)
	}

	str = "100Mi"
	unit = GetUnit(str)
	if unit != "Mi" {
		t.Errorf("getUnit error, expect Mi, get %s", unit)
	}
	
}


func TestPackQuantity(t *testing.T) {
	// test PackQuantity
	quantity := 0.1
	unit := "m"
	str := PackQuantity(quantity, unit)
	t.Logf("str: %s", str)

	quantity = 100.0
	unit = "Mi"
	str = PackQuantity(quantity, unit)
	t.Logf("str: %s", str)
}