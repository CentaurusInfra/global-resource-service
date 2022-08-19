/*
Copyright 2022 Authors of Global Resource Service.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"global-resource-service/resource-management/pkg/store/redis"
	"os"
)

// EtcdMain starts an etcd instance before running tests.
func RedisMain(tests func() int) {
	redis.NewRedisClient()
	/*
		var wg sync.WaitGroup
		wg.Add(2)
		go StartTestService(&wg)
		go StartTestSimulator(&wg)
		wg.Wait()
	*/
	result := tests()
	//stop() // Don't defer this. See os.Exit documentation.
	os.Exit(result)
}
