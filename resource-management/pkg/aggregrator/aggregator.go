package aggregrator

import (
	"encoding/json"
	"fmt"
	"global-resource-service/resource-management/pkg/common-lib/interfaces"
	"global-resource-service/resource-management/pkg/common-lib/types"
	"global-resource-service/resource-management/pkg/common-lib/types/event"
	"io/ioutil"
	"strings"

	"net/http"
	"sync"
	"time"
)

type Aggregator struct {
	urls          []string
	ProcessEvents interfaces.Interface
}

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

type Response struct {
	MinRecordNodeEvents []*event.NodeEvent
	RvMap               types.ResourceVersionMap
	Length              int
}

type PullWatchData struct {
	BatchLength int
	CRV         types.ResourceVersionMap
}

const (
	DefaultBatchLength = 1000
)

// Initialize aggregator
func NewAggregator(urls []string, ProcessEvents interfaces.Interface) *Aggregator {
	return &Aggregator{
		urls:          urls,
		ProcessEvents: ProcessEvents,
	}
}

// Main loop to get resources from resource region managers and send to distributor
func (a *Aggregator) Run() {
	//
	numberOfURLs := len(a.urls)

	var wg sync.WaitGroup
	wg.Add(numberOfURLs)

	fmt.Println("Running for loop to connect to to resource region manager...")

	for i := 0; i < numberOfURLs; i++ {
		go func(i int) {
			defer wg.Done()
			var crv types.ResourceVersionMap
			var minRecordNodeEvents []*event.NodeEvent

			for {
				// Connect to resource region manager
				c := a.createClient(a.urls[i])

				// Call the PULL methods, composite RV is nil, the first PULL List. otherwise call the PULL Watch
				if crv == nil {
					minRecordNodeEvents, _, _ = a.pullList(c)
				} else {
					minRecordNodeEvents, _ = a.pullWatch(c, DefaultBatchLength, crv)
				}

				// Call ProcessEvents() and get the CRV from distributor as default success
				_, crv = a.ProcessEvents.ProcessEvents(minRecordNodeEvents)

				// Call resource region manager, POST CRV to release old node events
				a.postCRV(c, crv)
			}
		}(i)
	}

	wg.Wait()

	fmt.Println("Finished for loop to connect to to resource region manager...")
}

// Connect to resource region manager
func (a *Aggregator) createClient(url string) *Client {
	return &Client{
		BaseURL: url,
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}
}

// Call resource region manager's PULLLIST method {url}/resources/pulllist
func (a *Aggregator) pullList(c *Client) ([]*event.NodeEvent, types.ResourceVersionMap, int) {
	path := "c.baseURL/resources/pullList"
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		fmt.Print(err.Error())
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}

	var ResponseObject Response
	json.Unmarshal(bodyBytes, &ResponseObject)

	return ResponseObject.MinRecordNodeEvents, ResponseObject.RvMap, ResponseObject.Length
}

// Call the resource region manager's PULLWATCH method {url}/resources/pullwatch
func (a *Aggregator) pullWatch(c *Client, batchLength int, crv types.ResourceVersionMap) ([]*event.NodeEvent, int) {
	path := "c.baseURL/resources/pullwatch"
	bytes, _ := json.Marshal(PullWatchData{BatchLength: batchLength, CRV: crv.Copy()})
	req, err := http.NewRequest(http.MethodGet, path, strings.NewReader((string(bytes))))

	if err != nil {
		fmt.Print(err.Error())
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}

	var ResponseObject Response
	json.Unmarshal(bodyBytes, &ResponseObject)

	return ResponseObject.MinRecordNodeEvents, ResponseObject.Length
}

// Call resource region manager's POST method {url}/resources/crv to update the CRV
// error indicate failed POST, CRV means Composite Resource Version
func (a *Aggregator) postCRV(c *Client, crv types.ResourceVersionMap) error {
	path := "c.baseURL/resources/crv"
	bytes, _ := json.Marshal(crv.Copy())
	req, err := http.NewRequest(http.MethodPost, path, strings.NewReader((string(bytes))))

	if err != nil {
		fmt.Print(err.Error())
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	_, err = c.HTTPClient.Do(req)

	return err
}
