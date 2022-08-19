package service_api

import (
	"global-resource-service/test/integration/framework"
	"testing"
)

func TestMain(m *testing.M) {
	framework.RedisMain(m.Run)
}
