package apis_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilxqx/vef-framework-go/apis"
	"github.com/ilxqx/vef-framework-go/orm"
)

func TestAction_ValidActionNames(t *testing.T) {
	validActions := []string{
		"create",
		"find_page",
		"get_user_info",
		"create_user",
		"delete_many",
	}

	for _, action := range validActions {
		t.Run(action, func(t *testing.T) {
			assert.NotPanics(t, func() {
				_ = apis.NewCreateApi[orm.Model, orm.Model]().Action(action)
			}, "Should accept valid action name: %s", action)
		})
	}
}

func TestAction_InvalidActionNames(t *testing.T) {
	invalidActions := []struct {
		name   string
		action string
	}{
		{"empty", ""},
		{"camelCase", "findPage"},
		{"PascalCase", "CreateUser"},
		{"starts_with_number", "1_create"},
		{"double_underscore", "create__user"},
		{"trailing_underscore", "create_user_"},
		{"leading_underscore", "_create"},
		{"contains_uppercase", "create_User"},
		{"contains_hyphen", "create-user"},
		{"contains_space", "create user"},
		{"contains_dot", "create.user"},
		{"contains_slash", "create/user"},
	}

	for _, tc := range invalidActions {
		t.Run(tc.name, func(t *testing.T) {
			assert.Panics(t, func() {
				_ = apis.NewCreateApi[orm.Model, orm.Model]().Action(tc.action)
			}, "Should panic for invalid action name (%s): %s", tc.name, tc.action)
		})
	}
}
