package security

import "github.com/guregu/null/v6"

type Gender string

const (
	GenderMale    Gender = "male"
	GenderFemale  Gender = "female"
	GenderUnknown Gender = "unknown"
)

type UserMenuType string

const (
	UserMenuTypeDirectory UserMenuType = "directory"
	UserMenuTypeMenu      UserMenuType = "menu"
	UserMenuTypeView      UserMenuType = "view"
	UserMenuTypeDashboard UserMenuType = "dashboard"
	UserMenuTypeReport    UserMenuType = "report"
)

type UserMenu struct {
	Type     UserMenuType   `json:"type"`
	Path     string         `json:"path"`
	Name     string         `json:"name"`
	Icon     null.String    `json:"icon"`
	Meta     map[string]any `json:"metadata,omitempty"`
	Children []UserMenu     `json:"children,omitempty"`
}

type UserInfo struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Gender     Gender      `json:"gender"`
	Avatar     null.String `json:"avatar"`
	PermTokens []string    `json:"permTokens"`
	Menus      []UserMenu  `json:"menus"`
	Details    any         `json:"details,omitempty"`
}
