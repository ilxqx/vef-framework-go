package apis_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilxqx/vef-framework-go/apis"
	"github.com/ilxqx/vef-framework-go/orm"
)

// TestAction tests the Action method validation for API builders.
func TestAction(t *testing.T) {
	t.Run("ValidActionNames", func(t *testing.T) {
		validActions := []struct {
			name   string
			action string
		}{
			{"SingleWord", "create"},
			{"SnakeCase", "find_page"},
			{"MultipleWords", "get_user_info"},
			{"TwoWords", "create_user"},
			{"WithMany", "delete_many"},
		}

		for _, tc := range validActions {
			t.Run(tc.name, func(t *testing.T) {
				assert.NotPanics(t, func() {
					_ = apis.NewCreate[orm.Model, orm.Model]().Action(tc.action)
				}, "Should accept valid snake_case action name")
			})
		}
	})

	t.Run("InvalidActionNames", func(t *testing.T) {
		invalidActions := []struct {
			name   string
			action string
		}{
			{"EmptyString", ""},
			{"CamelCase", "findPage"},
			{"PascalCase", "CreateUser"},
			{"StartsWithNumber", "1_create"},
			{"DoubleUnderscore", "create__user"},
			{"TrailingUnderscore", "create_user_"},
			{"LeadingUnderscore", "_create"},
			{"ContainsUppercase", "create_User"},
			{"ContainsHyphen", "create-user"},
			{"ContainsSpace", "create user"},
			{"ContainsDot", "create.user"},
			{"ContainsSlash", "create/user"},
		}

		for _, tc := range invalidActions {
			t.Run(tc.name, func(t *testing.T) {
				assert.Panics(t, func() {
					_ = apis.NewCreate[orm.Model, orm.Model]().Action(tc.action)
				}, "Should panic for invalid action name format")
			})
		}
	})
}
