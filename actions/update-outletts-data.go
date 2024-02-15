package actions

import (
	"errors"
	"fmt"
	"log"
	"os"
	"ras-runner/ras"
	"reflect"
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
		return fmt.Errorf("input source %s, was not found in local directory. Run copy-local first", bfilePath)
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
	destfile, err := hdf5.OpenFile(hdfFilePath, hdf5.F_ACC_RDWR)
	if err != nil {
		return err
	}
	options := hdf5utils.HdfReadOptions{
		Dtype:        reflect.Float32,
		File:         destfile,
		ReadOnCreate: true,
	}
	destVals, err := hdf5utils.NewHdfDataset(hdfDataPath, options)
	if err != nil {
		return err
	}
	defer destVals.Close()

	flows := make([]float32, outletTS.RowCount)
	err = destVals.ReadColumn(1, &flows)
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
