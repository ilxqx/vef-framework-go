package app_test

import (
	"io"
	"net/http/httptest"
	"os"
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
	"github.com/ilxqx/vef-framework-go/i18n"
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

// TestResource is a simple test resource for Api testing.
type TestResource struct {
	apiPkg.Resource
}

func NewTestResource() apiPkg.Resource {
	return &TestResource{
		Resource: apiPkg.NewResource(
			"test",
			apiPkg.WithApis(
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

// TestAppWithCustomResource tests app with custom Api resource.
func TestAppWithCustomResource(t *testing.T) {
	// Save and clear the environment variable to test with default language (zh-CN)
	originalEnv := os.Getenv("VEF_I18N_LANGUAGE")

	os.Unsetenv("VEF_I18N_LANGUAGE")
	defer func() {
		if originalEnv != "" {
			os.Setenv("VEF_I18N_LANGUAGE", originalEnv)
		}
	}()

	testApp, stop := appTest.NewTestApp(
		t,
		fx.Replace(&config.DatasourceConfig{
			Type: constants.DbSQLite,
		}),
		fx.Invoke(func() {
			// Re-initialize i18n with default language after clearing env var
			// This is necessary because i18n is initialized at package level
			_ = i18n.SetLanguage("")
		}),
		vef.ProvideApiResource(NewTestResource),
	)
	defer stop()

	require.NotNil(t, testApp)

	// Test the Api
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
