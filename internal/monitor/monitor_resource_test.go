package monitor_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/encoding"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/internal/apptest"
	"github.com/ilxqx/vef-framework-go/monitor"
	"github.com/ilxqx/vef-framework-go/result"
)

type MonitorResourceTestSuite struct {
	suite.Suite

	ctx     context.Context
	app     *app.App
	stop    func()
	service monitor.Service
}

func (suite *MonitorResourceTestSuite) SetupSuite() {
	suite.T().Log("Setting up MonitorResourceTestSuite - starting test app")

	suite.ctx = context.Background()

	monitorConfig := &config.MonitorConfig{
		SampleInterval: 100 * time.Millisecond,
		SampleDuration: 50 * time.Millisecond,
	}

	buildInfo := &monitor.BuildInfo{
		AppVersion: "v1.0.0-test",
		BuildTime:  "2024-01-01T00:00:00Z",
		GitCommit:  "test123abc",
	}

	suite.app, suite.stop = apptest.NewTestApp(
		suite.T(),
		fx.Replace(
			&config.DatasourceConfig{
				Type: "sqlite",
			},
			monitorConfig,
		),
		fx.Supply(buildInfo),
		fx.Populate(&suite.service),
	)

	time.Sleep(100 * time.Millisecond)
}

func (suite *MonitorResourceTestSuite) TearDownSuite() {
	suite.T().Log("Tearing down MonitorResourceTestSuite")

	if suite.stop != nil {
		suite.stop()
	}
}

func (suite *MonitorResourceTestSuite) makeAPIRequest(body api.Request) *http.Response {
	jsonBody, err := encoding.ToJSON(body)
	suite.Require().NoError(err, "Should encode request to JSON")

	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(jsonBody))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err, "API request should not fail")

	return resp
}

func (suite *MonitorResourceTestSuite) readBody(resp *http.Response) result.Result {
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	suite.Require().NoError(err, "Should read response body")
	res, err := encoding.FromJSON[result.Result](string(body))
	suite.Require().NoError(err, "Should decode response JSON")

	return *res
}

func (suite *MonitorResourceTestSuite) readDataAsMap(data any) map[string]any {
	m, ok := data.(map[string]any)
	suite.Require().True(ok, "Data should be a map")

	return m
}

func (suite *MonitorResourceTestSuite) TestGetOverview() {
	suite.T().Log("Testing get_overview endpoint")

	suite.Run("Success", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/monitor",
				Action:   "get_overview",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Overview request should succeed")

		data := suite.readDataAsMap(body.Data)

		suite.Contains(data, "host", "Should have host info")
		suite.Contains(data, "cpu", "Should have CPU info")
		suite.Contains(data, "memory", "Should have memory info")
		suite.Contains(data, "disk", "Should have disk info")
		suite.Contains(data, "network", "Should have network info")
		suite.Contains(data, "process", "Should have process info")
		suite.Contains(data, "load", "Should have load info")
		suite.Contains(data, "build", "Should have build info")

		buildInfo := suite.readDataAsMap(data["build"])
		suite.Equal("v1.0.0-test", buildInfo["appVersion"], "AppVersion should match")
		suite.NotEmpty(buildInfo["vefVersion"], "VEFVersion should be populated")
		suite.Equal("2024-01-01T00:00:00Z", buildInfo["buildTime"], "BuildTime should match")
		suite.Equal("test123abc", buildInfo["gitCommit"], "GitCommit should match")
	})
}

func (suite *MonitorResourceTestSuite) TestGetCPU() {
	suite.T().Log("Testing get_cpu endpoint")

	suite.Run("Success", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/monitor",
				Action:   "get_cpu",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "CPU request should succeed")

		data := suite.readDataAsMap(body.Data)

		suite.Contains(data, "physicalCores", "Should have physical cores")
		suite.Contains(data, "logicalCores", "Should have logical cores")
		suite.Contains(data, "modelName", "Should have model name")
		suite.Contains(data, "usagePercent", "Should have usage percent")
		suite.Contains(data, "totalPercent", "Should have total percent")

		physicalCores, ok := data["physicalCores"].(float64)
		suite.True(ok, "Physical cores should be a number")
		suite.Greater(physicalCores, float64(0), "Should have at least 1 physical core")

		logicalCores, ok := data["logicalCores"].(float64)
		suite.True(ok, "Logical cores should be a number")
		suite.GreaterOrEqual(logicalCores, physicalCores, "Logical cores should be >= physical cores")
	})
}

func (suite *MonitorResourceTestSuite) TestGetMemory() {
	suite.T().Log("Testing get_memory endpoint")

	suite.Run("Success", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/monitor",
				Action:   "get_memory",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Memory request should succeed")

		data := suite.readDataAsMap(body.Data)

		suite.Contains(data, "virtual", "Should have virtual memory info")

		virtual := suite.readDataAsMap(data["virtual"])
		suite.Contains(virtual, "total", "Should have total memory")
		suite.Contains(virtual, "used", "Should have used memory")
		suite.Contains(virtual, "usedPercent", "Should have used percent")

		total, ok := virtual["total"].(float64)
		suite.True(ok, "Total should be a number")
		suite.Greater(total, float64(0), "Total memory should be > 0")
	})
}

func (suite *MonitorResourceTestSuite) TestGetDisk() {
	suite.T().Log("Testing get_disk endpoint")

	suite.Run("Success", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/monitor",
				Action:   "get_disk",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Disk request should succeed")

		data := suite.readDataAsMap(body.Data)

		suite.Contains(data, "partitions", "Should have partitions")

		partitions, ok := data["partitions"].([]any)
		suite.True(ok, "Partitions should be an array")
		suite.NotEmpty(partitions, "Should have at least one partition")

		firstPartition := suite.readDataAsMap(partitions[0])
		suite.Contains(firstPartition, "mountPoint", "Should have mount point")
		suite.Contains(firstPartition, "total", "Should have total size")
	})
}

func (suite *MonitorResourceTestSuite) TestGetNetwork() {
	suite.T().Log("Testing get_network endpoint")

	suite.Run("Success", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/monitor",
				Action:   "get_network",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Network request should succeed")

		data := suite.readDataAsMap(body.Data)

		suite.Contains(data, "interfaces", "Should have interfaces")

		interfaces, ok := data["interfaces"].([]any)
		suite.True(ok, "Interfaces should be an array")
		suite.NotEmpty(interfaces, "Should have at least one network interface")

		firstInterface := suite.readDataAsMap(interfaces[0])
		suite.Contains(firstInterface, "name", "Should have interface name")
		suite.NotEmpty(firstInterface["name"], "Interface name should not be empty")
	})
}

func (suite *MonitorResourceTestSuite) TestGetHost() {
	suite.T().Log("Testing get_host endpoint")

	suite.Run("Success", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/monitor",
				Action:   "get_host",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Host request should succeed")

		data := suite.readDataAsMap(body.Data)

		suite.Contains(data, "hostname", "Should have hostname")
		suite.Contains(data, "os", "Should have OS")
		suite.Contains(data, "platform", "Should have platform")

		suite.NotEmpty(data["hostname"], "Hostname should not be empty")
		suite.NotEmpty(data["os"], "OS should not be empty")
	})

	suite.Run("ConsistentResults", func() {
		resp1 := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/monitor",
				Action:   "get_host",
				Version:  "v1",
			},
		})

		body1 := suite.readBody(resp1)
		data1 := suite.readDataAsMap(body1.Data)

		resp2 := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/monitor",
				Action:   "get_host",
				Version:  "v1",
			},
		})

		body2 := suite.readBody(resp2)
		data2 := suite.readDataAsMap(body2.Data)

		suite.Equal(data1["hostname"], data2["hostname"], "Hostname should be consistent")
		suite.Equal(data1["os"], data2["os"], "OS should be consistent")
	})
}

func (suite *MonitorResourceTestSuite) TestGetProcess() {
	suite.T().Log("Testing get_process endpoint")

	suite.Run("Success", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/monitor",
				Action:   "get_process",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Process request should succeed")

		data := suite.readDataAsMap(body.Data)

		suite.Contains(data, "pid", "Should have PID")
		suite.Contains(data, "name", "Should have process name")
		suite.Contains(data, "cpuPercent", "Should have CPU percent")
		suite.Contains(data, "memoryPercent", "Should have memory percent")
		suite.Contains(data, "memoryRss", "Should have memory RSS")

		pid, ok := data["pid"].(float64)
		suite.True(ok, "PID should be a number")
		suite.Greater(pid, float64(0), "PID should be > 0")

		suite.NotEmpty(data["name"], "Process name should not be empty")
	})
}

func (suite *MonitorResourceTestSuite) TestGetLoad() {
	suite.T().Log("Testing get_load endpoint")

	suite.Run("Success", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/monitor",
				Action:   "get_load",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Load request should succeed")

		data := suite.readDataAsMap(body.Data)

		suite.Contains(data, "load1", "Should have 1-minute load")
		suite.Contains(data, "load5", "Should have 5-minute load")
		suite.Contains(data, "load15", "Should have 15-minute load")

		load1, ok := data["load1"].(float64)
		suite.True(ok, "Load1 should be a number")
		suite.GreaterOrEqual(load1, float64(0), "Load1 should be >= 0")
	})
}

func (suite *MonitorResourceTestSuite) TestGetBuildInfo() {
	suite.T().Log("Testing get_build_info endpoint")

	suite.Run("Success", func() {
		resp := suite.makeAPIRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "sys/monitor",
				Action:   "get_build_info",
				Version:  "v1",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.True(body.IsOk(), "Build info request should succeed")

		data := suite.readDataAsMap(body.Data)

		suite.Equal("v1.0.0-test", data["appVersion"], "AppVersion should match")
		suite.NotEmpty(data["vefVersion"], "VEFVersion should be populated")
		suite.Equal("2024-01-01T00:00:00Z", data["buildTime"], "BuildTime should match")
		suite.Equal("test123abc", data["gitCommit"], "GitCommit should match")
	})
}

func TestMonitorResourceSuite(t *testing.T) {
	suite.Run(t, new(MonitorResourceTestSuite))
}
