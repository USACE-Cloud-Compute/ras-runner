package utils

import (
	"fmt"
	"reflect"

	"github.com/usace/go-hdf5"
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

		ctype := hdf5.CompoundType{*dtype}
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
