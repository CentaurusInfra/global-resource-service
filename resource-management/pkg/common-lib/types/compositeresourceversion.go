/*
Copyright 2022 Authors of Global Resource Service.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package types

import (
	"encoding/json"

	"global-resource-service/resource-management/pkg/common-lib/types/location"
)

type CompositeResourceVersion struct {
	RegionId            string
	ResourcePartitionId string
	ResourceVersion     uint64
}

type RvLocation struct {
	Region    location.Region
	Partition location.ResourcePartition
}

func (loc RvLocation) MarshalText() (text []byte, err error) {
	type l RvLocation
	return json.Marshal(l(loc))
}

func (loc *RvLocation) UnmarshalText(text []byte) error {
	type l RvLocation
	return json.Unmarshal(text, (*l)(loc))
}

// Map from (regionId, ResourcePartitionId) to resourceVersion
// used in REST API calls
type TransitResourceVersionMap map[RvLocation]uint64

// internally used in the eventqueue used in WATCH of nodes
type InternalResourceVersionMap map[location.Location]uint64

func ConvertToInternalResourceVersionMap(rvs TransitResourceVersionMap) InternalResourceVersionMap {
	internalMap := make(InternalResourceVersionMap)

	for k, v := range rvs {
		internalMap[*location.NewLocation(k.Region, k.Partition)] = v
	}

	return internalMap
}

func (rvs *TransitResourceVersionMap) Copy() TransitResourceVersionMap {
	dupRVs := make(TransitResourceVersionMap, len(*rvs))
	for loc, rv := range *rvs {
		dupRVs[loc] = rv
	}

	return dupRVs
}
