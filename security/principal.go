package security

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/mapx"
)

// PrincipalType is the type of the principal.
type PrincipalType string

const (
	// PrincipalTypeUser is the type of the user.
	PrincipalTypeUser PrincipalType = "user"
	// PrincipalTypeExternalApp is the type of the external app.
	PrincipalTypeExternalApp PrincipalType = "external_app"
	// PrincipalTypeSystem is the type of the system.
	PrincipalTypeSystem PrincipalType = constants.OperatorSystem
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

func SetUserDetailsType[T any]() {
	userDetailsType = reflect.TypeFor[T]()
	if userDetailsType.Kind() != reflect.Struct {
		panic(
			fmt.Errorf("%w, got %s", ErrUserDetailsTypeMustBeStruct, userDetailsType.Name()),
		)
	}
}

func SetExternalAppDetailsType[T any]() {
	externalAppDetailsType = reflect.TypeFor[T]()
	if externalAppDetailsType.Kind() != reflect.Struct {
		panic(
			fmt.Errorf("%w, got %s", ErrExternalAppDetailsTypeMustBeStruct, externalAppDetailsType.Name()),
		)
	}
}

// Principal is the principal of the user.
type Principal struct {
	// Type is the type of the principal.
	Type PrincipalType `json:"type"`
	// Id is the id of the user.
	Id string `json:"id"`
	// Name is the name of the user.
	Name string `json:"name"`
	// Roles is the roles of the user.
	Roles []string `json:"roles"`
	// Details is the details of the user.
	Details any `json:"details"`
}

// UnmarshalJSON implements custom JSON unmarshaling for Principal.
// This allows the Details field to be properly deserialized based on the Type field.
func (p *Principal) UnmarshalJSON(data []byte) error {
	// Define a temporary struct to unmarshal the JSON
	type Alias Principal

	aux := &struct {
		*Alias

		Details json.RawMessage `json:"details"`
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
			userDetails := reflect.New(userDetailsType).Interface()
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
			appDetails := reflect.New(externalAppDetailsType).Interface()
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
