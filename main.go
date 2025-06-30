package main

// #include <stdlib.h>
import "C"
import (
	"log"
	_ "ras-runner/actions/extract/hdf"
	_ "ras-runner/actions/link"
	_ "ras-runner/actions/run"
	_ "ras-runner/actions/utils"

	"github.com/usace/cc-go-sdk"
)

//@TODO fix action description logging back to action.Name

var modelPrefix string
var event int

func main() {
	pm, err := cc.InitPluginManager()
	if err != nil {
		log.Fatalf("unable to initialize the CC plugin manager: %s\n", err)
	}
	pm.RunActions()
	log.Println("Finished")
}

// func fetchInputSourceFiles(pm *cc.PluginManager) error {
// 	for _, ds := range pm.Inputs {
// 		err := func() error {
// 			source, err := pm.GetReader(cc.DataSourceOpInput{
// 				DataSourceName: ds.Name,
// 				PathKey:        "0",
// 			})
// 			if err != nil {
// 				return err
// 			}
// 			defer source.Close()
// 			destfile := fmt.Sprintf("%s/%s", actions.MODEL_DIR, filepath.Base(ds.Paths["0"]))
// 			log.Printf("Copying %s to %s\n", ds.Paths["0"], destfile)
// 			destination, err := os.Create(destfile)
// 			if err != nil {
// 				return err
// 			}
// 			defer destination.Close()
// 			_, err = io.Copy(destination, source)
// 			return err
// 		}()
// 		if err != nil {
// 			log.Printf("Error fetching %s", ds.Paths["0"])
// 			return err
// 		}
// 	}
// 	return nil
// }

// func writeFile(fExt string, obj io.ReadCloser) error {
// 	filepath := fmt.Sprintf("%s/%s.%s", actions.MODEL_DIR, modelPrefix, fExt)
// 	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
// 	if err != nil {
// 		return err
// 	}
// 	defer f.Close()
// 	_, err = io.Copy(f, obj)
// 	return err
// }
