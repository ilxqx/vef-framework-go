package api

type baseResource struct {
	version string
	name    string
	apis    []Config
}

func (b *baseResource) Version() string {
	return b.version
}

func (b *baseResource) Name() string {
	return b.name
}

func (b *baseResource) APIs() []Config {
	return b.apis
}

// NewResource creates a new resource with the given name and optional configuration.
// It initializes the resource with version v1 by default and applies any provided options.
func NewResource(name string, opts ...resourceOption) Resource {
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
func WithVersion(version string) resourceOption {
	return func(r *baseResource) {
		r.version = version
	}
}

// WithAPIs configures the API endpoints for the resource.
// It accepts a variadic list of Config objects that define the available APIs.
func WithAPIs(apis ...Config) resourceOption {
	return func(r *baseResource) {
		r.apis = apis
	}
}
