package api

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
)

var (
	versionPattern      = regexp.MustCompile(`^v\d+$`)
	snakeCasePattern    = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)
	resourceNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*(/[a-z][a-z0-9]*(_[a-z0-9]+)*)*$`)
)

// ValidateResourceName checks if the resource name follows the correct format.
// Resource names must be in snake_case and can contain slashes for namespacing.
// Valid examples: "user", "sys/user", "auth/get_user_info"
// Invalid examples: "User", "sys/User", "sys/getUserInfo", "sys/".
func ValidateResourceName(name string) error {
	if name == constants.Empty {
		return ErrResourceNameEmpty
	}

	if !resourceNamePattern.MatchString(name) {
		return fmt.Errorf("%w (e.g., user, sys/user, auth/get_user_info): %q", ErrResourceNameInvalidFormat, name)
	}

	// Additional check: no trailing or leading slashes, no consecutive slashes
	if strings.HasPrefix(name, constants.Slash) || strings.HasSuffix(name, constants.Slash) {
		return fmt.Errorf("%w: %q", ErrResourceNameInvalidSlash, name)
	}

	if strings.Contains(name, constants.DoubleSlash) {
		return fmt.Errorf("%w: %q", ErrResourceNameConsecutiveSlashes, name)
	}

	return nil
}

// ValidateActionName checks if the action name follows snake_case format.
// Valid examples: "create", "find_page", "get_user_info"
// Invalid examples: "Create", "findPage", "getUserInfo".
func ValidateActionName(action string) error {
	if action == constants.Empty {
		return ErrActionNameEmpty
	}

	if !snakeCasePattern.MatchString(action) {
		return fmt.Errorf("%w (e.g., create, find_page, get_user_info): %q", ErrActionNameInvalidFormat, action)
	}

	return nil
}

type baseResource struct {
	version string
	name    string
	apis    []Spec
}

func (b *baseResource) Version() string {
	return b.version
}

func (b *baseResource) Name() string {
	return b.name
}

func (b *baseResource) Apis() []Spec {
	return b.apis
}

// NewResource creates a new resource with the given name and optional configuration.
// It initializes the resource with version v1 by default and applies any provided options.
// The resource name must be in snake_case format and can contain slashes for namespacing.
// Panics if the resource name format is invalid.
func NewResource(name string, opts ...resourceOption) Resource {
	if err := ValidateResourceName(name); err != nil {
		panic(fmt.Sprintf("api: %v", err))
	}

	resource := &baseResource{
		version: VersionV1,
		name:    name,
	}

	for _, opt := range opts {
		opt(resource)
	}

	return resource
}

type resourceOption func(*baseResource)

// WithVersion sets the version for the resource.
// This option allows overriding the default v1 version with a custom version string.
// The version must match the pattern "v" followed by one or more digits (e.g., v1, v2, v10).
// Panics if the version format is invalid.
func WithVersion(version string) resourceOption {
	return func(r *baseResource) {
		if !versionPattern.MatchString(version) {
			panic(fmt.Sprintf("api: invalid version format %q, must match pattern v+digits (e.g., v1, v2, v10)", version))
		}

		r.version = version
	}
}

// WithApis configures the Api endpoints for the resource.
// It accepts a variadic list of Spec objects that define the available Apis.
// All action names in the Spec objects must be in snake_case format.
// Panics if any action name format is invalid.
func WithApis(apis ...Spec) resourceOption {
	return func(r *baseResource) {
		// Validate all action names
		for i, spec := range apis {
			if err := ValidateActionName(spec.Action); err != nil {
				panic(fmt.Sprintf("api: invalid action at index %d: %v", i, err))
			}
		}

		r.apis = apis
	}
}
