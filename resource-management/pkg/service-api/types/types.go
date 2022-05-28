package types

import "global-resource-service/resource-management/pkg/common-lib/types"

// WatchRequest is the request body of the Watch API call
// ResourceVersionMap is part of the return of the LIST API call
type WatchRequest struct {
	ResourceVersions types.ResourceVersionMap `json:"resource_versions"`
}

// ClientRegistrationRequest is the request body when a client register to the resource management service
// TBD: Optionally, client can set its customized name and initial quota
type ClientRegistrationRequest struct {
	ClientInfo types.ClientInfoType `json:"client_info,omitempty"`
	InitQuota types.ResourceQuota `json:"init_quota,omitempty"`
}

// ClientRegistrationResponse is the response body for approved client registration request
// ClientId is required for an approved client registration to the resource management service
// GrantedQuota is an info to client on the resource level the List OP it can request
type ClientRegistrationResponse struct {
	ClientId string `json:"client_id"`
	GrantedQuota types.ResourceQuota `json:"granted_quota,omitempty"`
}

// ResourceRequest is used in the http request body for client to List resources
// default to request all region, 10K nodes total, no special hardware
// for 630, request with default only. machine flavors, special-hardware request will be supported post 630
type ResourceRequest struct {
	TotalMachines int `json:"total_machines" default:"10000"`
	RequestInRegions []types.ResourcePerRegion `json:"resource_request,omitempty"`
}
