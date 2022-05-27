package app

import (
	"fmt"
	"global-resource-service/resource-management/pkg/common-lib/interfaces/store"
	"global-resource-service/resource-management/pkg/common-lib/types"
	"net/http"
	"time"

	"global-resource-service/resource-management/pkg/aggregrator"
	"global-resource-service/resource-management/pkg/distributor"
	"global-resource-service/resource-management/pkg/service-api/endpoints"

	"github.com/gorilla/mux"
	"k8s.io/klog/v2"
)

type Config struct {
	ResourceUrls []string
	MasterIp     string
	MasterPort   string
}

// replace this with the redis store
type fs struct {}

func (f *fs) PersistClient (id string, details *types.ClientInfoType) error {
	return fmt.Errorf("not impl")
}
func (f *fs) PersistNodes(nodesToPersist []*types.LogicalNode) bool {
	return false
}

func (f *fs) PersistNodeStoreStatus(nodeStoreStatus *store.NodeStoreStatus) bool {
	return false
}

func (f *fs) PersistVirtualNodesAssignments(assignment *store.VirtualNodeAssignment) bool {
	 return false
}
// Run and create new service-api.  This should never exit.
func Run(c *Config) error {
	klog.V(3).Infof("Starting the API server...")

	store := &fs{}
	dist := distributor.GetResourceDistributor()
	dist.SetPersistHelper(store)
	installer := endpoints.NewInstaller(dist)

	r := mux.NewRouter().StrictSlash(true)

	// TODO: reuse k8s mux wrapper, pathrecorder.go for simplify this handler by each path
	r.HandleFunc(endpoints.ListResourcePath, installer.ResourceHandler).Methods("GET")
	r.HandleFunc(endpoints.WatchResourcePath, installer.ResourceHandler)
	r.HandleFunc(endpoints.UpdateResourcePath, installer.ResourceHandler)
	r.HandleFunc(endpoints.ReduceResourcePath, installer.ResourceHandler)

	r.HandleFunc(endpoints.ClientAdminitrationPath, installer.ClientAdministrationHandler)
	r.HandleFunc(endpoints.ClientAdminitrationPath + "/{clientId}", installer.ClientAdministrationHandler)

	var addr string
	var p string

	if c.MasterIp == "" {
		addr = "localhost"
	}

	if c.MasterPort == "" {
		p = endpoints.InsecureServiceAPIPort
	}

	server := &http.Server{
		Handler:      r,
		Addr:         addr + p,
		WriteTimeout: 2 * time.Second,
		ReadTimeout:  2 * time.Second,
	}

	err := server.ListenAndServe()

	if err != nil {
		return err
	}

	// start the aggregator instance
	klog.V(3).Infof("Starting the Aggregator ...")

	aggregator := aggregrator.NewAggregator(c.ResourceUrls, dist)
	err = aggregator.Run()
	if err != nil {
		return err
	}

	return nil
}
