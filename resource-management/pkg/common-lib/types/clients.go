package types

const (
	// Max and Min request per client request during registration or update resources
	MaxTotalMachinesPerRequest = 25000
	MinTotalMachinesPerRequest = 1000
)

// ResourceQuota is client resource quota for a given client
// resource quota can be super set of each client list request of resources
// TBD post 630.
type ResourceQuota struct {
	TotalMachines int `json:"total_machines,omitempty" default:"10000"`
	RequestInRegions []ResourcePerRegion `json:"resource_request,omitempty"`
}

// Client is detailed client info for a registered client to the resource management service
type Client struct {
	ClientId string `json:"client_id"`
	ClientInfo ClientInfoType `json:"client_info,omitempty"`
}

type ClientInfoType struct {
	ClientName string `json:"client_name,omitempty"`
	// which region the client is from
	Region string `json:"client_region,omitempty"`
}

// ResourcePerRegion is resource request for each region
// sum of the ResourcePerRegion is the total request for a given client
type ResourcePerRegion struct {
	// Name of the region
	RegionName string `json:"region_name"`

	// Machines requested per host machine type; machine type defined as CPU type etc.
	// flavors
	Machines map[NodeMachineType]int `json:"request_machines,omitempty"`

	// Machines requested per special hardware type, e.g., GPU / FPGA machines
	SpecialHardwareMachines map[string]int `json:"request_special_hardware_machines,omitempty"`
}
