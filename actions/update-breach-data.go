package actions

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"ras-runner/fragilitycurve"
	"ras-runner/ras"

	"github.com/usace/cc-go-sdk"
)

// UpdateBfileAction reads a fragility curve output file and uses it to read and write a bfile with updated elevations.
func UpdateBfileAction(action cc.Action, modelDir string) error {
	// Assumes bFile and fragility curve file  were copied local with the CopyLocal action.
	log.Printf("Ready to update bFile.")
	bFileName := action.Parameters["bFile"].(string) //these may eventually need to be map[string]any instead of strings. Look at Kanawah-runner manifests as examples.
	bfilePath := fmt.Sprintf("%v/%v", modelDir, bFileName)
	if !fileExists(bfilePath) {
		log.Fatalf("Input source %s, was not found in local directory. Run copy-local first", bfilePath)
	}
	bf, err := ras.InitBFile(bfilePath)
	if err != nil {
		log.Fatal(err)
	}
	fcFileName := action.Parameters["fcFile"].(string)
	fcFilePath := fmt.Sprintf("%v/%v", modelDir, fcFileName)
	fcFileBytes, err := os.ReadFile(fcFilePath)
	if err != nil {
		log.Fatalf("Error getting input source %s", fcFileName) //why don't we use err?
		return err
	}
	var fcResult fragilitycurve.ModelResult
	err = json.Unmarshal(fcFileBytes, &fcResult)
	if err != nil {
		log.Fatalf("Error getting input source %s", fcFileName)
		return err
	}
	for _, fclr := range fcResult.Results {
		bf.AmmendBreachElevations(fclr.Name, fclr.FailureElevation)
	}
	resultBytes, err := bf.Write()
	if err != nil {
		log.Fatalf("Error getting input source %s", fcFileName)
		return err
	}
	return os.WriteFile(bfilePath, resultBytes, 0600)
}
func fileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	return !errors.Is(error, os.ErrNotExist)
}
