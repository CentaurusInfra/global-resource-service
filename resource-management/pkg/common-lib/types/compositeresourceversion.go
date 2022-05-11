package types

type CompositeResourceVersion struct {
	RegionId            string
	ResourcePartitionId string
	ResourceVersion     uint64
}

// Map from (regionId, ResourcePartitionId) to resourceVersion
type ResourceVersionMap map[Location]uint64

func (rvs *ResourceVersionMap) Copy() ResourceVersionMap {
	dupRVs := make(ResourceVersionMap, len(*rvs))
	for loc, rv := range *rvs {
		dupRVs[loc] = rv
	}

	return dupRVs
}
