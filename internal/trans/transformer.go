package trans

import (
	"github.com/go-playground/mold/v4"
	"github.com/ilxqx/vef-framework-go/trans"
)

// newTransformer creates a new transformer instance, integrating all registered transformers and interceptors
// Uses dependency injection to collect all extension components and build a complete transformer
func newTransformer(fieldTransformers []trans.FieldTransformer, structTransformers []trans.StructTransformer, interceptors []trans.Interceptor) trans.Transformer {
	// Create mold transformer instance and set tag name to "trans"
	transformer := mold.New()
	transformer.SetTagName("trans")

	// Register all field-level transformers, each transformer corresponds to a tag
	for _, ft := range fieldTransformers {
		transformer.Register(ft.Tag(), ft.Transform)
	}

	// Register all struct-level transformers for handling entire struct transformation logic
	for _, st := range structTransformers {
		transformer.RegisterStructLevel(st.Transform)
	}

	// Register all interceptors for handling special type value redirection
	for _, interceptor := range interceptors {
		transformer.RegisterInterceptor(interceptor.Intercept)
	}

	return transformer
}
