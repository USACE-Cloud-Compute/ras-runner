package actions

import (
	"testing"
)

const oneBreachBFile string = "/workspaces/cc-ras-runner/actions/testResources/DamBreachOverlapDem.b01"

func TestBreaching(t *testing.T) {
	updateBFile(oneBreachBFile)
}
