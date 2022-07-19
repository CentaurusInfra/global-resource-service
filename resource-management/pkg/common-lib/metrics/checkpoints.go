package metrics

type ResourceManagementCheckpoint int

const (
	Aggregator_Received ResourceManagementCheckpoint = 0

	Distributor_Received ResourceManagementCheckpoint = 1
	Distributor_Sending  ResourceManagementCheckpoint = 2
	Distributor_Sent     ResourceManagementCheckpoint = 3

	Serializer_Encoded ResourceManagementCheckpoint = 4
	Serializer_Sent    ResourceManagementCheckpoint = 5

	Len_ResourceManagementCheckpoint = 6
)

var ResourceManagementMeasurement_Enabled = true

func SetEnableResourceManagementMeasurement(enabled bool) {
	ResourceManagementMeasurement_Enabled = enabled
}
