package actions

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"ras-runner/fragilitycurve"
	"ras-runner/ras"

	"github.com/usace/cc-go-sdk"
)

func init() {
	cc.ActionRegistry.RegisterAction(&UpdateBfileAction2{
		ActionRunnerBase: cc.ActionRunnerBase{ActionName: "bfile-action"},
	})
}

type UpdateBfileAction2 struct {
	cc.ActionRunnerBase
	ModelDir string
}

func (uba *UpdateBfileAction2) Run() error {
	// Assumes bFile and fragility curve file  were copied local with the CopyLocal uba.Action.
	log.Printf("Ready to update bFile.")
	bFileName := uba.Action.Attributes.GetStringOrFail("bFile")
	bfilePath := fmt.Sprintf("%v/%v", uba.ModelDir, bFileName)
	if !fileExists(bfilePath) {
		log.Fatalf("Input source %s, was not found in local directory. Run copy-local first", bfilePath)
	}
	bf, err := ras.InitBFile(bfilePath)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Creating SNET ID to name map from geometry file")
	hdfFileName := uba.Action.Attributes.GetStringOrFail("geoHdfFile")
	hdfFilePath := fmt.Sprintf("%v/%v", uba.ModelDir, hdfFileName)
	err = bf.SetSNetIDToNameFromGeoHDF(hdfFilePath)
	if err != nil {
		return err
	}

	log.Print("Loading Fragility Curve Results")
	fcFileName := uba.Action.Attributes.GetStringOrFail("fcFile")
	fcFilePath := fmt.Sprintf("%v/%v", uba.ModelDir, fcFileName)
	fcFileBytes, err := os.ReadFile(fcFilePath)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	var fcResult fragilitycurve.ModelResult
	err = json.Unmarshal(fcFileBytes, &fcResult)
	if err != nil {
		log.Fatalf("Error unmarshaling fragility curve from %s", fcFileName)
		return err
	}

	for _, fclr := range fcResult.Results {
		err = bf.AmmendBreachElevations(fclr.Name, fclr.FailureElevation)
		if err != nil {
			//return err
			log.Printf("failed to ammend breach elevations: %s %f: %s\n", fclr.Name, fclr.FailureElevation, err)
		}
	}

	resultBytes, err := bf.Write()
	if err != nil {
		log.Fatal("Error writing b file")
		return err
	}

	return os.WriteFile(bfilePath, resultBytes, 0600)

}

// UpdateBfileAction reads a fragility curve output file and uses it to read and write a bfile with updated elevations.
//
// The function performs the following steps:
// 1. Retrieves the path to the bFile and checks if it exists locally.
// 2. Initializes the BFile using ras.InitBFile.
// 3. Creates an SNET ID to name map from a specified geometry file.
// 4. Loads fragility curve results from a JSON file.
// 5. Amends the breach elevations in the bFile based on the fragility curve results.
// 6. Writes the updated bFile back to disk.
//
// Parameters:
//   - action: The cc.Action object containing attributes needed for processing.
//   - modelDir: The directory path where model files are located.
//
// Returns:
//   - error: An error if any step fails, otherwise nil.
func UpdateBfileAction(action cc.Action, modelDir string) error {
	// Assumes bFile and fragility curve file  were copied local with the CopyLocal uba.Action.
	log.Printf("Ready to update bFile.")
	bFileName := action.Attributes.GetStringOrFail("bFile")
	bfilePath := fmt.Sprintf("%v/%v", modelDir, bFileName)
	if !fileExists(bfilePath) {
		log.Fatalf("Input source %s, was not found in local directory. Run copy-local first", bfilePath)
	}
	bf, err := ras.InitBFile(bfilePath)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Creating SNET ID to name map from geometry file")
	hdfFileName := action.Attributes.GetStringOrFail("geoHdfFile")
	hdfFilePath := fmt.Sprintf("%v/%v", modelDir, hdfFileName)
	err = bf.SetSNetIDToNameFromGeoHDF(hdfFilePath)
	if err != nil {
		return err
	}

	log.Print("Loading Fragility Curve Results")
	fcFileName := action.Attributes.GetStringOrFail("fcFile")
	fcFilePath := fmt.Sprintf("%v/%v", modelDir, fcFileName)
	fcFileBytes, err := os.ReadFile(fcFilePath)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	var fcResult fragilitycurve.ModelResult
	err = json.Unmarshal(fcFileBytes, &fcResult)
	if err != nil {
		log.Fatalf("Error unmarshaling fragility curve from %s", fcFileName)
		return err
	}

	for _, fclr := range fcResult.Results {
		err = bf.AmmendBreachElevations(fclr.Name, fclr.FailureElevation)
		if err != nil {
			//return err
			log.Printf("failed to ammend breach elevations: %s %f: %s\n", fclr.Name, fclr.FailureElevation, err)
		}
	}

	resultBytes, err := bf.Write()
	if err != nil {
		log.Fatal("Error writing b file")
		return err
	}

	return os.WriteFile(bfilePath, resultBytes, 0600)
}

// fileExists checks if a file exists at the specified path.
//
// Parameters:
//   - filePath: The full path to the file.
//
// Returns:
//   - bool: True if the file exists, otherwise false.
func fileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	return error != nil
}
