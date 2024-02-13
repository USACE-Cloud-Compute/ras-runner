package ras

import "testing"

func Test(t *testing.T) {
	actual, err := ReadSNetIDToNameFromGeoHDF(BALD_EAGLE_HDF_PATH)
	if err != nil {
		t.Fail()
	}
	if actual["Highway 120"] != 2 {
		t.Fail()
	}
	if len(actual) != 11 {
		t.Fail()
	}
}
