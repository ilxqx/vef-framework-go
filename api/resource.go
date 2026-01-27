package api

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ilxqx/go-collections"

	"github.com/ilxqx/vef-framework-go/constants"
)

type Kind uint8

const (
	// KindRPC represents the KindRPC kind.
	KindRPC Kind = iota + 1
	// KindREST represents the KindREST kind.
	KindREST
)

func (k Kind) String() string {
	switch k {
	case KindRPC:
		return "rpc"
	case KindREST:
		return "rest"
	default:
		return "unknown"
	}
}

var (
	versionPattern          = regexp.MustCompile(`^v\d+$`)
	snakeCasePattern        = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)
	resourceNamePattern     = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*(/[a-z][a-z0-9]*(_[a-z0-9]+)*)*$`)
	restResourceNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*(/[a-z][a-z0-9]*(-[a-z0-9]+)*)*$`)

	validHTTPVerbs = collections.NewHashSetFrom(
		"get",
		"post",
		"put",
		"delete",
		"patch",
		"head",
		"options",
		"trace",
		"connect",
		"all",
	)
)

// ValidateActionName validates the action name based on the resource kind.
// For RPC, action must be snake_case (e.g., create, find_page).
// For REST, action format is "<method>" or "<method> <sub-resource>" (e.g., "get", "post user-friends").
func ValidateActionName(action string, kind Kind) error {
	if action == constants.Empty {
		return ErrEmptyActionName
	}

	switch kind {
	case KindRPC:
		return validateRPCAction(action)
	case KindREST:
		return validateRESTAction(action)
	default:
		return fmt.Errorf("%w: invalid resource kind %d", ErrInvalidResourceKind, kind)
	}
}

func validateRPCAction(action string) error {
	if !snakeCasePattern.MatchString(action) {
		return fmt.Errorf("%w (e.g., create, find_page, get_user_info): %q", ErrInvalidActionName, action)
	}

	return nil
}

func validateRESTAction(action string) error {
	parts := strings.SplitN(action, constants.Space, 3)
	if len(parts) > 2 {
		return fmt.Errorf("%w (e.g., get, post, get user-friends): %q", ErrInvalidActionName, action)
	}

	verb := parts[0]
	if !validHTTPVerbs.Contains(verb) {
		return fmt.Errorf("%w (invalid verb %q): %q", ErrInvalidActionName, verb, action)
	}

	if len(parts) == 2 {
		subRes := parts[1]
		if subRes == constants.Empty || !restResourceNamePattern.MatchString(subRes) {
			return fmt.Errorf("%w (invalid sub-resource %q): %q", ErrInvalidActionName, subRes, action)
		}
	}

	return nil
}

// baseResource provides a basic implementation of the Resource interface.
type baseResource struct {
	kind       Kind
	name       string
	version    string
	auth       *AuthConfig
	operations []OperationSpec
}

func (r *baseResource) validate() error {
	if err := r.validateVersion(); err != nil {
		return err
	}

	if err := r.validateName(); err != nil {
		return err
	}

	return r.validateOperations()
}

func (r *baseResource) validateVersion() error {
	if r.version != constants.Empty && !versionPattern.MatchString(r.version) {
		return fmt.Errorf("%w: %q", ErrInvalidVersionFormat, r.version)
	}

	return nil
}

func (r *baseResource) validateName() error {
	if r.name == constants.Empty {
		return ErrEmptyResourceName
	}

	if strings.HasPrefix(r.name, constants.Slash) || strings.HasSuffix(r.name, constants.Slash) {
		return fmt.Errorf("%w: %q", ErrResourceNameSlash, r.name)
	}

	if strings.Contains(r.name, constants.DoubleSlash) {
		return fmt.Errorf("%w: %q", ErrResourceNameDoubleSlash, r.name)
	}

	switch r.kind {
	case KindRPC:
		if !resourceNamePattern.MatchString(r.name) {
			return fmt.Errorf("%w (RPC e.g., user, sys/user, sys/data_dict): %q", ErrInvalidResourceName, r.name)
		}
	case KindREST:
		if !restResourceNamePattern.MatchString(r.name) {
			return fmt.Errorf("%w (REST e.g., user, sys/user, sys/data-dict): %q", ErrInvalidResourceName, r.name)
		}
	default:
		return fmt.Errorf("%w: invalid resource kind %d", ErrInvalidResourceKind, r.kind)
	}

	return nil
}

func (r *baseResource) validateOperations() error {
	for i, op := range r.operations {
		if op.Action == constants.Empty {
			return fmt.Errorf("%w: operation index %d", ErrEmptyActionName, i)
		}

		if err := ValidateActionName(op.Action, r.kind); err != nil {
			return err
		}
	}

	return nil
}

// NewRESTResource creates a new REST resource with the given name and options.
func NewRESTResource(name string, opts ...ResourceOption) Resource {
	return newResource(KindREST, name, opts...)
}

// NewRPCResource creates a new baseResource with the given name and options.
func NewRPCResource(name string, opts ...ResourceOption) Resource {
	return newResource(KindRPC, name, opts...)
}

func newResource(kind Kind, name string, opts ...ResourceOption) Resource {
	r := baseResource{
		kind: kind,
		name: name,
	}
	for _, opt := range opts {
		opt(&r)
	}

	if err := r.validate(); err != nil {
		panic(err)
	}

	return r
}

// Kind returns the resource kind.
func (r baseResource) Kind() Kind { return r.kind }

// Name returns the resource name.
func (r baseResource) Name() string { return r.name }

// Version returns the resource version.
func (r baseResource) Version() string { return r.version }

// Auth returns the resource authentication configuration.
func (r baseResource) Auth() *AuthConfig { return r.auth }

// Operations returns the resource operations.
func (r baseResource) Operations() []OperationSpec { return r.operations }

// ResourceOption configures a baseResource.
type ResourceOption func(*baseResource)

// WithVersion sets the resource version.
func WithVersion(v string) ResourceOption {
	return func(r *baseResource) {
		r.version = v
	}
}

// WithOperations sets the resource operations.
func WithOperations(ops ...OperationSpec) ResourceOption {
	return func(r *baseResource) {
		r.operations = ops
	}
}

// WithAuth sets the resource authentication configuration.
func WithAuth(auth *AuthConfig) ResourceOption {
	return func(r *baseResource) {
		r.auth = auth
	}
}
