//go:build fixture

package hdf

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"ras-runner/internal/testutil"
	"testing"

	"github.com/usace-cloud-compute/cc-go-sdk"
	filestore "github.com/usace-cloud-compute/filesapi"
)

// copyFile copies a local file. Used to stage fixtures into temp dirs.
func copyFile(src, dest string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, in, 0644)
}

// TestRasExtractActionRun exercises RasExtractAction.Run end-to-end: the
// payload is decoded from the fixture, the action runs against a staged
// results file, and the extracted JSON is written to the output data source.
func TestRasExtractActionRun(t *testing.T) {
	resetAccumulator(t)
	testutil.SkipIfNoFile(t, duwamishHdfFixture)
	testutil.SkipIfNoFile(t, "actions/extract/ras-extract/payload")

	// Stage the results file where the action expects it: MODEL_DIR/<prefix>.p<plan>.tmp.hdf
	modelDir := t.TempDir()
	destName := "Duwamish_17110013.p01.tmp.hdf"
	if err := copyFile(testutil.FixturePath(t, duwamishHdfFixture), filepath.Join(modelDir, destName)); err != nil {
		t.Fatalf("failed to stage results file: %v", err)
	}
	t.Setenv("MODEL_DIR", modelDir)

	// Decode the payload and point its FS store at a temp dir.
	payloadPath := testutil.FixturePath(t, "actions/extract/ras-extract/payload")
	payloadBytes, err := os.ReadFile(payloadPath)
	if err != nil {
		t.Fatalf("failed to read payload: %v", err)
	}
	var payload cc.Payload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}
	storeRoot := t.TempDir()
	for i := range payload.Stores {
		if payload.Stores[i].StoreType == "FS" {
			payload.Stores[i].Parameters["root"] = storeRoot
		}
	}

	// Build the plugin manager without going through InitPluginManager (which
	// requires CC env vars and a backing store). We only need the IOManager
	// wiring and the FS store session for Put to work.
	pm := &cc.PluginManager{Payload: payload}
	if err := pm.IOManager.SetParent(&pm.IOManager); err != nil {
		t.Fatalf("failed to set parent: %v", err)
	}
	for i := range pm.Actions {
		pm.Actions[i].IOManager.SetParent(&pm.IOManager)
	}
	// Connect the FS store so Put can write.
	fs, err := filestore.NewFileStore(filestore.BlockFSConfig{})
	if err != nil {
		t.Fatalf("failed to create FS store: %v", err)
	}
	for i := range pm.Stores {
		if pm.Stores[i].StoreType == "FS" {
			if root, ok := pm.Stores[i].Parameters["root"]; ok {
				pm.Stores[i].Session = &filestore.BlockFS{filestore.BlockFSConfig{}}
				_ = fs
				_ = root
			}
		}
	}

	runner := RasExtractAction{
		ActionRunnerBase: cc.ActionRunnerBase{
			ActionName:    "ras-extract",
			Action:        pm.Actions[0],
			PluginManager: pm,
		},
	}

	if err := runner.Run(); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
}
