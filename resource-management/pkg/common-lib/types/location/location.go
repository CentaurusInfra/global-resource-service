package types

import "fmt"

const RingRange = float64(360)

type Location struct {
	regionId   string
	paritionId string
}

func NewLocation(regionId, partitionId string) *Location {
	return &Location{
		regionId:   regionId,
		paritionId: partitionId,
	}
}

type arc struct {
	lower float64
	upper float64
}

// Region defines the possible region location of a given node
// Defined and doced by region admin

const (
	// Regions
	Beijing   = "Beijing"
	Shanghai  = "Shanghai"
	Wulan     = "Wulan"
	Guizhou   = "Guizhou"
	Reserved1 = "Reserved1"
	Reserved2 = "Reserved2"
	Reserved3 = "Reserved3"
	Reserved4 = "Reserved4"
	Reserved5 = "Reserved5"
)

var regions = map[int]string{
	0: Beijing,
	1: Shanghai,
	2: Wulan,
	3: Guizhou,
	4: Reserved1,
	5: Reserved2,
	6: Reserved3,
	7: Reserved4,
	8: Reserved5,
}

var regionToArc map[string]arc

const (
	// Resource partitions
	ResourcePartition1  = "RP1"
	ResourcePartition2  = "RP2"
	ResourcePartition3  = "RP3"
	ResourcePartition4  = "RP4"
	ResourcePartition5  = "RP5"
	ResourcePartition6  = "RP6"
	ResourcePartition7  = "RP7"
	ResourcePartition8  = "RP8"
	ResourcePartition9  = "RP9"
	ResourcePartition10 = "RP10"
)

var ResourcePartitions = []string{ResourcePartition1, ResourcePartition2, ResourcePartition3, ResourcePartition4, ResourcePartition5,
	ResourcePartition6, ResourcePartition7, ResourcePartition8, ResourcePartition9, ResourcePartition10}
var regionRPToArc map[Location]arc

func init() {
	regionRPToArc = make(map[Location]arc)
	regionGrain := RingRange / float64(len(regions))

	regionLower := float64(0)
	regionUpper := regionGrain

	for i := 0; i < len(regions); i++ {
		region := regions[i]
		rps := GetRPsForRegion(region)

		rpLower := regionLower
		rpGrain := regionGrain / float64(len(rps))
		rpUpper := regionLower + rpGrain
		for j := 0; j < len(rps); j++ {
			rp := rps[j]
			loc := Location{
				regionId:   region,
				paritionId: rp,
			}
			if j == len(rps)-1 {
				if i == len(regions)-1 {
					regionRPToArc[loc] = arc{lower: rpLower, upper: RingRange}
				} else {
					regionRPToArc[loc] = arc{lower: rpLower, upper: regionUpper}
				}
			} else {
				regionRPToArc[loc] = arc{lower: rpLower, upper: rpUpper}
				rpLower = rpUpper
				rpUpper += rpGrain
			}
		}

		regionLower = regionUpper
		regionUpper += regionGrain
	}
}

func GetRegion(index int) string {
	if regionName, isOK := regions[index]; isOK {
		return regionName
	}
	return ""
}

func GetRegionNum() int {
	return len(regions)
}

func GetRPNum() int {
	return len(ResourcePartitions)
}

// TODO - read resource parition from configuration or metadata server
func GetRPsForRegion(region string) []string {
	rpsForRegion := make([]string, len(ResourcePartitions))
	for i := 0; i < len(ResourcePartitions); i++ {
		rpsForRegion[i] = fmt.Sprintf("%s_%s", region, ResourcePartitions[i])
	}
	return rpsForRegion
}

func (loc *Location) GetArcRangeFromLocation() (float64, float64) {
	locArc := regionRPToArc[*loc]
	return locArc.lower, locArc.upper
}

func (loc *Location) Equal(locToCompare Location) bool {
	return loc.regionId == loc.regionId && loc.paritionId == locToCompare.paritionId
}

func (loc *Location) String() string {
	return fmt.Sprintf("[Region %s, ResoucePartition %s]", loc.regionId, loc.paritionId)
}
