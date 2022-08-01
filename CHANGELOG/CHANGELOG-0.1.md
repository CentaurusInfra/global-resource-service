# Release Summary

This is the first release of the Global Resource Service, one of the corner stones for the Regionless Cloud Platform(Arktos 2.0) project.

### The 0.1.0 release includes the following components:

* Global Resource Service API server, that supports REST APIs for client registration, List Assigned nodes and Watch for node changes.
* Performant distributor, event queues and cache to support large scale of node changes
* Data Aggregator that collects nodes and node changes from each region
* Client development SDK that provides APIs for building scheduler or other clients to the Global Resource Service
* A Region manager simulator that provides region level, multiple Resource Provider simulation of data changes
* A simulated scheduler with the cache layer
* A test infrastructure to automate service deployment, cross region test setup, test execution and result collection



### Key Features:

* Client registration, Node List and Watch APIs
* Distributor algorithms to support multiple region and resource clusters
* Distributor algorithms for efficient, balanced node resource distribution to schedulers
* Scalability: Scale up to 1m nodes cross multiple regions, with up to 40 schedulers
* Performance: End to end latency just 300ms for normal node failures cases (Daily change pattern) and within 1.3 seconds for disaster scenarios(RP outage pattern)
* Abstraction of node resources, aka, logical min. record for node resource
* Abstraction of resource version, aka, Composite RV ( or CRV ) from nodes from different and global origins. 
* Cross region data change simulation of both "Daily" and "RP outage" test scenarios
* Automatic test environment setup, test execution and result collection routines.


### Performance test results for the 0.1.0 release:

|   Test   | Total Nodes | Regions|Nodes per Region| RPs per Region | Nodes per RP | Schedulers| Nodes per scheduler list | Notes | Register<br>Latency<br>(ms) | List<br>Latency<br>(ms) | Watch<br>P50(ms) | P90(ms) | P99(ms) |
|:--------:| :---: | :---:|----:|----:|----:|----:|----:| ----:|----:|----:|----:|----:| ----:|
|  test-1  | 1m  | 5 | 200K | 10 | 20K | 20 | 25K| disable metric, daily data change pattern|301|871|108|175|211|
| test-1.1 | 1m  | 5 | 200K | 10 | 20K | 20 | 25K| disable metric, RP down data change pattern|374|1012|1021|1137|1156|
|  test-2  | 1m  | 5 | 200K | 10 | 20K | 20 | 25K| enable metric, daily data change pattern|298|1097|116|181|201|
| test-2.1 | 1m  | 5 | 200K | 10 | 20K | 20 | 25K| enable metric,RP down data change pattern|359|1012|1002|1074|1093|
|  test-3  | 1m  | 5 | 200K | 10 | 20K | 20 | 50K| disable metric, daily data change pattern|369|1766|109|173|217|
| test-3.1 | 1m  | 5 | 200K | 10 | 20K | 20 | 50K| disable metric, RP down data change pattern|337|1679|877|1174|1200|
|  test-4  | 1m  | 5 | 200K | 10 | 20K | 40 | 25K| disable metric, daily data change pattern|135|811|92|161|195|
|  test-5  | 1m  | 5 | 200K | 10 | 20K | 20 | 25K| disable metric, rp down, all in one region|209|641|451|513|529|

*Service is gce n1-standard-32 VM with 32 core and 120GB Ram, 500GB SSD, premium network. 
*Scheduler and the resource simulators are can with n1-standard-8 VMs*


#### Regions each component deployed


#### Service:

|        Region |             Location |
|--------------:|---------------------:|
| us Central-1a | Council bluffs, IOWA |



#### Simulators:

|        Region |             Location |
|--------------:|---------------------:|
| us Central-1a | Council bluffs, IOWA |
|    us east1-b |    Moncks COrner, SC |
|    us west2-a |               LA, CA |
|    us west3-c | Salt Lake city, Utah |
|    us west4-a |    las Vegas, Nevada |


#### Schedulers:

|     Region |             Location |
|-----------:|---------------------:|
| us west3-b | Salt Lake city, Utah |
| us east4-b |    Ashburn, Virginia |