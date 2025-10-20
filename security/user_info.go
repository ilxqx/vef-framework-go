package security

import "github.com/guregu/null/v6"

// Gender represents the gender of the user.
type Gender string

const (
	// GenderMale represents the male gender.
	GenderMale Gender = "male"
	// GenderFemale represents the female gender.
	GenderFemale Gender = "female"
	// GenderUnknown represents an unknown gender.
	GenderUnknown Gender = "unknown"
)

// UserMenuType represents the type of the menu.
type UserMenuType string

const (
	// UserMenuTypeDirectory represents a directory menu type.
	UserMenuTypeDirectory UserMenuType = "directory"
	// UserMenuTypeMenu represents a menu type.
	UserMenuTypeMenu UserMenuType = "menu"
	// UserMenuTypeView represents a view menu type.
	UserMenuTypeView UserMenuType = "view"
	// UserMenuTypeDashboard represents a dashboard menu type.
	UserMenuTypeDashboard UserMenuType = "dashboard"
	// UserMenuTypeReport represents a report menu type.
	UserMenuTypeReport UserMenuType = "report"
)

// UserMenu represents a menu item in the user's navigation.
type UserMenu struct {
	// Type of the menu
	Type UserMenuType `json:"type"`
	// Path of the menu
	Path string `json:"path"`
	// Name of the menu
	Name string `json:"name"`
	// Icon of the menu (optional)
	Icon null.String `json:"icon"`
	// Meta of the menu (optional)
	Meta map[string]any `json:"metadata"`
	// Children of the menu (optional)
	Children []UserMenu `json:"children"`
}

// UserInfo represents detailed information about the authenticated user.
type UserInfo struct {
	// Id of the user
	Id string `json:"id"`
	// Name of the user
	Name string `json:"name"`
	// Gender of the user
	Gender Gender `json:"gender"`
	// Avatar URL of the user (optional)
	Avatar null.String `json:"avatar"`
	// Authorized permission tokens
	PermTokens []string `json:"permTokens"`
	// Authorized menus
	Menus []UserMenu `json:"menus"`
}
