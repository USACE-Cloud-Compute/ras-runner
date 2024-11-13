package ras

import "testing"

func Test(t *testing.T) {
	actual, err := ReadSNetIDToNameFromGeoHDF("/workspaces/cc-ras-runner/testData/Duwamish_17110013.g01.hdf")
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
