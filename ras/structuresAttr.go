package ras

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/usace-cloud-compute/go-hdf5"
	"github.com/usace-cloud-compute/go-hdf5/util"
)

type Hdf5Float32 float32

func (value Hdf5Float32) MarshalJSON() ([]byte, error) {
	f := float64(value)
	if math.IsNaN(f) {
		return []byte("null"), nil
	}
	return json.Marshal(f)
}

const STRUCTURE_DATA_PATH string = "Geometry/Structures/Attributes/"

// nominated for an ig-nobel award in data structure science.......
type structuresAttr struct {
	Type                     string      `strlen:"16" hdf:"Type"`
	Mode                     string      `strlen:"18" hdf:"Mode"`
	River                    string      `strlen:"16" hdf:"River"`
	Reach                    string      `strlen:"16" hdf:"Reach"`
	RS                       string      `strlen:"8" hdf:"RS"`
	Connection               string      `strlen:"16" hdf:"Connection"`
	GroupName                string      `strlen:"45" hdf:"Groupname"`
	UsType                   string      `strlen:"16" hdf:"US Type"`
	UsRiver                  string      `strlen:"16" hdf:"US River"`
	UsReach                  string      `strlen:"16" hdf:"US Reach"`
	UsRs                     string      `strlen:"8" hdf:"US RS"`
	UsSa2d                   string      `strlen:"16" hdf:"US SA/2D"`
	DsType                   string      `strlen:"16" hdf:"DS Type"`
	DsRiver                  string      `strlen:"16" hdf:"DS River"`
	DsReach                  string      `strlen:"16" hdf:"DS Reach"`
	DsRs                     string      `strlen:"8" hdf:"DS RS"`
	DsSa2d                   string      `strlen:"16" hdf:"DS SA/2D"`
	NodeName                 string      `strlen:"16" hdf:"Node Name"`
	Description              string      `strlen:"512" hdf:"Description"`
	LastEdited               string      `strlen:"18" hdf:"Last Edited"`
	UpstreamDistance         Hdf5Float32 `hdf:"Upstream Distance"`
	WeirWidth                Hdf5Float32 `hdf:"Weir Width"`
	WeirMaxSubmergence       Hdf5Float32 `hdf:"Weir Max Submergence"`
	WeirMinElev              Hdf5Float32 `hdf:"Weir Min Elevation"`
	WeirCoef                 Hdf5Float32 `hdf:"Weir Coef"`
	WeirShape                string      `strlen:"16" hdf:"Weir Shape"`
	WeirDesignEgHead         Hdf5Float32 `hdf:"Weir Design EG Head"`
	WeirDesignSpillHt        Hdf5Float32 `hdf:"Weir Design Spillway HT"`
	WeirUsSlope              Hdf5Float32 `hdf:"Weir US Slope"`
	WeirDsSlope              Hdf5Float32 `hdf:"Weir DS Slope"`
	LinearRoutingPosCoef     Hdf5Float32 `hdf:"Linear Routing Positive Coef"`
	LinearRoutingNegCoef     Hdf5Float32 `hdf:"Linear Routing Negative Coef"`
	LinearRoutingElev        Hdf5Float32 `hdf:"Linear Routing Elevation"`
	LwHwPosition             int32       `hdf:"LW HW Position"`
	LwTwPosition             int32       `hdf:"LW TW Position"`
	LwHwDistance             Hdf5Float32 `hdf:"LW HW Distance"`
	LwTwDistance             Hdf5Float32 `hdf:"LW TW Distance"`
	LwSpanMultiple           uint8       `hdf:"LW Span Multiple"`
	Use2DForOverflow         uint8       `hdf:"Use 2D for Overflow"`
	UseVelocityInto2D        uint8       `hdf:"Use Velocity into 2D"`
	HagarsWeirCoef           Hdf5Float32 `hdf:"Hagers Weir Coef"`
	HagarsHeight             Hdf5Float32 `hdf:"Hagers Height"`
	HagarsSlope              Hdf5Float32 `hdf:"Hagers Slope"`
	HagarsAngle              Hdf5Float32 `hdf:"Hagers Angle"`
	HagarsRadius             Hdf5Float32 `hdf:"Hagers Radius"`
	UseWsForWeirRef          uint8       `hdf:"Use WS for Weir Reference"`
	PilotFlow                Hdf5Float32 `hdf:"Pilot Flow"`
	CulvertGroups            int32       `hdf:"Culvert Groups"`
	CulvertsFlapGates        int32       `hdf:"Culverts Flap Gates"`
	GateGroups               int32       `hdf:"Gate Groups"`
	HtabFfPoints             int32       `hdf:"HTAB FF Points"`
	HtabRcCounts             int32       `hdf:"HTAB RC Count"`
	HtabRcPoints             int32       `hdf:"HTAB RC Points"`
	HtabHwMax                Hdf5Float32 `hdf:"HTAB HW Max"`
	HtabTwMax                Hdf5Float32 `hdf:"HTAB TW Max"`
	HtabMaxFlow              Hdf5Float32 `hdf:"HTAB Max Flow"`
	CellSpacingNear          float32     `hdf:"Cell Spacing Near"`
	CellSpacingFar           float32     `hdf:"Cell Spacing Far"`
	NearRepeats              int32       `hdf:"Near Repeats"`
	ProtectionRadius         uint8       `hdf:"Protection Radius"`
	UseFrictionInMomentum    uint8       `hdf:"Use Friction in Momentum"`
	UseWeightInMomentum      uint8       `hdf:"Use Weight in Momentum"`
	UseCriticalUs            uint8       `hdf:"Use Critical US"`
	UseEgforPressureCriteria uint8       `hdf:"Use EG for Pressure Criteria"`
	IceOption                int32       `hdf:"Ice Option"`
	WeirSkew                 Hdf5Float32 `hdf:"Weir Skew"`
	PierSkew                 Hdf5Float32 `hdf:"Pier Skew"`
	BrContraction            Hdf5Float32 `hdf:"BR Contraction"`
	BrExpansion              Hdf5Float32 `hdf:"BR Expansion"`
	BrUsLeftBank             Hdf5Float32 `hdf:"BR US Left Bank"`
	BrUsRightBank            Hdf5Float32 `hdf:"BR US Right Bank"`
	BrDsLeftBank             Hdf5Float32 `hdf:"BR DS Left Bank"`
	BrDsrightBank            Hdf5Float32 `hdf:"BR DS Right Bank"`
	XsUsLeftBank             Hdf5Float32 `hdf:"XS US Left Bank"`
	XsUsRightBank            Hdf5Float32 `hdf:"XS US Right Bank"`
	XsDsLeftBank             Hdf5Float32 `hdf:"XS DS Left Bank"`
	XsDsRightBank            Hdf5Float32 `hdf:"XS DS Right Bank"`
	UsIneffLeftSta           Hdf5Float32 `hdf:"US Ineff Left Sta"`
	UsIneffLeftElev          Hdf5Float32 `hdf:"US Ineff Left Elev"`
	UsIneffRightSta          Hdf5Float32 `hdf:"US Ineff Right Sta"`
	UsIneffRightElev         Hdf5Float32 `hdf:"US Ineff Right Elev"`
	DsIneffLeftSta           Hdf5Float32 `hdf:"DS Ineff Left Sta"`
	DsIneffLeftElev          Hdf5Float32 `hdf:"DS Ineff Left Elev"`
	DsIneffRightSta          Hdf5Float32 `hdf:"DS Ineff Right Sta"`
	DsIneffRightElev         Hdf5Float32 `hdf:"DS Ineff Right Elev"`
	UseOverrideHwCon         uint8       `hdf:"Use Override HW Connectivity"`
	UseOverrideTwCon         uint8       `hdf:"Use Override TW Connectivity"`
	UseRCFamily              uint8       `hdf:"Use RC Family"` //This wasn't present in Randy's example, But is in Bald Eagle from HEC-RAS 6.5
	UseOverideHTabIBCurves   uint8       `hdf:"Use Override HTabIBCurves"`
	SnnID                    int32       `hdf:"SNN ID"`
	DefaultCenterline        uint8       `hdf:"Default Centerline"`
	Culverts                 int32       `hdf:"Culverts"` //think starting here that these are old depricated attributes from v5
	Gates                    int32       `hdf:"Gates"`
	LwVelocityInto2D         uint8       `hdf:"LW Velocity Into 2D"`
}
type structuresAttr66 struct {
	Type                     string      `strlen:"16" hdf:"Type"`
	Mode                     string      `strlen:"18" hdf:"Mode"`
	River                    string      `strlen:"16" hdf:"River"`
	Reach                    string      `strlen:"16" hdf:"Reach"`
	RS                       string      `strlen:"8" hdf:"RS"`
	Connection               string      `strlen:"16" hdf:"Connection"`
	GroupName                string      `strlen:"45" hdf:"Groupname"`
	UsType                   string      `strlen:"16" hdf:"US Type"`
	UsRiver                  string      `strlen:"16" hdf:"US River"`
	UsReach                  string      `strlen:"16" hdf:"US Reach"`
	UsRs                     string      `strlen:"8" hdf:"US RS"`
	UsSa2d                   string      `strlen:"16" hdf:"US SA/2D"`
	DsType                   string      `strlen:"16" hdf:"DS Type"`
	DsRiver                  string      `strlen:"16" hdf:"DS River"`
	DsReach                  string      `strlen:"16" hdf:"DS Reach"`
	DsRs                     string      `strlen:"8" hdf:"DS RS"`
	DsSa2d                   string      `strlen:"16" hdf:"DS SA/2D"`
	NodeName                 string      `strlen:"16" hdf:"Node Name"`
	Description              string      `strlen:"512" hdf:"Description"`
	LastEdited               string      `strlen:"18" hdf:"Last Edited"`
	UpstreamDistance         Hdf5Float32 `hdf:"Upstream Distance"`
	WeirWidth                Hdf5Float32 `hdf:"Weir Width"`
	WeirMaxSubmergence       Hdf5Float32 `hdf:"Weir Max Submergence"`
	WeirMinElev              Hdf5Float32 `hdf:"Weir Min Elevation"`
	WeirCoef                 Hdf5Float32 `hdf:"Weir Coef"`
	WeirShape                string      `strlen:"16" hdf:"Weir Shape"`
	WeirDesignEgHead         Hdf5Float32 `hdf:"Weir Design EG Head"`
	WeirDesignSpillHt        Hdf5Float32 `hdf:"Weir Design Spillway HT"`
	WeirUsSlope              Hdf5Float32 `hdf:"Weir US Slope"`
	WeirDsSlope              Hdf5Float32 `hdf:"Weir DS Slope"`
	LinearRoutingPosCoef     Hdf5Float32 `hdf:"Linear Routing Positive Coef"`
	LinearRoutingNegCoef     Hdf5Float32 `hdf:"Linear Routing Negative Coef"`
	LinearRoutingElev        Hdf5Float32 `hdf:"Linear Routing Elevation"`
	LwHwPosition             int32       `hdf:"LW HW Position"`
	LwTwPosition             int32       `hdf:"LW TW Position"`
	LwHwDistance             Hdf5Float32 `hdf:"LW HW Distance"`
	LwTwDistance             Hdf5Float32 `hdf:"LW TW Distance"`
	LwSpanMultiple           uint8       `hdf:"LW Span Multiple"`
	Use2DForOverflow         uint8       `hdf:"Use 2D for Overflow"`
	UseVelocityInto2D        uint8       `hdf:"Use Velocity into 2D"`
	HagarsWeirCoef           Hdf5Float32 `hdf:"Hagers Weir Coef"`
	HagarsHeight             Hdf5Float32 `hdf:"Hagers Height"`
	HagarsSlope              Hdf5Float32 `hdf:"Hagers Slope"`
	HagarsAngle              Hdf5Float32 `hdf:"Hagers Angle"`
	HagarsRadius             Hdf5Float32 `hdf:"Hagers Radius"`
	UseWsForWeirRef          uint8       `hdf:"Use WS for Weir Reference"`
	PilotFlow                Hdf5Float32 `hdf:"Pilot Flow"`
	CulvertGroups            int32       `hdf:"Culvert Groups"`
	CulvertsFlapGates        int32       `hdf:"Culverts Flap Gates"`
	GateGroups               int32       `hdf:"Gate Groups"`
	HtabFfPoints             int32       `hdf:"HTAB FF Points"`
	HtabRcCounts             int32       `hdf:"HTAB RC Count"`
	HtabRcPoints             int32       `hdf:"HTAB RC Points"`
	HtabHwMax                Hdf5Float32 `hdf:"HTAB HW Max"`
	HtabTwMax                Hdf5Float32 `hdf:"HTAB TW Max"`
	HtabMaxFlow              Hdf5Float32 `hdf:"HTAB Max Flow"`
	CellSpacingNear          float32     `hdf:"Cell Spacing Near"`
	CellSpacingFar           float32     `hdf:"Cell Spacing Far"`
	NearRepeats              int32       `hdf:"Near Repeats"`
	ProtectionRadius         uint8       `hdf:"Protection Radius"`
	UseFrictionInMomentum    uint8       `hdf:"Use Friction in Momentum"`
	UseWeightInMomentum      uint8       `hdf:"Use Weight in Momentum"`
	UseCriticalUs            uint8       `hdf:"Use Critical US"`
	UseEgforPressureCriteria uint8       `hdf:"Use EG for Pressure Criteria"`
	IceOption                int32       `hdf:"Ice Option"`
	WeirSkew                 Hdf5Float32 `hdf:"Weir Skew"`
	PierSkew                 Hdf5Float32 `hdf:"Pier Skew"`
	BrContraction            Hdf5Float32 `hdf:"BR Contraction"`
	BrExpansion              Hdf5Float32 `hdf:"BR Expansion"`
	BrPierK                  Hdf5Float32 `hdf:"BR Pier K"`
	BrPierElev               Hdf5Float32 `hdf:"BR Pier Elev"`
	BrStructK                Hdf5Float32 `hdf:"BR Struct K"`
	BrStructElev             Hdf5Float32 `hdf:"BR Struct Elev"`
	BrStructMann             Hdf5Float32 `hdf:"BR Struct Mann"`
	BrUsLeftBank             Hdf5Float32 `hdf:"BR US Left Bank"`
	BrUsRightBank            Hdf5Float32 `hdf:"BR US Right Bank"`
	BrDsLeftBank             Hdf5Float32 `hdf:"BR DS Left Bank"`
	BrDsrightBank            Hdf5Float32 `hdf:"BR DS Right Bank"`
	XsUsLeftBank             Hdf5Float32 `hdf:"XS US Left Bank"`
	XsUsRightBank            Hdf5Float32 `hdf:"XS US Right Bank"`
	XsDsLeftBank             Hdf5Float32 `hdf:"XS DS Left Bank"`
	XsDsRightBank            Hdf5Float32 `hdf:"XS DS Right Bank"`
	UsIneffLeftSta           Hdf5Float32 `hdf:"US Ineff Left Sta"`
	UsIneffLeftElev          Hdf5Float32 `hdf:"US Ineff Left Elev"`
	UsIneffRightSta          Hdf5Float32 `hdf:"US Ineff Right Sta"`
	UsIneffRightElev         Hdf5Float32 `hdf:"US Ineff Right Elev"`
	DsIneffLeftSta           Hdf5Float32 `hdf:"DS Ineff Left Sta"`
	DsIneffLeftElev          Hdf5Float32 `hdf:"DS Ineff Left Elev"`
	DsIneffRightSta          Hdf5Float32 `hdf:"DS Ineff Right Sta"`
	DsIneffRightElev         Hdf5Float32 `hdf:"DS Ineff Right Elev"`
	UseOverrideHwCon         uint8       `hdf:"Use Override HW Connectivity"`
	UseOverrideTwCon         uint8       `hdf:"Use Override TW Connectivity"`
	UseRCFamily              uint8       `hdf:"Use RC Family"` //This wasn't present in Randy's example, But is in Bald Eagle from HEC-RAS 6.5
	UseOverideHTabIBCurves   uint8       `hdf:"Use Override HTabIBCurves"`
	SnnID                    int32       `hdf:"SNN ID"`
	DefaultCenterline        uint8       `hdf:"Default Centerline"`
}

func ReadSNetIDToNameFromGeoHDF(filePath string) (map[string]int, error) {

	f, err := util.OpenFile(filePath)
	if err != nil {
		log.Fatalln(err) //does this actually return error? if not, why have the method return error?
	}
	defer f.Close()

	isversion66 := checkVersion(f)

	if isversion66 {
		data := []structuresAttr66{}
		err = util.ReadCompoundAttributes(f, STRUCTURE_DATA_PATH, &data, nil)
		if err != nil {
			log.Fatalln(err)
		}
		snetToName := make(map[string]int, len(data))
		for _, structure := range data {
			fmt.Println(structure.Connection)
			fmt.Println(structure.SnnID)
			snetToName[structure.Connection] = int(structure.SnnID)
		}
		return snetToName, nil
	} else {
		data := []structuresAttr{}
		err = util.ReadCompoundAttributes(f, STRUCTURE_DATA_PATH, &data, nil)
		if err != nil {
			log.Fatalln(err)
		}
		snetToName := make(map[string]int, len(data))
		for _, structure := range data {
			fmt.Println(structure.Connection)
			fmt.Println(structure.SnnID)
			snetToName[structure.Connection] = int(structure.SnnID)
		}
		return snetToName, nil
	}
}
func checkVersion(f *hdf5.File) bool {
	srcrootgroup, err := f.OpenGroup("/")
	if err != nil {
		log.Fatalln(err)
	}
	defer srcrootgroup.Close()
	attr, err := srcrootgroup.OpenAttribute("File Version")
	if err != nil {
		log.Fatalln(err)
	}
	defer attr.Close()

	var attrdata string
	attr.Read(&attrdata, hdf5.T_GO_STRING)
	return strings.Contains(attrdata, "6.6")

}
