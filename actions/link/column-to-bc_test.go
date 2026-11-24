package actions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"ras-runner/actions"

	cc "github.com/usace-cloud-compute/cc-go-sdk"
	"github.com/usace-cloud-compute/go-hdf5"
	"github.com/usace-cloud-compute/go-hdf5/util"
)

const c2bDir = "testdata/column-to-bc"

const (
	c2bSrcFlowPath = "Results/Unsteady/Output/Output Blocks/DSS Hydrograph Output/" +
		"Unsteady Time Series/2D Flow Areas/ray-roberts/Reference Lines/Flow"
	c2bSrcTimePath = "Results/Unsteady/Output/Output Blocks/DSS Hydrograph Output/" +
		"Unsteady Time Series/Time"
	c2bDestPath = "Event Conditions/Unsteady/Boundary Conditions/Flow Hydrographs/" +
		"2D: Lewisville BCLine: ray-roberts_to_lewisville_1"
)

// copyFile copies a local file. Used to stage fixtures into temp dirs.
func copyFile(t *testing.T, src, dest string) {
	t.Helper()
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read %s: %v", src, err)
	}
	if err := os.WriteFile(dest, b, 0644); err != nil {
		t.Fatalf("write %s: %v", dest, err)
	}
}

// requireFixture skips the test if the named fixture is not present under c2bDir.
func requireFixture(t *testing.T, name string) string {
	t.Helper()
	p := filepath.Join(c2bDir, name)
	if _, err := os.Stat(p); err != nil {
		t.Skipf("fixture %s not found: %v", name, err)
	}
	return p
}

// openSrc opens the src.hdf fixture and returns the Flow (f32, 841x6) and
// Time (f64, 841) datasets plus the underlying file. Caller closes all three.
func openSrc(t *testing.T) (*util.HdfDataset, *util.HdfDataset, *hdf5.File) {
	t.Helper()
	f, err := hdf5.OpenFile(filepath.Join(c2bDir, "src.hdf"), hdf5.F_ACC_RDONLY)
	if err != nil {
		t.Fatalf("open src.hdf: %v", err)
	}
	flow, err := util.NewHdfDataset(c2bSrcFlowPath, util.HdfReadOptions{
		Dtype: reflect.Float32, File: f, ReadOnCreate: true,
	})
	if err != nil {
		f.Close()
		t.Fatalf("open Flow: %v", err)
	}
	tm, err := util.NewHdfDataset(c2bSrcTimePath, util.HdfReadOptions{
		Dtype: reflect.Float64, File: f, ReadOnCreate: true,
	})
	if err != nil {
		flow.Close()
		f.Close()
		t.Fatalf("open Time: %v", err)
	}
	return flow, tm, f
}

// readFlowColumn returns the named 1-based column of the src Flow dataset.
func readFlowColumn(t *testing.T, flow *util.HdfDataset, readcol int) []float32 {
	t.Helper()
	col := make([]float32, flow.Rows())
	for i := 0; i < flow.Rows(); i++ {
		row := make([]float32, flow.Cols())
		if err := flow.ReadRow(i, &row); err != nil {
			t.Fatalf("ReadRow(%d): %v", i, err)
		}
		col[i] = row[readcol-1]
	}
	return col
}

// readDestCols returns col0 (time) and col1 (value) of the dest dataset.
func readDestCols(t *testing.T, path string) ([]float32, []float32) {
	t.Helper()
	f, err := hdf5.OpenFile(path, hdf5.F_ACC_RDONLY)
	if err != nil {
		t.Fatalf("open dest %s: %v", path, err)
	}
	defer f.Close()
	d, err := util.NewHdfDataset(c2bDestPath, util.HdfReadOptions{
		Dtype: reflect.Float32, File: f, ReadOnCreate: true,
	})
	if err != nil {
		t.Fatalf("open dest dataset: %v", err)
	}
	defer d.Close()
	col0 := make([]float32, d.Rows())
	col1 := make([]float32, d.Rows())
	for i := 0; i < d.Rows(); i++ {
		row := make([]float32, d.Cols())
		if err := d.ReadRow(i, &row); err != nil {
			t.Fatalf("dest ReadRow(%d): %v", i, err)
		}
		col0[i], col1[i] = row[0], row[1]
	}
	return col0, col1
}

func near(a, b float64, relTol float64) bool {
	if a == b {
		return true
	}
	if a == 0 {
		return b == 0
	}
	d := a - b
	if d < 0 {
		d = -d
	}
	return d/(a) < relTol
}

// TestMigrateColumnData_TimeMatched is the primary coverage: the migrated dest
// col1 must equal the src Flow column 5 (index 4) for all rows, col0 must be
// preserved, and the result must differ from both the original dest col1 and
// from Flow column 6 (index 5) to guard against an off-by-one.
func TestMigrateColumnData_TimeMatched(t *testing.T) {
	src := requireFixture(t, "src.hdf")
	absSrc, err := filepath.Abs(src)
	if err != nil {
		t.Fatalf("abs src: %v", err)
	}
	modelDir := t.TempDir()
	copyFile(t, filepath.Join(c2bDir, "dest.hdf"), filepath.Join(modelDir, "dest.hdf"))
	// MODEL_DIR is a package-level var initialized once at import; t.Setenv
	// after import has no effect. Override the var directly and restore on exit.
	origModelDir := actions.MODEL_DIR
	actions.MODEL_DIR = modelDir
	t.Cleanup(func() { actions.MODEL_DIR = origModelDir })

	// Golden: read Flow[i][4] (column 5, 1-based) for all rows.
	flow, tm, f := openSrc(t)
	defer flow.Close()
	defer tm.Close()
	defer f.Close()
	golden := readFlowColumn(t, flow, 5)
	wrongCol := readFlowColumn(t, flow, 6) // column 6 — off-by-one guard

	// Snapshot the pristine dest before migration.
	before0, before1 := readDestCols(t, filepath.Join(modelDir, "dest.hdf"))

	if err := MigrateColumnData(absSrc, &cc.DataStore{StoreType: "FS"},
		c2bSrcFlowPath, "dest.hdf", c2bDestPath, 5, 0.00001); err != nil {
		t.Fatalf("MigrateColumnData: %v", err)
	}

	after0, after1 := readDestCols(t, filepath.Join(modelDir, "dest.hdf"))
	if len(after0) != len(before0) || len(after1) != len(before1) {
		t.Fatalf("row count changed: before=%d after=%d", len(before0), len(after1))
	}
	for i := range after0 {
		// col0 (time) must be preserved exactly.
		if after0[i] != before0[i] {
			t.Fatalf("row %d: col0 changed from %v to %v", i, before0[i], after0[i])
		}
		// col1 must equal the golden Flow[i][4] within relative 1e-3.
		if !near(float64(after1[i]), float64(golden[i]), 1e-3) {
			t.Fatalf("row %d: col1=%v want=%v", i, after1[i], golden[i])
		}
	}

	// The migration must have visibly changed the series (middle of the range).
	mid := len(after1) / 2
	if after1[mid] == before1[mid] {
		t.Errorf("row %d: col1 unchanged after migration (=%v); migration may be a no-op", mid, after1[mid])
	}

	// Off-by-one guard: col1 must not equal Flow[i][5] (column 6) for at least
	// one row.
	mismatched := 0
	for i := range after1 {
		if !near(float64(after1[i]), float64(wrongCol[i]), 1e-3) {
			mismatched++
		}
	}
	if mismatched == 0 {
		t.Errorf("col1 matches Flow column 6 (index 5) for every row; wrong column was read")
	}
}

// TestMigrateColumnData_NoTimeMatch drives the no-match error path with a
// synthetic src/dest whose times do not overlap.
func TestMigrateColumnData_NoTimeMatch(t *testing.T) {
	dir := t.TempDir()
	srcPath := buildTinySrc(t, dir, []float64{1.0, 3.0, 5.0}) // times 1,3,5
	destPath := buildTinyDest(t, dir, []float64{2.5})         // 2.5 matches none
	origModelDir := actions.MODEL_DIR
	actions.MODEL_DIR = dir
	t.Cleanup(func() { actions.MODEL_DIR = origModelDir })

	err := MigrateColumnData(srcPath, &cc.DataStore{StoreType: "FS"},
		c2bSrcFlowPath, filepath.Base(destPath), c2bDestPath, 2, 0.001)
	if err == nil {
		t.Fatal("expected error for unmatched time, got nil")
	}
	want := "unable to find corresponding input source record for time 2.500000"
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}
}

// TestMigrateColumnData_ZeroTimeShortcut pins the timeval<=0 shortcut: a dest
// whose row-0 time is 0.0 must take src row 0 regardless of the source times.
func TestMigrateColumnData_ZeroTimeShortcut(t *testing.T) {
	dir := t.TempDir()
	srcPath := buildTinySrc(t, dir, []float64{5.0, 6.0}) // Time[0] != 0
	destPath := buildTinyDest(t, dir, []float64{0.0})    // row 0 time = 0.0
	origModelDir := actions.MODEL_DIR
	actions.MODEL_DIR = dir
	t.Cleanup(func() { actions.MODEL_DIR = origModelDir })

	err := MigrateColumnData(srcPath, &cc.DataStore{StoreType: "FS"},
		c2bSrcFlowPath, filepath.Base(destPath), c2bDestPath, 2, 0.001)
	if err != nil {
		t.Fatalf("MigrateColumnData: %v", err)
	}
	_, col1 := readDestCols(t, destPath)
	// buildTinySrc writes Flow row 0 as [100, 200]. The timeval<=0 shortcut
	// returns srcdata[0] (column 0) unconditionally — it ignores readcol.
	if col1[0] != 100 {
		t.Fatalf("row 0 col1 = %v, want 100 (src row 0 col 0, no time matching)", col1[0])
	}
}

// TestGetRowVal2_TimeMatch: a timeval equal to a source Time[i] returns
// Flow[i][readcol-1].
func TestGetRowVal2_TimeMatch(t *testing.T) {
	flow, tm, f := openSrc(t)
	defer flow.Close()
	defer tm.Close()
	defer f.Close()

	i := 10
	times := make([]float64, tm.Rows())
	for r := 0; r < tm.Rows(); r++ {
		row := make([]float64, 1)
		if err := tm.ReadRow(r, &row); err != nil {
			t.Fatalf("Time ReadRow(%d): %v", r, err)
		}
		times[r] = row[0]
	}
	want := readFlowColumn(t, flow, 5)[i]

	got, err := getRowVal2(flow, tm, float32(times[i]), 5, 1e-5)
	if err != nil {
		t.Fatalf("getRowVal2: %v", err)
	}
	if !near(float64(got), float64(want), 1e-6) {
		t.Fatalf("getRowVal2 = %v, want %v", got, want)
	}
}

// TestGetRowVal2_ToleranceBoundary: a timeval within the tolerance of a source
// time matches; one just outside it does not.
func TestGetRowVal2_ToleranceBoundary(t *testing.T) {
	flow, tm, f := openSrc(t)
	defer flow.Close()
	defer tm.Close()
	defer f.Close()

	i := 50
	times := make([]float64, tm.Rows())
	for r := 0; r < tm.Rows(); r++ {
		row := make([]float64, 1)
		if err := tm.ReadRow(r, &row); err != nil {
			t.Fatalf("Time ReadRow(%d): %v", r, err)
		}
		times[r] = row[0]
	}
	want := readFlowColumn(t, flow, 5)[i]
	tol := 0.01

	// Within tolerance (strictly less than).
	got, err := getRowVal2(flow, tm, float32(times[i]+tol/2), 5, tol)
	if err != nil {
		t.Fatalf("within-tolerance: %v", err)
	}
	if !near(float64(got), float64(want), 1e-6) {
		t.Fatalf("within-tolerance: got %v, want %v", got, want)
	}

	// Just outside tolerance -> no match.
	if _, err := getRowVal2(flow, tm, float32(times[i]+tol*2), 5, tol); err == nil {
		t.Fatal("outside-tolerance: expected no-match error, got nil")
	}
}

// TestGetRowVal2_ZeroTimeShortcut: timeval <= 0 returns src row 0 col 0 with no
// time comparison, regardless of Time[0]. (The shortcut hardcodes srcdata[0]
// and ignores readcol — see lines 247–253 of column-to-bc.go.)
func TestGetRowVal2_ZeroTimeShortcut(t *testing.T) {
	dir := t.TempDir()
	srcPath := buildTinySrc(t, dir, []float64{5.0, 6.0})
	f, err := hdf5.OpenFile(srcPath, hdf5.F_ACC_RDONLY)
	if err != nil {
		t.Fatalf("open synthetic src: %v", err)
	}
	defer f.Close()
	flow, err := util.NewHdfDataset(c2bSrcFlowPath, util.HdfReadOptions{
		Dtype: reflect.Float32, File: f, ReadOnCreate: true,
	})
	if err != nil {
		t.Fatalf("open Flow: %v", err)
	}
	defer flow.Close()
	tm, err := util.NewHdfDataset(c2bSrcTimePath, util.HdfReadOptions{
		Dtype: reflect.Float64, File: f, ReadOnCreate: true,
	})
	if err != nil {
		t.Fatalf("open Time: %v", err)
	}
	defer tm.Close()

	// Time[0]=5.0, so a non-shortcut call with timeval=0 would fail; the
	// shortcut must still return row 0. The shortcut returns srcdata[0]
	// (column 0) unconditionally — it ignores readcol.
	got, err := getRowVal2(flow, tm, 0.0, 2, 0.001)
	if err != nil {
		t.Fatalf("getRowVal2(0.0): %v", err)
	}
	if got != 100 { // buildTinySrc Flow row 0 = [100, 200]; shortcut returns col 0
		t.Fatalf("getRowVal2(0.0) = %v, want 100", got)
	}
	// Negative timeval also takes the shortcut (and likewise returns col 0,
	// not readcol-1).
	got, err = getRowVal2(flow, tm, -1.0, 2, 0.001)
	if err != nil {
		t.Fatalf("getRowVal2(-1.0): %v", err)
	}
	if got != 100 { // readcol=2 but shortcut returns col 0
		t.Fatalf("getRowVal2(-1.0) = %v, want 100", got)
	}
}

// TestGetRowVal2_NoMatch: a timeval matching no source time returns the
// documented error.
func TestGetRowVal2_NoMatch(t *testing.T) {
	flow, tm, f := openSrc(t)
	defer flow.Close()
	defer tm.Close()
	defer f.Close()

	// Pick a positive time guaranteed not to be in the source series. (A
	// negative timeval would hit the timeval<=0 shortcut and return row 0.)
	_, err := getRowVal2(flow, tm, 99999, 5, 1e-5)
	if err == nil {
		t.Fatal("expected no-match error, got nil")
	}
	want := "unable to find corresponding input source record for time 99999.000000"
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}
}

// TestGetRowVal2_ColumnSelection: readcol is 1-based; readcol=1 returns col 0,
// readcol=6 returns col 5 (off-by-one guard).
func TestGetRowVal2_ColumnSelection(t *testing.T) {
	flow, tm, f := openSrc(t)
	defer flow.Close()
	defer tm.Close()
	defer f.Close()

	i := 200
	times := make([]float64, tm.Rows())
	for r := 0; r < tm.Rows(); r++ {
		row := make([]float64, 1)
		if err := tm.ReadRow(r, &row); err != nil {
			t.Fatalf("Time ReadRow(%d): %v", r, err)
		}
		times[r] = row[0]
	}
	wantCol1 := readFlowColumn(t, flow, 1)[i]
	wantCol6 := readFlowColumn(t, flow, 6)[i]

	got1, err := getRowVal2(flow, tm, float32(times[i]), 1, 1e-5)
	if err != nil {
		t.Fatalf("readcol=1: %v", err)
	}
	if !near(float64(got1), float64(wantCol1), 1e-6) {
		t.Fatalf("readcol=1: got %v, want %v", got1, wantCol1)
	}

	got6, err := getRowVal2(flow, tm, float32(times[i]), 6, 1e-5)
	if err != nil {
		t.Fatalf("readcol=6: %v", err)
	}
	if !near(float64(got6), float64(wantCol6), 1e-6) {
		t.Fatalf("readcol=6: got %v, want %v", got6, wantCol6)
	}

	// Guard: the two columns must actually differ, or the test is tautological.
	if wantCol1 == wantCol6 {
		t.Errorf("Flow row %d col0 == col5 (%v); column-selection assertion is tautological", i, wantCol1)
	}
}

// makeGroups creates each directory component of the given dataset path as a
// nested HDF5 group, one level at a time.
func makeGroups(t *testing.T, f *hdf5.File, datapath string) {
	t.Helper()
	parts := strings.Split(datapath, "/")
	cur := ""
	for i := 0; i < len(parts)-1; i++ {
		cur = cur + "/" + parts[i]
		if _, err := f.CreateGroup(cur); err != nil {
			t.Fatalf("CreateGroup(%s): %v", cur, err)
		}
	}
}

// buildTinySrc writes a minimal source HDF5 with the same dataset/group layout
// as the real fixture: a 2x2 float32 Flow ([100,200; 300,400]) and a float64
// Time of the given values. The layout lets the same datapath constants be used.
func buildTinySrc(t *testing.T, dir string, times []float64) string {
	t.Helper()
	path := filepath.Join(dir, "tiny-src.hdf")
	f, err := hdf5.CreateFile(path, hdf5.F_ACC_TRUNC)
	if err != nil {
		t.Fatalf("CreateFile: %v", err)
	}
	defer f.Close()
	makeGroups(t, f, c2bSrcFlowPath)
	flowspace, err := hdf5.CreateSimpleDataspace([]uint{2, 2}, []uint{2, 2})
	if err != nil {
		t.Fatalf("CreateSimpleDataspace: %v", err)
	}
	flowds, err := f.CreateDataset(c2bSrcFlowPath, hdf5.T_NATIVE_FLOAT, flowspace)
	if err != nil {
		t.Fatalf("CreateDataset(Flow): %v", err)
	}
	flow := []float32{100, 200, 300, 400}
	if err := flowds.Write(&flow); err != nil {
		t.Fatalf("Write(Flow): %v", err)
	}
	flowds.Close()
	flowspace.Close()

	timespace, err := hdf5.CreateSimpleDataspace([]uint{uint(len(times))}, []uint{uint(len(times))})
	if err != nil {
		t.Fatalf("CreateSimpleDataspace(Time): %v", err)
	}
	tm, err := f.CreateDataset(c2bSrcTimePath, hdf5.T_NATIVE_DOUBLE, timespace)
	if err != nil {
		t.Fatalf("CreateDataset(Time): %v", err)
	}
	if err := tm.Write(&times); err != nil {
		t.Fatalf("Write(Time): %v", err)
	}
	tm.Close()
	timespace.Close()
	return path
}

// buildTinyDest writes a minimal destination HDF5 with the same dataset/group
// layout as the real fixture: a (n x 2) float32 dataset where col0 is the given
// times and col1 is a pre-existing BC value (999) that differs from any source
// value.
func buildTinyDest(t *testing.T, dir string, times []float64) string {
	t.Helper()
	path := filepath.Join(dir, "tiny-dest.hdf")
	f, err := hdf5.CreateFile(path, hdf5.F_ACC_TRUNC)
	if err != nil {
		t.Fatalf("CreateFile: %v", err)
	}
	defer f.Close()
	makeGroups(t, f, c2bDestPath)
	dspace, err := hdf5.CreateSimpleDataspace([]uint{uint(len(times)), 2}, []uint{uint(len(times)), 2})
	if err != nil {
		t.Fatalf("CreateSimpleDataspace: %v", err)
	}
	ds, err := f.CreateDataset(c2bDestPath, hdf5.T_NATIVE_FLOAT, dspace)
	if err != nil {
		t.Fatalf("CreateDataset: %v", err)
	}
	data := make([]float32, len(times)*2)
	for i, tv := range times {
		data[i*2] = float32(tv)
		data[i*2+1] = 999
	}
	if err := ds.Write(&data); err != nil {
		t.Fatalf("Write: %v", err)
	}
	ds.Close()
	dspace.Close()
	return path
}

// loadC2bPayload decodes the column-to-bc payload fixture from disk. No CcStore,
// no InitPluginManager, no CC_* env — the payload is a plain file.
func loadC2bPayload(t *testing.T) cc.Payload {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(c2bDir, "payload-column-to-bc"))
	if err != nil {
		t.Fatalf("read payload: %v", err)
	}
	var payload cc.Payload
	if err := json.Unmarshal(b, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	return payload
}

// TestColumnToBcRun_Adapter exercises the thin Run() adapter: it wires the
// cc-go-sdk PluginManager from the on-disk payload, runs the action, and
// asserts Run() returns nil and the dest file was modified. It deliberately
// does NOT assert on migrated values — that is MigrateColumnData's job, already
// covered by the direct tests above.
func TestColumnToBcRun_Adapter(t *testing.T) {
	payload := loadC2bPayload(t)

	// Stage the fixtures: src.hdf at the store root (the input data source path),
	// dest.hdf under MODEL_DIR (where MigrateColumnData resolves the dest).
	storeRoot := t.TempDir()
	copyFile(t, filepath.Join(c2bDir, "src.hdf"), filepath.Join(storeRoot, "src.hdf"))
	modelDir := t.TempDir()
	copyFile(t, filepath.Join(c2bDir, "dest.hdf"), filepath.Join(modelDir, "dest.hdf"))
	// MODEL_DIR is a package-level var initialized once at import; t.Setenv
	// after import has no effect. Override the var directly and restore on exit.
	origModelDir := actions.MODEL_DIR
	actions.MODEL_DIR = modelDir
	t.Cleanup(func() { actions.MODEL_DIR = origModelDir })

	// Register the FS store type (normally done by InitPluginManager, which we
	// avoid) and connect the payload's own stores the way production does.
	cc.RegisterStoreTypes()
	if err := cc.ConnectStores(&payload.Stores); err != nil {
		t.Fatalf("ConnectStores: %v", err)
	}

	// MigrateColumnData reads Paths["hdf"] raw (no store-root prepending), so the
	// input path must be absolute; otherwise util.OpenFile resolves the relative
	// "src.hdf" against the test CWD and won't find it.
	payload.Inputs[0].Paths["hdf"] = filepath.Join(storeRoot, "src.hdf")

	pm := &cc.PluginManager{Payload: payload}
	pm.IOManager.SetParent(&pm.IOManager)
	for i := range pm.Actions {
		pm.Actions[i].IOManager.SetParent(&pm.IOManager)
	}

	runner := ColumnToBcAction{
		ActionRunnerBase: cc.ActionRunnerBase{
			ActionName:    "column-to-boundary-condition",
			Action:        pm.Actions[0],
			PluginManager: pm,
		},
	}

	destPath := filepath.Join(modelDir, "dest.hdf")
	before, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("stat dest before: %v", err)
	}

	if err := runner.Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}

	after, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("stat dest after: %v", err)
	}
	if before.ModTime().Equal(after.ModTime()) && before.Size() == after.Size() {
		t.Error("dest file was not modified by Run()")
	}
}

// TestColumnToBcRun_AttributeValidation covers the Run() attribute-parsing error
// paths, which need no HDF5 data.
func TestColumnToBcRun_AttributeValidation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(a *cc.Action)
		want   string
	}{
		{
			name: "invalid column index",
			mutate: func(a *cc.Action) {
				a.Attributes["column_index"] = "abc"
			},
			want: "invalid column index",
		},
		{
			name: "missing src",
			mutate: func(a *cc.Action) {
				delete(a.Attributes, "src")
			},
			want: "missing src attribute data",
		},
		{
			name: "missing dest",
			mutate: func(a *cc.Action) {
				delete(a.Attributes, "dest")
			},
			want: "missing dest attribute data",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload := loadC2bPayload(t)
			action := payload.Actions[0]
			tc.mutate(&action)
			payload.Actions[0] = action

			cc.RegisterStoreTypes()
			if err := cc.ConnectStores(&payload.Stores); err != nil {
				t.Fatalf("ConnectStores: %v", err)
			}

			pm := &cc.PluginManager{Payload: payload}
			pm.IOManager.SetParent(&pm.IOManager)
			for i := range pm.Actions {
				pm.Actions[i].IOManager.SetParent(&pm.IOManager)
			}

			runner := ColumnToBcAction{
				ActionRunnerBase: cc.ActionRunnerBase{
					ActionName:    "column-to-boundary-condition",
					Action:        pm.Actions[0],
					PluginManager: pm,
				},
			}
			err := runner.Run()
			if err == nil {
				t.Fatalf("expected error %q, got nil", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %q, want it to contain %q", err.Error(), tc.want)
			}
		})
	}
}
