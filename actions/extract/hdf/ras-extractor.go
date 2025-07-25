package hdf

import (
	"fmt"
	"reflect"

	"github.com/usace/go-hdf5"
	"github.com/usace/hdf5utils"
)

type RasExtractConfig struct {
	Match       string
	Postprocess []string
	Omitdata    bool
	Colnames    []string //optional: list of column names for writing output headers
	Coldata     string   //optional: data path to hdf dataset containing column names
	StringSizes []int
}

type RasExtractDataTypes interface {
	int | int8 | int16 | int32 | int64 | float32 | float64 | string
}

type ReadRasDataInput struct {
	DataPath string
	Config   RasExtractConfig

	//Dataset string
	//Summary int
}

type RasExtractReader[T RasExtractDataTypes] struct {
	f *hdf5.File
}

func NewRasExtractReader[T RasExtractDataTypes](filepath string) (*RasExtractReader[T], error) {
	f, err := hdf5utils.OpenFile(filepath)
	if err != nil {
		return nil, err
	}
	return &RasExtractReader[T]{f}, nil
}

type RasExtractData[T RasExtractDataTypes] struct {
	data      [][]T
	summaries map[string][]T
}

func (rr *RasExtractReader[T]) Read(input ReadRasDataInput) (RasExtractData[T], error) {
	data, err := rr.ReadArray(input)

	output := RasExtractData[T]{}
	output.summaries = make(map[string][]T)
	if err != nil {
		return output, err
	}

	if len(data) > 0 {
		output.data = data
		cols := len(data[0])
		colSummary := make([]T, cols)

		for k := 0; k < cols; k++ {
			colSummary[k] = columnMax(data, k)
		}
		output.summaries["max"] = colSummary
	}
	return output, nil
}

func (rr *RasExtractReader[T]) GroupMembers(groupPath string) ([]string, error) {
	group, err := hdf5utils.NewHdfGroup(rr.f, groupPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read the hdf group '%s': %s", groupPath, err)
	}
	defer group.Close()
	return group.ObjectNames()
}

func (rr *RasExtractReader[T]) ReadArray(input ReadRasDataInput) ([][]T, error) {
	var strSet hdf5utils.HdfStrSet

	if len(input.Config.StringSizes) > 0 {
		strSet = hdf5utils.NewHdfStrSet(input.Config.StringSizes...)
	}

	dataSet, err := hdf5utils.NewHdfDataset(string(input.DataPath), hdf5utils.HdfReadOptions{
		Dtype:        reflect.TypeFor[T]().Kind(),
		Strsizes:     strSet,
		File:         rr.f,
		ReadOnCreate: true,
	})

	if err != nil {
		return nil, err
	}

	return toSlice[T](dataSet)
}

func toSlice[T any](dataset *hdf5utils.HdfDataset) ([][]T, error) {
	rows := make([][]T, dataset.Rows())
	for i := range rows {
		dest := []T{}
		err := dataset.ReadRow(i, &dest)
		if err != nil {
			return nil, fmt.Errorf("failed to read row %d: %s", i, err)
		}
		rows[i] = dest
	}
	return rows, nil
}

func columnMax[T RasExtractDataTypes](data [][]T, col int) T {
	maxVal := data[0][col]
	for i := range data {
		if data[i][col] > maxVal {
			maxVal = data[i][col]
		}
	}
	return maxVal
}

// ////////////////
// ////////////////
type RasDataExtractWriter[T RasExtractDataTypes] interface {
	Write(RasExtractData[T]) error
}

type ConsoleRasExtractWriter[T RasExtractDataTypes] struct{}

func (rw *ConsoleRasExtractWriter[T]) Write(data RasExtractData[T]) error {
	fmt.Println(data)
	return nil
}
