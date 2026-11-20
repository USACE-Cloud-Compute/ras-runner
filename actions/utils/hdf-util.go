package utils

import (
	"fmt"
	"reflect"

	"github.com/usace-cloud-compute/filesapi"
	"github.com/usace-cloud-compute/go-hdf5"
)

type Hdf5MetadataType string

const (
	GroupMetadata    Hdf5MetadataType = "GROUP"
	CompoundMetadata Hdf5MetadataType = "COMPOUND"
	DatasetMetadata  Hdf5MetadataType = "DATASET"
)

type GoHdfAttr struct {
	AttrType reflect.Type
	AttrSize uint
}

func GetAttrMetadata(f *hdf5.File, metaType Hdf5MetadataType, metaPath string, metaField string) (*GoHdfAttr, error) {

	switch metaType {
	case DatasetMetadata:
		dset, err := f.OpenDataset(metaPath)
		if err != nil {
			return nil, err
		}
		defer dset.Close()

		dtype, err := dset.Datatype()
		if err != nil {
			return nil, err
		}
		defer dtype.Close()

		return &GoHdfAttr{
			AttrType: dtype.GoType(),
			AttrSize: dtype.Size(),
		}, nil
	case GroupMetadata:
		grp, err := f.OpenGroup(metaPath)
		if err != nil {
			return nil, err
		}
		defer grp.Close()

		attr, err := grp.OpenAttribute(metaField)
		if err != nil {
			return nil, err
		}
		defer attr.Close()

		attrType := attr.GetType()

		dt := hdf5.NewDatatype(attrType.HID())
		defer dt.Close()

		return &GoHdfAttr{
			AttrType: dt.GoType(),
			AttrSize: dt.Size(),
		}, nil

	case CompoundMetadata:
		dset, err := f.OpenDataset(metaPath)
		if err != nil {
			return nil, err
		}
		defer dset.Close()

		dtype, err := dset.Datatype()
		if err != nil {
			return nil, err
		}
		defer dtype.Close()

		ctype := hdf5.CompoundType{Datatype: *dtype}
		defer ctype.Close()

		nm := ctype.NMembers()

		for i := 0; i < nm; i++ {
			name := ctype.MemberName(i)
			if name == metaField {
				mftype, err := ctype.MemberType(i)
				if err != nil {
					return nil, err
				}
				defer mftype.Close()

				return &GoHdfAttr{
					AttrType: mftype.GoType(),
					AttrSize: mftype.Size(),
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("invalid meta type: %s", metaType)
}

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

func S3Fapl(awsconfig filesapi.S3FSConfig) hdf5.H5FD_ROS3_FAPL {
	fapl := hdf5.H5FD_ROS3_FAPL{
		Version:    1,
		AWS_REGION: awsconfig.S3Region,
	}

	if screds, ok := awsconfig.Credentials.(filesapi.S3FS_Static); ok {
		fapl.AWS_ACCESS_KEY_ID = screds.S3Id
		fapl.AWS_SECRET_ACCESS_KEY = screds.S3Key
	}

	if awsconfig.AltEndpoint != "" {
		fapl.AWS_ENDPOINT = awsconfig.AltEndpoint
	}
	return fapl
}
