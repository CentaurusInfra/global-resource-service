package types

import "global-resource-service/resource-management/pkg/common-lib/types"

// The request content of the Watch API call
// ResourceVersionMap is part of the return of the LIST API call
type WatchRequest struct {
	ResourceVersions types.ResourceVersionMap `json:"resource_versions"`
}

// Optionally, client can set its customized name and initial quota
type ClientRegistrationRequest struct {
	ClientInfo types.ClientInfoType `json:"client_info",omitempty`
	InitQuota types.ResourceQuota `json:"init_quota",omitempty`
}

type ClientRegistrationResponse struct {
	ClientId string `json:"client_id"`
	GrantedQuota types.ResourceQuota `json:"granted_quota"`
}

// resourceReq
// default to request all region, 10K nodes total, no special hardware
// for 630, request with default only. machine flavors, special-hardware request will be supported post 630
type ResourceRequest struct {
	TotalRequest []types.ResourcePerRegion `json:"resource_request",omitempty`
}
