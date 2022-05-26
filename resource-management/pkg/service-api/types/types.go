package types

import "global-resource-service/resource-management/pkg/common-lib/types"

// The request content of the Watch API call
// ResourceVersionMap is part of the return of the LIST API call
type WatchRequest struct {
	ResourceVersions types.ResourceVersionMap `json:"resource_versions"`
}


// quota
type ResourceQuota struct {
	TotalMachines int `json:"total_machines",omitempty`
	// TODO: add map for machine types and special hardware request quotas
}

// client
type Client struct {
	ClientId string `json:"client_id"`
	ClientInfo ClientInfoType `json:"client_info",omitempty`
}

type ClientInfoType struct {
	ClientName string `json:"client_name",omitempty`
	Region string `json:"client_region",omitempty`
}

// Optionally, client can set its customized name and initial quota
type ClientRegistrationRequest struct {
	ClientInfo ClientInfoType `json:"client_info",omitempty`
	InitQuota ResourceQuota `json:"init_quota",omitempty`
}

type ClientRegistrationResponse struct {
	ClientId string `json:"client_id"`
	GrantedQuota ResourceQuota `json:"granted_quota"`
}

// resourceReq
// default to request all region, 10K nodes total, no special hardware
// for 630, request with default only. machine flavors, special-hardware request will be supported post 630
type ResourceRequest struct {
	TotalRequest []RequestPerRegion `json:"resource_request",omitempty`
}

// per selected region
type RequestPerRegion struct {
	// Name of the region
	RegionName string `json:"region_name"`

	// Machines requested per host machine type; machine type defined as CPU type etc.
	// flavors
	Machines map[MachineType]int `json:"request_machines",omitempty`

	// Machines requested per special hardware type, e.g., GPU / FPGA machines
	SpecialHardwareMachines map[string]int `json:"request_special_hardware_machines",omitempty`
}

// host with different hardware variations, such as CPU categories, ARM x86 etc.
type MachineType types.NodeMachineType
