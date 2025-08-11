package hdf

import "github.com/usace/cc-go-sdk"

func init() {
	cc.ActionRegistry.RegisterAction("bcline-peak-outputs", &RasExtractAction{})
}

type RasExtractAction struct {
	cc.ActionRunnerBase
}

func (a *RasExtractAction) Run() error {
	return nil
}
