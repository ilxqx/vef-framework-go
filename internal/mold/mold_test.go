package mold

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilxqx/vef-framework-go/mold"
)

func TestBadValues(t *testing.T) {
	transformer := New()
	transformer.Register("blah", func(ctx context.Context, fl mold.FieldLevel) error { return nil })

	type InvalidTagStruct struct {
		Ignore string `mold:"-"`
		String string `mold:"blah,,blah"`
	}

	t.Run("invalid tag format", func(t *testing.T) {
		var tt InvalidTagStruct

		err := transformer.Struct(context.Background(), &tt)
		assert.Error(t, err)
		assert.Equal(t, "invalid tag '' found on field String", err.Error())
	})

	t.Run("non-pointer struct", func(t *testing.T) {
		var tt InvalidTagStruct

		err := transformer.Struct(context.Background(), tt)
		assert.Error(t, err)
		assert.Equal(t, "mold: Struct(non-pointer mold.InvalidTagStruct)", err.Error())
	})

	t.Run("nil struct", func(t *testing.T) {
		err := transformer.Struct(context.Background(), nil)
		assert.Error(t, err)
		assert.Equal(t, "mold: Struct(nil)", err.Error())
	})

	t.Run("invalid type - int", func(t *testing.T) {
		var i int

		err := transformer.Struct(context.Background(), &i)
		assert.Error(t, err)
		assert.Equal(t, "mold: (nil *int)", err.Error())
	})

	t.Run("interface nil", func(t *testing.T) {
		var iface any

		err := transformer.Struct(context.Background(), iface)
		assert.Error(t, err)
		assert.Equal(t, "mold: Struct(nil)", err.Error())
	})

	t.Run("interface pointer to nil", func(t *testing.T) {
		var iface any = nil

		err := transformer.Struct(context.Background(), &iface)
		assert.Error(t, err)
		assert.Equal(t, "mold: (nil *interface {})", err.Error())
	})

	t.Run("nil pointer to struct", func(t *testing.T) {
		var tst *InvalidTagStruct

		err := transformer.Struct(context.Background(), tst)
		assert.Error(t, err)
		assert.Equal(t, "mold: Struct(nil *mold.InvalidTagStruct)", err.Error())
	})

	t.Run("nil pointer field", func(t *testing.T) {
		var tm *time.Time

		err := transformer.Field(context.Background(), tm, "blah")
		assert.Error(t, err)
		assert.Equal(t, "mold: Field(nil *time.Time)", err.Error())
	})

	t.Run("registration panics", func(t *testing.T) {
		assert.PanicsWithValue(t, "mold: transformation tag cannot be empty", func() {
			transformer.Register("", nil)
		})
		assert.PanicsWithValue(t, "mold: transformation function cannot be nil", func() {
			transformer.Register("test", nil)
		})
		assert.PanicsWithValue(t, "mold: tag ',' either contains restricted characters or is the same as a restricted tag needed for normal operation", func() {
			transformer.Register(",", func(ctx context.Context, fl mold.FieldLevel) error { return nil })
		})
	})

	t.Run("alias registration panics", func(t *testing.T) {
		assert.PanicsWithValue(t, "mold: transformation alias cannot be empty", func() {
			transformer.RegisterAlias("", "")
		})
		assert.PanicsWithValue(t, "mold: aliased tags cannot be empty", func() {
			transformer.RegisterAlias("test", "")
		})
		assert.PanicsWithValue(t, "mold: alias ',' either contains restricted characters or is the same as a restricted tag needed for normal operation", func() {
			transformer.RegisterAlias(",", "test")
		})
	})
}

func TestBasicTransform(t *testing.T) {
	type BasicStruct struct {
		String string `mold:"repl"`
	}

	transformer := New()
	transformer.Register("repl", func(ctx context.Context, fl mold.FieldLevel) error {
		fl.Field().SetString("test")

		return nil
	})

	t.Run("basic struct transformation", func(t *testing.T) {
		var tt BasicStruct

		val := reflect.ValueOf(tt)
		// trigger a wait in struct parsing
		for range 3 {
			_, err := transformer.extractStructCache(val)
			require.NoError(t, err)
		}

		err := transformer.Struct(context.Background(), &tt)
		require.NoError(t, err)
		assert.Equal(t, "test", tt.String)
	})

	t.Run("nested struct transformation", func(t *testing.T) {
		type NestedStruct struct {
			Basic  BasicStruct
			String string `mold:"repl"`
		}

		var tt2 NestedStruct

		err := transformer.Struct(context.Background(), &tt2)
		require.NoError(t, err)
		assert.Equal(t, "test", tt2.Basic.String)
		assert.Equal(t, "test", tt2.String)
	})

	t.Run("embedded struct transformation", func(t *testing.T) {
		type EmbeddedStruct struct {
			BasicStruct

			String string `mold:"repl"`
		}

		var tt3 EmbeddedStruct

		err := transformer.Struct(context.Background(), &tt3)
		require.NoError(t, err)
		assert.Equal(t, "test", tt3.BasicStruct.String)
		assert.Equal(t, "test", tt3.String)
	})

	t.Run("nil pointer field", func(t *testing.T) {
		type NilPointerStruct struct {
			Basic  *BasicStruct
			String string `mold:"repl"`
		}

		var tt4 NilPointerStruct

		err := transformer.Struct(context.Background(), &tt4)
		require.NoError(t, err)
		assert.Nil(t, tt4.Basic)
		assert.Equal(t, "test", tt4.String)
	})

	t.Run("non-nil pointer field", func(t *testing.T) {
		type NonNilPointerStruct struct {
			Basic  *BasicStruct
			String string `mold:"repl"`
		}

		tt5 := NonNilPointerStruct{Basic: &BasicStruct{}}
		err := transformer.Struct(context.Background(), &tt5)
		require.NoError(t, err)
		assert.Equal(t, "test", tt5.Basic.String)
		assert.Equal(t, "test", tt5.String)
	})

	t.Run("default tag for pointer", func(t *testing.T) {
		type DefaultPointerStruct struct {
			Basic  *BasicStruct `mold:"default"`
			String string       `mold:"repl"`
		}

		var tt6 DefaultPointerStruct

		transformer.Register("default", func(ctx context.Context, fl mold.FieldLevel) error {
			fl.Field().Set(reflect.New(fl.Field().Type().Elem()))

			return nil
		})
		err := transformer.Struct(context.Background(), &tt6)
		require.NoError(t, err)
		assert.NotNil(t, tt6.Basic)
		assert.Equal(t, "test", tt6.Basic.String)
		assert.Equal(t, "test", tt6.String)
	})

	t.Run("field transformation", func(t *testing.T) {
		type FieldTransformStruct struct {
			Basic  *BasicStruct `mold:"default"`
			String string       `mold:"repl"`
		}

		tt6 := FieldTransformStruct{}
		tt6.String = "BAD"

		var tString string

		// wil invoke one processing and one waiting
		go func() {
			err := transformer.Field(context.Background(), &tString, "repl")
			assert.NoError(t, err)
		}()

		err := transformer.Field(context.Background(), &tt6.String, "repl")
		require.NoError(t, err)
		assert.Equal(t, "test", tt6.String)
	})

	t.Run("empty and skip tags", func(t *testing.T) {
		type EmptyTagStruct struct {
			String string `mold:"repl"`
		}

		tt6 := EmptyTagStruct{String: "BAD"}
		err := transformer.Field(context.Background(), &tt6.String, "")
		require.NoError(t, err)

		err = transformer.Field(context.Background(), &tt6.String, "-")
		require.NoError(t, err)
	})

	t.Run("field errors", func(t *testing.T) {
		type FieldErrorStruct struct {
			String string `mold:"repl"`
		}

		tt6 := FieldErrorStruct{}
		err := transformer.Field(context.Background(), tt6.String, "test")
		assert.Error(t, err)
		assert.Equal(t, "mold: Field(non-pointer string)", err.Error())

		err = transformer.Field(context.Background(), nil, "test")
		assert.Error(t, err)
		assert.Equal(t, "mold: Field(nil)", err.Error())

		var iface any

		err = transformer.Field(context.Background(), iface, "test")
		assert.Error(t, err)
		assert.Equal(t, "mold: Field(nil)", err.Error())
	})

	t.Run("nonexistent transformation", func(t *testing.T) {
		type NonexistentTransformStruct struct {
			String string `mold:"repl"`
		}

		tt6 := NonexistentTransformStruct{}

		var tString string

		done := make(chan struct{})

		go func() {
			err := transformer.Field(context.Background(), &tString, "nonexistant")
			assert.Error(t, err)
			close(done)
		}()

		err := transformer.Field(context.Background(), &tt6.String, "nonexistant")
		assert.Error(t, err)
		assert.Equal(t, "unregistered/undefined transformation 'nonexistant' found on field", err.Error())

		<-done
		transformer.Register("dummy", func(ctx context.Context, fl mold.FieldLevel) error { return nil })
		err = transformer.Field(context.Background(), &tt6.String, "dummy")
		assert.NoError(t, err)
	})
}

func TestAlias(t *testing.T) {
	transformer := New()
	transformer.Register("repl", func(ctx context.Context, fl mold.FieldLevel) error {
		fl.Field().SetString("test")

		return nil
	})
	transformer.Register("repl2", func(ctx context.Context, fl mold.FieldLevel) error {
		fl.Field().SetString("test2")

		return nil
	})

	t.Run("multiple transformations", func(t *testing.T) {
		type MultipleTransformStruct struct {
			String string `mold:"repl,repl2"`
		}

		var tt MultipleTransformStruct

		err := transformer.Struct(context.Background(), &tt)
		require.NoError(t, err)
		assert.Equal(t, "test2", tt.String)
	})

	t.Run("alias registration and usage", func(t *testing.T) {
		transformer.RegisterAlias("rep", "repl,repl2")
		transformer.RegisterAlias("bad", "repl,,repl2")

		type AliasStruct struct {
			String string `mold:"rep"`
		}

		var tt2 AliasStruct

		err := transformer.Struct(context.Background(), &tt2)
		require.NoError(t, err)
		assert.Equal(t, "test2", tt2.String)
	})

	t.Run("invalid alias with empty tag", func(t *testing.T) {
		var s string

		err := transformer.Field(context.Background(), &s, "bad")
		assert.Error(t, err)
	})

	t.Run("combined alias usage", func(t *testing.T) {
		var s string

		err := transformer.Field(context.Background(), &s, "repl,rep,bad")
		assert.Error(t, err)
	})
}

func TestArray(t *testing.T) {
	transformer := New()
	transformer.Register("defaultArr", func(ctx context.Context, fl mold.FieldLevel) error {
		if hasValue(fl.Field()) {
			return nil
		}

		fl.Field().Set(reflect.MakeSlice(fl.Field().Type(), 2, 2))

		return nil
	})
	transformer.Register("defaultStr", func(ctx context.Context, fl mold.FieldLevel) error {
		if fl.Field().String() == "ok" {
			return errors.New("ALREADY OK")
		}

		fl.Field().SetString("default")

		return nil
	})

	t.Run("default array creation and dive", func(t *testing.T) {
		type ArrayStruct struct {
			Arr []string `mold:"defaultArr,dive,defaultStr"`
		}

		var tt ArrayStruct

		err := transformer.Struct(context.Background(), &tt)
		require.NoError(t, err)
		assert.Len(t, tt.Arr, 2)
		assert.Equal(t, "default", tt.Arr[0])
		assert.Equal(t, "default", tt.Arr[1])
	})

	t.Run("existing array transformation", func(t *testing.T) {
		type ExistingArrayStruct struct {
			Arr []string `mold:"defaultArr,dive,defaultStr"`
		}

		tt2 := ExistingArrayStruct{
			Arr: make([]string, 1),
		}
		err := transformer.Struct(context.Background(), &tt2)
		require.NoError(t, err)
		assert.Len(t, tt2.Arr, 1)
		assert.Equal(t, "default", tt2.Arr[0])
	})

	t.Run("transformation error in array element", func(t *testing.T) {
		type ArrayErrorStruct struct {
			Arr []string `mold:"defaultArr,dive,defaultStr"`
		}

		tt3 := ArrayErrorStruct{
			Arr: []string{"ok"},
		}
		err := transformer.Struct(context.Background(), &tt3)
		assert.Error(t, err)
		assert.Equal(t, "ALREADY OK", err.Error())
	})
}

func TestMap(t *testing.T) {
	transformer := New()
	transformer.Register("defaultMap", func(ctx context.Context, fl mold.FieldLevel) error {
		if hasValue(fl.Field()) {
			return nil
		}

		fl.Field().Set(reflect.MakeMap(fl.Field().Type()))

		return nil
	})
	transformer.Register("defaultStr", func(ctx context.Context, fl mold.FieldLevel) error {
		if fl.Field().String() == "ok" {
			return errors.New("ALREADY OK")
		}

		fl.Field().SetString("default")

		return nil
	})

	t.Run("default map creation", func(t *testing.T) {
		type MapStruct struct {
			Map map[string]string `mold:"defaultMap,dive,defaultStr"`
		}

		var tt MapStruct

		err := transformer.Struct(context.Background(), &tt)
		require.NoError(t, err)
		assert.Len(t, tt.Map, 0)
	})

	t.Run("existing map transformation", func(t *testing.T) {
		type ExistingMapStruct struct {
			Map map[string]string `mold:"defaultMap,dive,defaultStr"`
		}

		tt2 := ExistingMapStruct{
			Map: map[string]string{"key": ""},
		}
		err := transformer.Struct(context.Background(), &tt2)
		require.NoError(t, err)
		assert.Len(t, tt2.Map, 1)
		assert.Equal(t, "default", tt2.Map["key"])
	})

	t.Run("transformation error in map value", func(t *testing.T) {
		type MapErrorStruct struct {
			Map map[string]string `mold:"defaultMap,dive,defaultStr"`
		}

		tt3 := MapErrorStruct{
			Map: map[string]string{"key": "ok"},
		}
		err := transformer.Struct(context.Background(), &tt3)
		assert.Error(t, err)
		assert.Equal(t, "ALREADY OK", err.Error())
	})
}

func TestInterface(t *testing.T) {
	type InnerStruct struct {
		STR    string
		String string `mold:"defaultStr"`
	}

	type InnerErrorStruct struct {
		String string `mold:"error"`
	}

	transformer := New()
	transformer.Register("default", func(ctx context.Context, fl mold.FieldLevel) error {
		fl.Field().Set(reflect.ValueOf(InnerStruct{STR: "test"}))

		return nil
	})
	transformer.Register("default2", func(ctx context.Context, fl mold.FieldLevel) error {
		fl.Field().Set(reflect.ValueOf(InnerErrorStruct{}))

		return nil
	})
	transformer.Register("defaultStr", func(ctx context.Context, fl mold.FieldLevel) error {
		if hasValue(fl.Field()) && fl.Field().String() == "ok" {
			return errors.New("ALREADY OK")
		}

		fl.Field().Set(reflect.ValueOf("default"))

		return nil
	})
	transformer.Register("error", func(ctx context.Context, fl mold.FieldLevel) error {
		return errors.New("BAD VALUE")
	})

	t.Run("interface with struct transformation", func(t *testing.T) {
		type InterfaceStruct struct {
			Iface any `mold:"default"`
		}

		var tt InterfaceStruct

		err := transformer.Struct(context.Background(), &tt)
		require.NoError(t, err)
		assert.NotNil(t, tt.Iface)

		inner, ok := tt.Iface.(InnerStruct)
		assert.True(t, ok)
		assert.Equal(t, "default", inner.String)
		assert.Equal(t, "test", inner.STR)
	})

	t.Run("interface transformation error", func(t *testing.T) {
		type InterfaceErrorStruct struct {
			Iface any `mold:"default2"`
		}

		var tt2 InterfaceErrorStruct

		err := transformer.Struct(context.Background(), &tt2)
		assert.Error(t, err)
	})

	t.Run("interface string transformation", func(t *testing.T) {
		type InterfaceStringStruct struct {
			Iface any `mold:"defaultStr"`
		}

		var tt3 InterfaceStringStruct

		tt3.Iface = "String"
		err := transformer.Struct(context.Background(), &tt3)
		require.NoError(t, err)
		assert.Equal(t, "default", tt3.Iface.(string))
	})

	t.Run("interface nil transformation", func(t *testing.T) {
		type InterfaceNilStruct struct {
			Iface any `mold:"defaultStr,defaultStr"`
		}

		var tt4 InterfaceNilStruct

		tt4.Iface = nil
		err := transformer.Struct(context.Background(), &tt4)
		require.NoError(t, err)
		assert.Equal(t, "default", tt4.Iface.(string))
	})

	t.Run("interface transformation chain error", func(t *testing.T) {
		type InterfaceChainErrorStruct struct {
			Iface any `mold:"defaultStr,error"`
		}

		var tt5 InterfaceChainErrorStruct

		tt5.Iface = "String"
		err := transformer.Struct(context.Background(), &tt5)
		assert.Error(t, err)
	})
}

func TestInterfacePtr(t *testing.T) {
	type InnerPtrStruct struct {
		String string `mold:"defaultStr"`
	}

	transformer := New()
	transformer.Register("default", func(ctx context.Context, fl mold.FieldLevel) error {
		fl.Field().Set(reflect.ValueOf(new(InnerPtrStruct)))

		return nil
	})
	transformer.Register("defaultStr", func(ctx context.Context, fl mold.FieldLevel) error {
		if fl.Field().String() == "ok" {
			return errors.New("ALREADY OK")
		}

		fl.Field().SetString("default")

		return nil
	})

	t.Run("interface pointer transformation", func(t *testing.T) {
		type InterfacePtrStruct struct {
			Iface any `mold:"default"`
		}

		var tt InterfacePtrStruct

		err := transformer.Struct(context.Background(), &tt)
		require.NoError(t, err)
		assert.NotNil(t, tt.Iface)

		inner, ok := tt.Iface.(*InnerPtrStruct)
		assert.True(t, ok)
		assert.Equal(t, "default", inner.String)
	})

	t.Run("interface struct without transformation", func(t *testing.T) {
		type InterfaceNoTransformStruct struct {
			Iface any
		}

		var tt2 InterfaceNoTransformStruct

		tt2.Iface = InnerPtrStruct{}
		err := transformer.Struct(context.Background(), &tt2)
		require.NoError(t, err)
	})
}

func TestStructLevel(t *testing.T) {
	type StructLevelStruct struct {
		String string
	}

	transformer := New()
	transformer.RegisterStructLevel(func(ctx context.Context, sl mold.StructLevel) error {
		s := sl.Struct().Interface().(StructLevelStruct)
		if s.String == "error" {
			return errors.New("BAD VALUE")
		}

		s.String = "test"
		sl.Struct().Set(reflect.ValueOf(s))

		return nil
	}, StructLevelStruct{})

	t.Run("struct level transformation success", func(t *testing.T) {
		var tt StructLevelStruct

		err := transformer.Struct(context.Background(), &tt)
		require.NoError(t, err)
		assert.Equal(t, "test", tt.String)
	})

	t.Run("struct level transformation error", func(t *testing.T) {
		var tt StructLevelStruct

		tt.String = "error"
		err := transformer.Struct(context.Background(), &tt)
		assert.Error(t, err)
	})
}

func TestTimeType(t *testing.T) {
	transformer := New()
	transformer.Register("default", func(ctx context.Context, fl mold.FieldLevel) error {
		fl.Field().Set(reflect.ValueOf(time.Now()))

		return nil
	})

	t.Run("time field transformation", func(t *testing.T) {
		var tt time.Time

		err := transformer.Field(context.Background(), &tt, "default")
		require.NoError(t, err)
	})

	t.Run("time field with invalid dive", func(t *testing.T) {
		var tt time.Time

		err := transformer.Field(context.Background(), &tt, "default,dive")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidDive))
	})
}

func TestParam(t *testing.T) {
	transformer := New()
	transformer.Register("ltrim", func(ctx context.Context, fl mold.FieldLevel) error {
		fl.Field().SetString(strings.TrimLeft(fl.Field().String(), fl.Param()))

		return nil
	})

	t.Run("parameter transformation", func(t *testing.T) {
		type ParameterStruct struct {
			String string `mold:"ltrim=#$_"`
		}

		tt := ParameterStruct{String: "_test"}
		err := transformer.Struct(context.Background(), &tt)
		require.NoError(t, err)
		assert.Equal(t, "test", tt.String)
	})
}

func TestDiveKeys(t *testing.T) {
	transformer := New()
	transformer.Register("default", func(ctx context.Context, fl mold.FieldLevel) error {
		fl.Field().Set(reflect.ValueOf("after"))

		return nil
	})
	transformer.Register("err", func(ctx context.Context, fl mold.FieldLevel) error {
		return errors.New("err")
	})

	t.Run("dive keys transformation", func(t *testing.T) {
		type DiveKeysStruct struct {
			Map map[string]string `mold:"dive,keys,default,endkeys,default"`
		}

		test := DiveKeysStruct{
			Map: map[string]string{
				"b4": "b4",
			},
		}

		err := transformer.Struct(context.Background(), &test)
		require.NoError(t, err)

		val := test.Map["after"]
		assert.Equal(t, "after", val)
	})

	t.Run("field dive keys transformation", func(t *testing.T) {
		m := map[string]string{
			"b4": "b4",
		}

		err := transformer.Field(context.Background(), &m, "dive,keys,default,endkeys,default")
		require.NoError(t, err)

		val := m["after"]
		assert.Equal(t, "after", val)
	})

	t.Run("invalid keys tag usage", func(t *testing.T) {
		m := map[string]string{"b4": "b4"}
		err := transformer.Field(context.Background(), &m, "keys,endkeys,default")
		assert.Equal(t, ErrInvalidKeysTag, err)
	})

	t.Run("undefined keys tag", func(t *testing.T) {
		m := map[string]string{"b4": "b4"}
		err := transformer.Field(context.Background(), &m, "dive,endkeys,default")
		assert.Equal(t, ErrUndefinedKeysTag, err)
	})

	t.Run("undefined tag in keys", func(t *testing.T) {
		m := map[string]string{"b4": "b4"}
		err := transformer.Field(context.Background(), &m, "dive,keys,undefinedtag")

		var undefinedErr *ErrUndefinedTag
		assert.ErrorAs(t, err, &undefinedErr)
		assert.Equal(t, "undefinedtag", undefinedErr.tag)
	})

	t.Run("error in keys transformation", func(t *testing.T) {
		m := map[string]string{"b4": "b4"}
		err := transformer.Field(context.Background(), &m, "dive,keys,err,endkeys")
		assert.Error(t, err)
	})

	t.Run("error in values transformation", func(t *testing.T) {
		m := map[string]string{"b4": "b4"}
		err := transformer.Field(context.Background(), &m, "dive,keys,default,endkeys,err")
		assert.Error(t, err)
	})
}

func TestStructArray(t *testing.T) {
	type ArrayInnerStruct struct {
		String string `mold:"defaultStr"`
	}

	transformer := New()
	transformer.Register("defaultArr", func(ctx context.Context, fl mold.FieldLevel) error {
		if hasValue(fl.Field()) {
			return nil
		}

		fl.Field().Set(reflect.MakeSlice(fl.Field().Type(), 2, 2))

		return nil
	})
	transformer.Register("defaultStr", func(ctx context.Context, fl mold.FieldLevel) error {
		if fl.Field().String() == "ok" {
			return errors.New("ALREADY OK")
		}

		fl.Field().SetString("default")

		return nil
	})

	t.Run("struct array transformation", func(t *testing.T) {
		type StructArrayStruct struct {
			Inner    ArrayInnerStruct
			Arr      []ArrayInnerStruct `mold:"defaultArr"`
			ArrDive  []ArrayInnerStruct `mold:"defaultArr,dive"`
			ArrNoTag []ArrayInnerStruct
		}

		var tt StructArrayStruct

		err := transformer.Struct(context.Background(), &tt)
		require.NoError(t, err)
		assert.Len(t, tt.Arr, 2)
		assert.Len(t, tt.ArrDive, 2)
		assert.Empty(t, tt.Arr[0].String)
		assert.Empty(t, tt.Arr[1].String)
		assert.Equal(t, "default", tt.ArrDive[0].String)
		assert.Equal(t, "default", tt.ArrDive[1].String)
		assert.Equal(t, "default", tt.Inner.String)
	})

	t.Run("existing array without dive", func(t *testing.T) {
		type ExistingArrayNoDiveStruct struct {
			Arr []ArrayInnerStruct `mold:"defaultArr"`
		}

		tt2 := ExistingArrayNoDiveStruct{
			Arr: make([]ArrayInnerStruct, 1),
		}
		err := transformer.Struct(context.Background(), &tt2)
		require.NoError(t, err)
		assert.Len(t, tt2.Arr, 1)
		assert.Empty(t, tt2.Arr[0].String)
	})

	t.Run("existing array values preserved", func(t *testing.T) {
		type PreservedArrayStruct struct {
			Arr []ArrayInnerStruct `mold:"defaultArr"`
		}

		tt3 := PreservedArrayStruct{
			Arr: []ArrayInnerStruct{{"ok"}},
		}
		err := transformer.Struct(context.Background(), &tt3)
		require.NoError(t, err)
		assert.Len(t, tt3.Arr, 1)
		assert.Equal(t, "ok", tt3.Arr[0].String)
	})

	t.Run("dive transformation error", func(t *testing.T) {
		type DiveErrorStruct struct {
			ArrDive []ArrayInnerStruct `mold:"defaultArr,dive"`
		}

		tt4 := DiveErrorStruct{
			ArrDive: []ArrayInnerStruct{{"ok"}},
		}
		err := transformer.Struct(context.Background(), &tt4)
		assert.Error(t, err)
		assert.Equal(t, "ALREADY OK", err.Error())
	})

	t.Run("no tag array", func(t *testing.T) {
		type NoTagArrayStruct struct {
			ArrNoTag []ArrayInnerStruct
		}

		tt5 := NoTagArrayStruct{
			ArrNoTag: make([]ArrayInnerStruct, 1),
		}
		err := transformer.Struct(context.Background(), &tt5)
		require.NoError(t, err)
		assert.Len(t, tt5.ArrNoTag, 1)
		assert.Empty(t, tt5.ArrNoTag[0].String)
	})
}

func TestSiblingField(t *testing.T) {
	transformer := New()

	// Register a transformation function that uses SiblingField
	transformer.Register("translate", func(ctx context.Context, fl mold.FieldLevel) error {
		// Get current field value (Status)
		statusValue := fl.Field().Int()

		// Translate based on status value
		var translatedName string

		switch statusValue {
		case 1:
			translatedName = "Active"
		case 2:
			translatedName = "Inactive"
		case 3:
			translatedName = "Pending"
		default:
			translatedName = "Unknown"
		}

		// Get sibling field StatusName and set value
		if statusNameField, ok := fl.SiblingField("StatusName"); ok {
			if statusNameField.CanSet() {
				statusNameField.SetString(translatedName)
			}
		}

		return nil
	})

	type UserStruct struct {
		Status     int `mold:"translate"`
		StatusName string
		Name       string
	}

	t.Run("sibling field translation", func(t *testing.T) {
		user := &UserStruct{
			Status: 2,
			Name:   "John Doe",
		}

		err := transformer.Struct(context.Background(), user)
		require.NoError(t, err)

		// Status should remain unchanged
		assert.Equal(t, 2, user.Status)
		// StatusName should be translated
		assert.Equal(t, "Inactive", user.StatusName)
		// Name should remain unchanged
		assert.Equal(t, "John Doe", user.Name)
	})

	t.Run("sibling field not found", func(t *testing.T) {
		// Register a function that tries to access non-existent sibling
		transformer.Register("test_missing", func(ctx context.Context, fl mold.FieldLevel) error {
			if _, ok := fl.SiblingField("NonExistentField"); ok {
				return errors.New("should not find non-existent field")
			}

			return nil
		})

		type TestStruct struct {
			Field1 string `mold:"test_missing"`
			Field2 string
		}

		test := &TestStruct{Field1: "test"}
		err := transformer.Struct(context.Background(), test)
		require.NoError(t, err)
	})

	t.Run("multiple status translations", func(t *testing.T) {
		tests := []struct {
			status   int
			expected string
		}{
			{1, "Active"},
			{2, "Inactive"},
			{3, "Pending"},
			{99, "Unknown"},
		}

		for _, tt := range tests {
			user := &UserStruct{
				Status: tt.status,
				Name:   "Test User",
			}

			err := transformer.Struct(context.Background(), user)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, user.StatusName)
		}
	})
}
