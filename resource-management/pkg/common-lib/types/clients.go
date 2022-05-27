package types

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

// per selected region
type ResourcePerRegion struct {
	// Name of the region
	RegionName string `json:"region_name"`

	// Machines requested per host machine type; machine type defined as CPU type etc.
	// flavors
	Machines map[NodeMachineType]int `json:"request_machines",omitempty`

	// Machines requested per special hardware type, e.g., GPU / FPGA machines
	SpecialHardwareMachines map[string]int `json:"request_special_hardware_machines",omitempty`
}
