package copier

import "github.com/jinzhu/copier"

type copyOption func(option *copier.Option)

// WithIgnoreEmpty ignore empty fields
func WithIgnoreEmpty() copyOption {
	return func(option *copier.Option) {
		option.IgnoreEmpty = true
	}
}

// WithDeepCopy deep copy fields
func WithDeepCopy() copyOption {
	return func(option *copier.Option) {
		option.DeepCopy = true
	}
}

// WithCaseInsensitive case-insensitive
func WithCaseInsensitive() copyOption {
	return func(option *copier.Option) {
		option.CaseSensitive = false
	}
}

// WithFieldNameMapping field name mapping
func WithFieldNameMapping(fieldMapping ...copier.FieldNameMapping) copyOption {
	return func(option *copier.Option) {
		option.FieldNameMapping = append(option.FieldNameMapping, fieldMapping...)
	}
}

// Copy src to dst
func Copy(src any, dst any, options ...copyOption) error {
	option := copier.Option{
		CaseSensitive: true,
	}
	for _, opt := range options {
		opt(&option)
	}

	return copier.CopyWithOption(dst, src, option)
}
