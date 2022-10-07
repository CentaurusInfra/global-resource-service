## Table of Contents
- [Test Environment Configurations](#test-environment)
- [Performance Result for Release 0.2.0](#performance-result-for-release-020)
- [Performance Result for Release 0.1.0](#performance-result-for-release-010)

## Test Environment

### Test Config
* 5 regions in GCE cross US continent, premium network
* Each region uses one simulator to simulate 40 clusters; each cluster has 25K nodes, total 1M nodes per region.
* 100 schedulers (clients of GRS) each request 50K nodes.

### Test hosts configuration
* GRS service: n1-standard-32 VM with 32 core and 120GB Ram, 500GB SSD,
* Scheduler and simulators: n1-standard-8 VMs

### Services locations
* GRS Service:

|        Region |             Location |
|--------------:|---------------------:|
| us Central-1a | Council bluffs, IOWA |

* Simulators:

|        Region |             Location |
|--------------:|---------------------:|
| us Central-1a | Council bluffs, IOWA |
|    us east1-b |    Moncks COrner, SC |
|    us west2-a |               LA, CA |
|    us west3-c | Salt Lake city, Utah |
|    us west4-a |    las Vegas, Nevada |

* Schedulers:

|     Region |             Location |
|-----------:|---------------------:|
| us west3-b | Salt Lake city, Utah |
| us east4-b |    Ashburn, Virginia |

### Performance Result for Release 0.2.0
<table>
<tr rowspan=2>
<td rowspan=2>Test Case</td>
<td colspan=3>Watch Latency (ms)</td>
<td rowspan=2>List (ms)</td>
<td rowspan=2>Register (ms)</td>
<td rowspan=2>Throughput <br>(events/s)</td>
</tr>
<tr>
<td>P50</td>
<td>P90</td>
<td>P99</td>
</tr>
<tr>
<td>Regular changes</td>
<td align=right>32</td>
<td align=right>73</td>
<td align=right>88</td>
<td align=right>1,389</td>
<td align=right>64</td>
<td align=right>N/A</td>
</tr>
<tr>
<td>Massive outages</td>
<td align=right>23,606</td>
<td align=right>44,900</td>
<td align=right>47,810</td>
<td align=right>1,381</td>
<td align=right>66</td>
<td align=right>96,241</td>
</tr>
</table>


### Performance Result for Release 0.1.0:
* Total number of nodes: 1M
* Number of regions: 5
* Number of nodes per region: 200K
* Resouce parition (RP) per region: 10
* Number of nodes per RP: 20K
* Daily pattern performance data:

|   Test   | Schedulers| Nodes per scheduler list | Register<br>Latency<br>(ms) | List<br>Latency<br>(ms) | Watch<br>P50(ms) | P90(ms) | P99(ms) | Metrics |
|:--------:| ----:|----:| ----:|----:|----:|----:|----:|------:|
|  test-1  | 20 | 25K| 301|871|108|175|211| Disabled| 
|  test-2  | 20 | 25K| 298|1097|116|181|201| Enabled|
|  test-3  | 20 | 50K| 369|1766|109|173|217| Disabled|
|  test-4  | 40 | 25K| 135|811|92|161|195| Disabled|

* 1 RP per region outage performance data:

|   Test   | Schedulers| Nodes per scheduler list | Register<br>Latency<br>(ms) | List<br>Latency<br>(ms) | Watch<br>P50(ms) | P90(ms) | P99(ms) | Metrics |
|:--------:| ----:|----:| ----:|----:|----:|----:|----:|---------:|
| test-1.1 | 20 | 25K| 374|1012|1021|1137|1156| Disabled |
| test-2.1 | 20 | 25K| 359|1012|1002|1074|1093|  Enabled |
| test-3.1 | 20 | 50K| 337|1679|877|1174|1200| Disabled |
|  test-5  | 20 | 25K| 209|641|451|513|529| Disabled |
