package actions

import (
	"net/url"
	"os"
	"strings"
	//_ "actions/utils"
)

const (
	MODEL_DIR           = "/sim/model"
	MODEL_SCRIPT        = "run-model.sh"
	MODEL_SCRIPT_PATH   = "/ras"
	GEOM_PREPROC        = "run-geom-preproc.sh"
	RASTIMEPATH         = "Unsteady Time Series/Time"
	AWSBUCKET           = "AWS_S3_BUCKET"
	RAS_SCRIPT_PATH_ENV = "RAS_SCRIPT_PATH"
)

// this is the tolerance we will use when comparing float64 values for comparison
// specifically it is used to compare RAS time values
const Tolerance float64 = 0.000001

const S3BucketTemplate = "https://%s.s3.amazonaws.com%s/%s"

// fileExists checks if a file exists at the specified path.
//
// Parameters:
//   - filePath: The full path to the file.
//
// Returns:
//   - bool: True if the file exists, otherwise false.
func FileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	return error == nil
}

func EncodeUrlPath(src string) string {
	srcvals := strings.Split(src, "/")
	srcencoded := strings.Builder{}
	for i, sv := range srcvals {
		if i == 0 {
			srcencoded.WriteString(url.PathEscape(sv))
		} else {
			srcencoded.WriteString("/" + url.PathEscape(sv))
		}
	}
	return srcencoded.String()
}

func TimePath(datapath string) string {
	tsroot := datapath[:strings.Index(datapath, "Unsteady Time Series")]
	return tsroot + RASTIMEPATH
}
