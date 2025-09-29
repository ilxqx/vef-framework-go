package mold

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/mold"
)

var (
	timeType           = reflect.TypeFor[time.Time]()
	restrictedAliasErr = "mold: alias '%s' either contains restricted characters or is the same as a restricted tag needed for normal operation"
	restrictedTagErr   = "mold: tag '%s' either contains restricted characters or is the same as a restricted tag needed for normal operation"
)

// MoldTransformer is the base controlling object which contains
// all necessary information
type MoldTransformer struct {
	tagName          string
	aliases          map[string]string
	transformations  map[string]mold.Func
	structLevelFuncs map[reflect.Type]mold.StructLevelFunc
	interceptors     map[reflect.Type]mold.InterceptorFunc
	cCache           *structCache
	tCache           *tagCache
}

// New creates a new Transform object with default tag name of 'mold'
func New() *MoldTransformer {
	tc := new(tagCache)
	tc.m.Store(make(map[string]*cTag))

	sc := new(structCache)
	sc.m.Store(make(map[reflect.Type]*cStruct))

	return &MoldTransformer{
		tagName:         "mold",
		aliases:         make(map[string]string),
		transformations: make(map[string]mold.Func),
		interceptors:    make(map[reflect.Type]mold.InterceptorFunc),
		cCache:          sc,
		tCache:          tc,
	}
}

// Register adds a transformation with the given tag
//
// NOTES:
// - if the key already exists, the previous transformation function will be replaced.
// - this method is not thread-safe it is intended that these all be registered before hand
func (t *MoldTransformer) Register(tag string, fn mold.Func) {
	if len(tag) == 0 {
		panic("mold: transformation tag cannot be empty")
	}

	if fn == nil {
		panic("mold: transformation function cannot be nil")
	}

	_, ok := restrictedTags[tag]

	if ok || strings.ContainsAny(tag, restrictedTagChars) {
		panic(fmt.Sprintf(restrictedTagErr, tag))
	}
	t.transformations[tag] = fn
}

// RegisterAlias registers a mapping of a single transform tag that
// defines a common or complex set of transformations to simplify adding transforms
// to structs.
//
// NOTE: this function is not thread-safe it is intended that these all be registered before hand
func (t *MoldTransformer) RegisterAlias(alias, tags string) {
	if len(alias) == 0 {
		panic("mold: transformation alias cannot be empty")
	}

	if len(tags) == 0 {
		panic("mold: aliased tags cannot be empty")
	}

	_, ok := restrictedTags[alias]

	if ok || strings.ContainsAny(alias, restrictedTagChars) {
		panic(fmt.Sprintf(restrictedAliasErr, alias))
	}
	t.aliases[alias] = tags
}

// RegisterStructLevel registers a StructLevelFunc against a number of types.
// Why does this exist? For structs for which you may not have access or rights to add tags too,
// from other packages your using.
//
// NOTES:
// - this method is not thread-safe it is intended that these all be registered prior to any validation
func (t *MoldTransformer) RegisterStructLevel(fn mold.StructLevelFunc, types ...any) {
	if t.structLevelFuncs == nil {
		t.structLevelFuncs = make(map[reflect.Type]mold.StructLevelFunc)
	}

	for _, typ := range types {
		t.structLevelFuncs[reflect.TypeOf(typ)] = fn
	}
}

// RegisterInterceptor registers a new interceptor functions agains one or more types.
// This InterceptorFunc allows one to intercept the incoming to to redirect the application of modifications
// to an inner type/value.
//
// eg. sql.NullString
func (t *MoldTransformer) RegisterInterceptor(fn mold.InterceptorFunc, types ...any) {
	for _, typ := range types {
		t.interceptors[reflect.TypeOf(typ)] = fn
	}
}

// Struct applies transformations against the provided struct
func (t *MoldTransformer) Struct(ctx context.Context, v any) error {
	orig := reflect.ValueOf(v)

	if orig.Kind() != reflect.Ptr || orig.IsNil() {
		return &ErrInvalidTransformValue{typ: reflect.TypeOf(v), fn: "Struct"}
	}

	val := orig.Elem()
	typ := val.Type()

	if val.Kind() != reflect.Struct || val.Type() == timeType {
		return &ErrInvalidTransformation{typ: reflect.TypeOf(v)}
	}
	return t.setByStruct(ctx, orig, val, typ)
}

func (t *MoldTransformer) setByStruct(ctx context.Context, parent, current reflect.Value, typ reflect.Type) (err error) {
	cs, ok := t.cCache.Get(typ)
	if !ok {
		if cs, err = t.extractStructCache(current); err != nil {
			return
		}
	}

	// run is struct has a corresponding struct level transformation
	if cs.fn != nil {
		if err = cs.fn(ctx, MoldStructLevel{
			transformer: t,
			parent:      parent,
			current:     current,
		}); err != nil {
			return
		}
	}

	for name, field := range cs.fields {
		if err = t.setByFieldWithContainer(ctx, name, current.Field(field.idx), field.cTags, current, cs); err != nil {
			return
		}
	}
	return nil
}

// Field applies the provided transformations against the variable
func (t *MoldTransformer) Field(ctx context.Context, v any, tags string) (err error) {
	if len(tags) == 0 || tags == ignoreTag {
		return nil
	}

	val := reflect.ValueOf(v)

	if val.Kind() != reflect.Pointer || val.IsNil() {
		return &ErrInvalidTransformValue{typ: reflect.TypeOf(v), fn: "Field"}
	}
	val = val.Elem()

	// find cached tag
	ctag, ok := t.tCache.Get(tags)
	if !ok {
		t.tCache.lock.Lock()

		// could have been multiple trying to access, but once first is done this ensures tag
		// isn't parsed again.
		ctag, ok = t.tCache.Get(tags)
		if !ok {
			if ctag, _, err = t.parseFieldTagsRecursive(tags, constants.Empty, constants.Empty, false); err != nil {
				t.tCache.lock.Unlock()
				return
			}
			t.tCache.Set(tags, ctag)
		}
		t.tCache.lock.Unlock()
	}
	err = t.setByField(ctx, val, ctag)
	return
}

func (t *MoldTransformer) setByFieldWithContainer(ctx context.Context, name string, original reflect.Value, ct *cTag, structValue reflect.Value, structCache *cStruct) (err error) {
	current, kind := t.extractType(original)

	if ct != nil && ct.hasTag {
		for ct != nil {
			switch ct.typeof {
			case typeEndKeys:
				return
			case typeDive:
				ct = ct.next

				switch kind {
				case reflect.Slice, reflect.Array:
					err = t.setByIterable(ctx, current, ct)
				case reflect.Map:
					err = t.setByMap(ctx, current, ct)
				case reflect.Pointer:
					innerKind := current.Type().Elem().Kind()
					if innerKind == reflect.Slice || innerKind == reflect.Map {
						// is a nil pointer to a slice or map, nothing to do.
						return nil
					}
					// not a valid use of the dive tag
					fallthrough
				default:
					err = ErrInvalidDive
				}
				return

			default:
				if !current.CanAddr() {
					newVal := reflect.New(current.Type()).Elem()
					newVal.Set(current)
					if err = ct.fn(ctx, MoldFieldLevel{
						transformer: t,
						name:        name,
						parent:      original,
						current:     newVal,
						param:       ct.param,
						container:   structValue,
						sc:          structCache,
					}); err != nil {
						return
					}
					original.Set(reflect.Indirect(newVal))
					current, kind = t.extractType(original)
				} else {
					if err = ct.fn(ctx, MoldFieldLevel{
						transformer: t,
						name:        name,
						parent:      original,
						current:     current,
						param:       ct.param,
						container:   structValue,
						sc:          structCache,
					}); err != nil {
						return
					}
					// value could have been changed or reassigned
					current, kind = t.extractType(current)
				}
				ct = ct.next
			}
		}
	}

	// need to do this again because one of the previous
	// sets could have set a struct value, where it was a
	// nil pointer before
	original2 := current
	current, kind = t.extractType(current)

	if kind == reflect.Struct {
		typ := current.Type()
		if typ == timeType {
			return
		}

		if !current.CanAddr() {
			newVal := reflect.New(typ).Elem()
			newVal.Set(current)

			if err = t.setByStruct(ctx, original, newVal, typ); err != nil {
				return
			}
			original.Set(reflect.Indirect(newVal))
			return
		}
		err = t.setByStruct(ctx, original2, current, typ)
	}
	return
}

func (t *MoldTransformer) setByField(ctx context.Context, original reflect.Value, ct *cTag) (err error) {
	current, kind := t.extractType(original)

	if ct != nil && ct.hasTag {
		for ct != nil {
			switch ct.typeof {
			case typeEndKeys:
				return
			case typeDive:
				ct = ct.next

				switch kind {
				case reflect.Slice, reflect.Array:
					err = t.setByIterable(ctx, current, ct)
				case reflect.Map:
					err = t.setByMap(ctx, current, ct)
				case reflect.Pointer:
					innerKind := current.Type().Elem().Kind()
					if innerKind == reflect.Slice || innerKind == reflect.Map {
						// is a nil pointer to a slice or map, nothing to do.
						return nil
					}
					// not a valid use of the dive tag
					fallthrough
				default:
					err = ErrInvalidDive
				}
				return

			default:
				if !current.CanAddr() {
					newVal := reflect.New(current.Type()).Elem()
					newVal.Set(current)
					if err = ct.fn(ctx, MoldFieldLevel{
						transformer: t,
						parent:      original,
						current:     newVal,
						param:       ct.param,
					}); err != nil {
						return
					}
					original.Set(reflect.Indirect(newVal))
					current, kind = t.extractType(original)
				} else {
					if err = ct.fn(ctx, MoldFieldLevel{
						transformer: t,
						parent:      original,
						current:     current,
						param:       ct.param,
					}); err != nil {
						return
					}
					// value could have been changed or reassigned
					current, kind = t.extractType(current)
				}
				ct = ct.next
			}
		}
	}

	// need to do this again because one of the previous
	// sets could have set a struct value, where it was a
	// nil pointer before
	original2 := current
	current, kind = t.extractType(current)

	if kind == reflect.Struct {
		typ := current.Type()
		if typ == timeType {
			return
		}

		if !current.CanAddr() {
			newVal := reflect.New(typ).Elem()
			newVal.Set(current)

			if err = t.setByStruct(ctx, original, newVal, typ); err != nil {
				return
			}
			original.Set(reflect.Indirect(newVal))
			return
		}
		err = t.setByStruct(ctx, original2, current, typ)
	}
	return
}

func (t *MoldTransformer) setByIterable(ctx context.Context, current reflect.Value, ct *cTag) (err error) {
	for i := 0; i < current.Len(); i++ {
		if err = t.setByField(ctx, current.Index(i), ct); err != nil {
			return
		}
	}
	return
}

func (t *MoldTransformer) setByMap(ctx context.Context, current reflect.Value, ct *cTag) error {
	for _, key := range current.MapKeys() {
		newVal := reflect.New(current.Type().Elem()).Elem()
		newVal.Set(current.MapIndex(key))

		if ct != nil && ct.typeof == typeKeys && ct.keys != nil {
			// remove current map key as we may be changing it
			// and re-add to the map afterwards
			current.SetMapIndex(key, reflect.Value{})

			newKey := reflect.New(current.Type().Key()).Elem()
			newKey.Set(key)
			key = newKey

			// handle map key
			if err := t.setByField(ctx, key, ct.keys); err != nil {
				return err
			}

			// can be nil when just keys being validated
			if ct.next != nil {
				if err := t.setByField(ctx, newVal, ct.next); err != nil {
					return err
				}
			}
		} else {
			if err := t.setByField(ctx, newVal, ct); err != nil {
				return err
			}
		}
		current.SetMapIndex(key, newVal)
	}

	return nil
}
