package utils

import (
	"fmt"

	"github.com/usace-cloud-compute/go-hdf5"
)

type HdfGroup struct {
	groupPath string
	group     *hdf5.Group
}

func NewHdfGroup(f *hdf5.File, groupPath string) (*HdfGroup, error) {
	group, err := f.OpenGroup(groupPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open group '%s': %s", groupPath, err)
	}
	return &HdfGroup{groupPath, group}, nil
}

func (g *HdfGroup) ObjectNames() ([]string, error) {
	numberObjects, err := g.group.NumObjects()
	if err != nil {
		return nil, fmt.Errorf("unable to get the number of object in group '%s': %s", g.groupPath, err)
	}

	objectNames := make([]string, numberObjects)

	var i uint
	for i = 0; i < numberObjects; i++ {
		name, err := g.group.ObjectNameByIndex(i)
		if err != nil {
			return nil, fmt.Errorf("error reading name of object for group '%s' index %d : %s", g.groupPath, i, err)
		}
		objectNames[i] = name
	}
	return objectNames, nil
}

func (g *HdfGroup) Close() error {
	return g.group.Close()
}