package actions

import (
	"errors"
	"fmt"
	"log"
	"os"
	"ras-runner/ras"
	"strings"

	"github.com/usace/cc-go-sdk"
	"github.com/usace/go-hdf5"
	"github.com/usace/hdf5utils"
)

// UpdateOutletTSAction reads an hdf file and uses it to read and write a bfile with updated outlet time series.
func UpdateOutletTSAction(action cc.Action, modelDir string) error {
	// Assumes bFile and hdf file  were copied local with the CopyLocal action.
	log.Printf("Ready to update bFile with new observed flows.")
	bFileName := action.Parameters.GetStringOrFail("bFile")
	outletTSName := action.Parameters.GetStringOrFail("outletTS")
	bfilePath := fmt.Sprintf("%v/%v", modelDir, bFileName)
	if !fileExists(bfilePath) {
		return errors.New(fmt.Sprintf("Input source %s, was not found in local directory. Run copy-local first", bfilePath))
	}
	bf, err := ras.InitBFile(bfilePath)
	if err != nil {
		return err
	}
	outletTSIdx := 0
	var outletTS *ras.OutletTS
	for idx, block := range bf.BfileBlocks {
		outts, ok := block.(*ras.OutletTS)
		if ok {
			if strings.Contains(outts.Name, outletTSName) {
				outletTSIdx = idx //will never be zero, that is the ras version block.
				outletTS = outts
			}
		}
	}
	if outletTSIdx == 0 {
		return errors.New("could not find the outlet TS named " + outletTSName)
	}
	hdfDataPath := action.Parameters.GetStringOrFail("hdfDataPath")
	hdfFileName := action.Parameters.GetStringOrFail("hdfFile")
	hdfFilePath := fmt.Sprintf("%v/%v", modelDir, hdfFileName)
	hdfReadOptions := hdf5utils.HdfReadOptions{
		Dtype:              0,
		Strsizes:           hdf5utils.HdfStrSet{},
		IncrementalRead:    false,
		IncrementalReadDir: 0,
		IncrementSize:      0,
		ReadOnCreate:       false,
		Filepath:           hdfFilePath,
		File:               &hdf5.File{},
	}
	dataset, err := hdf5utils.NewHdfDataset(hdfDataPath, hdfReadOptions)
	if err != nil {
		return err
	}
	flows := make([]float64, outletTS.RowCount)
	err = dataset.ReadColumn(3, flows)
	if err != nil {
		return err
	}
	err = outletTS.UpdateFloatArray(flows)
	if err != nil {
		return err
	}
	bf.BfileBlocks[outletTSIdx] = outletTS
	resultBytes, err := bf.Write()
	if err != nil {
		return err
	}
	return os.WriteFile(bfilePath, resultBytes, 0600)
}
