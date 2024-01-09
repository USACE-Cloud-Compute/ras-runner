package actions

import (
	"fmt"
	"testing"
)

const oneBreachBFile string = "/workspaces/cc-ras-runner/actions/testResources/DamBreachOverlapDem.b01"

func TestBreaching(t *testing.T) {
	updateBFile(oneBreachBFile)
	row, err := splitRowsIntoCells("000000010000004500000005", 8)
	if err != nil {
		fmt.Println(row[0:3])
	}

}
