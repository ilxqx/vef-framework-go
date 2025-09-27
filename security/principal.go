package security

import (
	"fmt"
	"reflect"

	"github.com/goccy/go-json"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/mapx"
	"github.com/samber/lo"
)

// PrincipalType is the type of the principal.
type PrincipalType string

const (
	PrincipalTypeUser        PrincipalType = "user"                   // PrincipalTypeUser is the type of the user.
	PrincipalTypeExternalApp PrincipalType = "external_app"           // PrincipalTypeExternalApp is the type of the external app.
	PrincipalTypeSystem      PrincipalType = constants.OperatorSystem // PrincipalTypeSystem is the type of the system.
)

var (
	PrincipalSystem = &Principal{
		Type: PrincipalTypeSystem,
		Id:   constants.OperatorSystem,
		Name: "系统",
	}
	PrincipalAnonymous = NewUser(constants.OperatorAnonymous, "匿名")

	userDetailsType        = reflect.TypeFor[map[string]any]()
	externalAppDetailsType = reflect.TypeFor[map[string]any]()
)

// SetUserDetailsType sets the type of the user details.
func SetUserDetailsType[T any]() {
	userDetailsType = reflect.TypeFor[T]()
	if userDetailsType.Kind() != reflect.Struct {
		panic(
			fmt.Errorf("user details type must be a struct, got %s", userDetailsType.Name()),
		)
	}
}

// SetExternalAppDetailsType sets the type of the external app details.
func SetExternalAppDetailsType[T any]() {
	externalAppDetailsType = reflect.TypeFor[T]()
	if externalAppDetailsType.Kind() != reflect.Struct {
		panic(
			fmt.Errorf("external app details type must be a struct, got %s", externalAppDetailsType.Name()),
		)
	}
}

// Principal is the principal of the user.
type Principal struct {
	Type    PrincipalType `json:"type"`    // Type is the type of the principal.
	Id      string        `json:"id"`      // Id is the id of the user.
	Name    string        `json:"name"`    // Name is the name of the user.
	Roles   []string      `json:"roles"`   // Roles is the roles of the user.
	Details any           `json:"details"` // Details is the details of the user.
}

// UnmarshalJSON implements custom JSON unmarshaling for Principal.
// This allows the Details field to be properly deserialized based on the Type field.
func (p *Principal) UnmarshalJSON(data []byte) error {
	// Define a temporary struct to unmarshal the JSON
	type Alias Principal
	aux := &struct {
		Details json.RawMessage `json:"details"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle Details based on Type
	if aux.Details != nil {
		switch p.Type {
		case PrincipalTypeUser:
			// For user type, Details could be a UserDetails struct
			var userDetails = reflect.New(userDetailsType).Interface()
			if err := json.Unmarshal(aux.Details, &userDetails); err != nil {
				// If it fails, try to unmarshal as a generic map
				var detailsMap map[string]any
				if err := json.Unmarshal(aux.Details, &detailsMap); err != nil {
					return fmt.Errorf("failed to unmarshal user details: %w", err)
				}
				p.Details = detailsMap
			} else {
				p.Details = userDetails
			}
		case PrincipalTypeExternalApp:
			// For external app type, Details could be an ExternalAppDetails struct
			var appDetails = reflect.New(externalAppDetailsType).Interface()
			if err := json.Unmarshal(aux.Details, &appDetails); err != nil {
				// If it fails, try to unmarshal as a generic map
				var detailsMap map[string]any
				if err := json.Unmarshal(aux.Details, &detailsMap); err != nil {
					return fmt.Errorf("failed to unmarshal external app details: %w", err)
				}
				p.Details = detailsMap
			} else {
				p.Details = appDetails
			}
		case PrincipalTypeSystem:
			// For system type, Details is usually nil or empty
			p.Details = nil
		default:
			// For unknown types, unmarshal as a generic map
			var detailsMap map[string]any
			if err := json.Unmarshal(aux.Details, &detailsMap); err != nil {
				return fmt.Errorf("failed to unmarshal details for unknown type %s: %w", p.Type, err)
			}
			p.Details = detailsMap
		}
	}

	return nil
}

// AttemptUnmarshalDetails attempts to unmarshal the details into the principal.
func (p *Principal) AttemptUnmarshalDetails(details any) {
	if p.Type != PrincipalTypeUser && p.Type != PrincipalTypeExternalApp {
		p.Details = details
		return
	}

	detailsType := lo.Ternary(p.Type == PrincipalTypeUser, userDetailsType, externalAppDetailsType)
	// If details is not a map, return it
	if _, ok := details.(map[string]any); !ok || detailsType.AssignableTo(reflect.TypeFor[map[string]any]()) {
		p.Details = details
		return
	}

	value := reflect.New(detailsType).Interface()
	decoder, err := mapx.NewDecoder(value)
	if err != nil {
		p.Details = details
		return
	}

	if err := decoder.Decode(details); err != nil {
		p.Details = details
		return
	}

	p.Details = value
}

// WithRoles adds roles to the principal.
func (p *Principal) WithRoles(roles ...string) *Principal {
	p.Roles = append(p.Roles, roles...)
	return p
}

// NewUser is the function to create a new user principal.
func NewUser(id, name string, roles ...string) *Principal {
	return &Principal{
		Type:  PrincipalTypeUser,
		Id:    id,
		Name:  name,
		Roles: roles,
	}
}

// NewExternalApp is the function to create a new external app principal.
func NewExternalApp(id, name string, roles ...string) *Principal {
	return &Principal{
		Type:  PrincipalTypeExternalApp,
		Id:    id,
		Name:  name,
		Roles: roles,
	}
}
