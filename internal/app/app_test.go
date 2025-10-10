package app_test

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go"
	apiPkg "github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	appTest "github.com/ilxqx/vef-framework-go/internal/app/test"
	"github.com/ilxqx/vef-framework-go/result"
)

// TestAppStartStop tests basic app lifecycle using fx.
func TestAppStartStop(t *testing.T) {
	testApp, stop := appTest.NewTestApp(
		t,
		fx.Replace(&config.DatasourceConfig{
			Type: constants.DbSQLite,
		}),
	)
	defer stop()

	require.NotNil(t, testApp)

	// Start app
	errChan := testApp.Start()
	err := <-errChan
	require.NoError(t, err)

	// Give it a moment to fully start
	time.Sleep(100 * time.Millisecond)

	// Stop app
	err = testApp.Stop()
	require.NoError(t, err)
}

// TestResource is a simple test resource for API testing.
type TestResource struct {
	apiPkg.Resource
}

func NewTestResource() apiPkg.Resource {
	return &TestResource{
		Resource: apiPkg.NewResource(
			"test",
			apiPkg.WithAPIs(
				apiPkg.Spec{
					Action: "ping",
					Public: true,
				},
			),
		),
	}
}

func (r *TestResource) Ping(ctx fiber.Ctx) error {
	return result.Ok("pong").Response(ctx)
}

// TestAppWithCustomResource tests app with custom API resource.
func TestAppWithCustomResource(t *testing.T) {
	testApp, stop := appTest.NewTestApp(
		t,
		fx.Replace(&config.DatasourceConfig{
			Type: constants.DbSQLite,
		}),
		vef.ProvideAPIResource(NewTestResource),
	)
	defer stop()

	require.NotNil(t, testApp)

	// Test the API
	req := httptest.NewRequest(
		fiber.MethodPost,
		"/api",
		strings.NewReader(`{"resource": "test", "action": "ping", "version": "v1"}`),
	)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := testApp.Test(req, 2*time.Second)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, `{"code":0,"message":"成功","data":"pong"}`, string(body))
}
