package api_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go"
	apiPkg "github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	appTest "github.com/ilxqx/vef-framework-go/internal/app/test"
)

// TestDuplicateApiDetection tests that duplicate Api definitions are properly detected and rejected.
func TestDuplicateApiDetection(t *testing.T) {
	t.Run("DetectDuplicateInSameResource", func(t *testing.T) {
		resource := &DuplicateActionResource{
			Resource: apiPkg.NewResource(
				"test/duplicate",
				apiPkg.WithApis(
					apiPkg.Spec{Action: "test_action"},
					apiPkg.Spec{Action: "test_action"}, // ❌ Duplicate!
				),
			),
		}

		opts := []fx.Option{
			vef.ProvideApiResource(func() apiPkg.Resource {
				return resource
			}),
			fx.Replace(&config.DatasourceConfig{
				Type: constants.DbSQLite,
			}),
		}

		_, stop, err := appTest.NewTestAppWithErr(t, opts...)
		if stop != nil {
			defer stop()
		}

		require.Error(t, err, "Duplicate API registration should fail")
		// Error is wrapped by fx, check error message contains duplicate info
		assert.Contains(t, err.Error(), "duplicate api definition",
			"Error should mention duplicate API definition")
		assert.Contains(t, err.Error(), `resource="test/duplicate"`,
			"Error should include resource name")
		assert.Contains(t, err.Error(), `action="test_action"`,
			"Error should include action name")
	})

	t.Run("DetectDuplicateAcrossResources", func(t *testing.T) {
		resource1 := &DuplicateActionResource{
			Resource: apiPkg.NewResource(
				"test/conflict",
				apiPkg.WithApis(
					apiPkg.Spec{Action: "shared_action"},
				),
			),
		}

		resource2 := &DuplicateActionResource{
			Resource: apiPkg.NewResource(
				"test/conflict", // ❌ Same resource name!
				apiPkg.WithApis(
					apiPkg.Spec{Action: "shared_action"}, // ❌ Same action!
				),
			),
		}

		opts := []fx.Option{
			vef.ProvideApiResource(func() apiPkg.Resource {
				return resource1
			}),
			vef.ProvideApiResource(func() apiPkg.Resource {
				return resource2
			}),
			fx.Replace(&config.DatasourceConfig{
				Type: constants.DbSQLite,
			}),
		}

		_, stop, err := appTest.NewTestAppWithErr(t, opts...)
		if stop != nil {
			defer stop()
		}

		require.Error(t, err, "Duplicate API across resources should fail")
		// Error is wrapped by fx, check error message contains duplicate info
		assert.Contains(t, err.Error(), "duplicate api definition",
			"Error should mention duplicate API definition")
		assert.Contains(t, err.Error(), `resource="test/conflict"`,
			"Error should include resource name")
		assert.Contains(t, err.Error(), `action="shared_action"`,
			"Error should include action name")
	})

	t.Run("AllowDifferentVersions", func(t *testing.T) {
		resource := &DuplicateActionResource{
			Resource: apiPkg.NewResource(
				"test/versioned",
				apiPkg.WithApis(
					apiPkg.Spec{Action: "test_action", Version: "v1"},
					apiPkg.Spec{Action: "test_action", Version: "v2"}, // ✓ Different version - OK
				),
			),
		}

		opts := []fx.Option{
			vef.ProvideApiResource(func() apiPkg.Resource {
				return resource
			}),
			fx.Replace(&config.DatasourceConfig{
				Type: constants.DbSQLite,
			}),
		}

		_, stop, err := appTest.NewTestAppWithErr(t, opts...)
		if stop != nil {
			defer stop()
		}

		assert.NoError(t, err, "Different versions of same action should be allowed")
	})

	t.Run("DetectSystemApiOverride", func(t *testing.T) {
		resource := &DuplicateActionResource{
			Resource: apiPkg.NewResource(
				"security/auth", // ❌ System resource!
				apiPkg.WithApis(
					apiPkg.Spec{Action: "login"}, // ❌ System action!
				),
			),
		}

		opts := []fx.Option{
			vef.ProvideApiResource(func() apiPkg.Resource {
				return resource
			}),
			fx.Replace(&config.DatasourceConfig{
				Type: constants.DbSQLite,
			}),
		}

		_, stop, err := appTest.NewTestAppWithErr(t, opts...)
		if stop != nil {
			defer stop()
		}

		require.Error(t, err, "Overriding system authentication API should fail")
		// Error is wrapped by fx, check error message contains duplicate info
		assert.Contains(t, err.Error(), "duplicate api definition",
			"Error should mention duplicate API definition")
		assert.Contains(t, err.Error(), `resource="security/auth"`,
			"Error should include system resource name")
		assert.Contains(t, err.Error(), `action="login"`,
			"Error should include system action name")
	})

	t.Run("DetectStorageApiOverride", func(t *testing.T) {
		resource := &DuplicateActionResource{
			Resource: apiPkg.NewResource(
				"sys/storage", // ❌ System resource!
				apiPkg.WithApis(
					apiPkg.Spec{Action: "upload"}, // ❌ System action!
				),
			),
		}

		opts := []fx.Option{
			vef.ProvideApiResource(func() apiPkg.Resource {
				return resource
			}),
			fx.Replace(&config.DatasourceConfig{
				Type: constants.DbSQLite,
			}),
		}

		_, stop, err := appTest.NewTestAppWithErr(t, opts...)
		if stop != nil {
			defer stop()
		}

		require.Error(t, err, "Overriding system storage API should fail")
		// Error is wrapped by fx, check error message contains duplicate info
		assert.Contains(t, err.Error(), "duplicate api definition",
			"Error should mention duplicate API definition")
		assert.Contains(t, err.Error(), `resource="sys/storage"`,
			"Error should include system resource name")
		assert.Contains(t, err.Error(), `action="upload"`,
			"Error should include system action name")
	})
}

// DuplicateActionResource is a test resource used for duplicate detection tests.
type DuplicateActionResource struct {
	apiPkg.Resource
}

// TestAction is a placeholder handler.
func (r *DuplicateActionResource) TestAction() error {
	return nil
}

// SharedAction is a placeholder handler.
func (r *DuplicateActionResource) SharedAction() error {
	return nil
}

// Login is a placeholder handler for system Api override tests.
func (r *DuplicateActionResource) Login() error {
	return nil
}

// Upload is a placeholder handler for system Api override tests.
func (r *DuplicateActionResource) Upload() error {
	return nil
}
