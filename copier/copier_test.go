package copier

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilxqx/vef-framework-go/datetime"
	"github.com/ilxqx/vef-framework-go/decimal"
	"github.com/ilxqx/vef-framework-go/null"
)

func TestCopyBasic(t *testing.T) {
	t.Run("Struct", func(t *testing.T) {
		type Source struct {
			Name string
			Age  int
		}

		type Dest struct {
			Name string
			Age  int
		}

		src := Source{Name: "John", Age: 30}

		var dst Dest

		require.NoError(t, Copy(src, &dst))
		assert.Equal(t, "John", dst.Name)
		assert.Equal(t, 30, dst.Age)
	})
}

func TestCopyConverters(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "NullStringToString",
			run: func(t *testing.T) {
				type Source struct {
					Value null.String
				}
				type Dest struct {
					Value string
				}

				src := Source{Value: null.StringFrom("test")}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.Equal(t, "test", dst.Value)
			},
		},
		{
			name: "StringToNullString",
			run: func(t *testing.T) {
				type Source struct {
					Value string
				}
				type Dest struct {
					Value null.String
				}

				src := Source{Value: "test"}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.Equal(t, "test", dst.Value.String)
			},
		},
		{
			name: "NullIntToInt64",
			run: func(t *testing.T) {
				type Source struct {
					Value null.Int
				}
				type Dest struct {
					Value int64
				}

				src := Source{Value: null.IntFrom(42)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.Equal(t, int64(42), dst.Value)
			},
		},
		{
			name: "Int64ToNullInt",
			run: func(t *testing.T) {
				type Source struct {
					Value int64
				}
				type Dest struct {
					Value null.Int
				}

				src := Source{Value: 42}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.Equal(t, int64(42), dst.Value.Int64)
			},
		},
		{
			name: "NullInt16ToInt16",
			run: func(t *testing.T) {
				type Source struct {
					Value null.Int16
				}
				type Dest struct {
					Value int16
				}

				src := Source{Value: null.Int16From(100)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.Equal(t, int16(100), dst.Value)
			},
		},
		{
			name: "Int16ToNullInt16",
			run: func(t *testing.T) {
				type Source struct {
					Value int16
				}
				type Dest struct {
					Value null.Int16
				}

				src := Source{Value: 200}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.Equal(t, int16(200), dst.Value.Int16)
			},
		},
		{
			name: "NullInt32ToInt32",
			run: func(t *testing.T) {
				type Source struct {
					Value null.Int32
				}
				type Dest struct {
					Value int32
				}

				src := Source{Value: null.Int32From(12345)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.Equal(t, int32(12345), dst.Value)
			},
		},
		{
			name: "Int32ToNullInt32",
			run: func(t *testing.T) {
				type Source struct {
					Value int32
				}
				type Dest struct {
					Value null.Int32
				}

				src := Source{Value: 54321}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.Equal(t, int32(54321), dst.Value.Int32)
			},
		},
		{
			name: "NullFloatToFloat64",
			run: func(t *testing.T) {
				type Source struct {
					Value null.Float
				}
				type Dest struct {
					Value float64
				}

				src := Source{Value: null.FloatFrom(3.14)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.Equal(t, 3.14, dst.Value)
			},
		},
		{
			name: "Float64ToNullFloat",
			run: func(t *testing.T) {
				type Source struct {
					Value float64
				}
				type Dest struct {
					Value null.Float
				}

				src := Source{Value: 3.14}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.Equal(t, 3.14, dst.Value.Float64)
			},
		},
		{
			name: "NullByteToByte",
			run: func(t *testing.T) {
				type Source struct {
					Value null.Byte
				}
				type Dest struct {
					Value byte
				}

				src := Source{Value: null.ByteFrom(255)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.Equal(t, byte(255), dst.Value)
			},
		},
		{
			name: "ByteToNullByte",
			run: func(t *testing.T) {
				type Source struct {
					Value byte
				}
				type Dest struct {
					Value null.Byte
				}

				src := Source{Value: 128}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.Equal(t, byte(128), dst.Value.Byte)
			},
		},
		{
			name: "NullBoolToBool",
			run: func(t *testing.T) {
				type Source struct {
					Value null.Bool
				}
				type Dest struct {
					Value bool
				}

				src := Source{Value: null.BoolFrom(true)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value)
			},
		},
		{
			name: "BoolToNullBool",
			run: func(t *testing.T) {
				type Source struct {
					Value bool
				}
				type Dest struct {
					Value null.Bool
				}

				src := Source{Value: true}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.True(t, dst.Value.Bool)
			},
		},
		{
			name: "NullDateTimeToDateTime",
			run: func(t *testing.T) {
				type Source struct {
					Value null.DateTime
				}
				type Dest struct {
					Value datetime.DateTime
				}

				testValue := datetime.Of(time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC))
				src := Source{Value: null.DateTimeFrom(testValue)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.Equal(t, testValue, dst.Value)
			},
		},
		{
			name: "DateTimeToNullDateTime",
			run: func(t *testing.T) {
				type Source struct {
					Value datetime.DateTime
				}
				type Dest struct {
					Value null.DateTime
				}

				testValue := datetime.Of(time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC))
				src := Source{Value: testValue}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.Equal(t, testValue, dst.Value.V)
			},
		},
		{
			name: "NullDateToDate",
			run: func(t *testing.T) {
				type Source struct {
					Value null.Date
				}
				type Dest struct {
					Value datetime.Date
				}

				testValue := datetime.DateOf(time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC))
				src := Source{Value: null.DateFrom(testValue)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.Equal(t, testValue, dst.Value)
			},
		},
		{
			name: "DateToNullDate",
			run: func(t *testing.T) {
				type Source struct {
					Value datetime.Date
				}
				type Dest struct {
					Value null.Date
				}

				testValue := datetime.DateOf(time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC))
				src := Source{Value: testValue}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.Equal(t, testValue, dst.Value.V)
			},
		},
		{
			name: "NullTimeToTime",
			run: func(t *testing.T) {
				type Source struct {
					Value null.Time
				}
				type Dest struct {
					Value datetime.Time
				}

				testValue := datetime.TimeOf(time.Date(0, 1, 1, 15, 30, 45, 0, time.UTC))
				src := Source{Value: null.TimeFrom(testValue)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.Equal(t, testValue, dst.Value)
			},
		},
		{
			name: "TimeToNullTime",
			run: func(t *testing.T) {
				type Source struct {
					Value datetime.Time
				}
				type Dest struct {
					Value null.Time
				}

				testValue := datetime.TimeOf(time.Date(0, 1, 1, 15, 30, 45, 0, time.UTC))
				src := Source{Value: testValue}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.Equal(t, testValue, dst.Value.V)
			},
		},
		{
			name: "NullDecimalToDecimal",
			run: func(t *testing.T) {
				type Source struct {
					Value null.Decimal
				}
				type Dest struct {
					Value decimal.Decimal
				}

				testDecimal := decimal.NewFromFloat(123.45)
				src := Source{Value: null.DecimalFrom(testDecimal)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, testDecimal.Equal(dst.Value))
			},
		},
		{
			name: "DecimalToNullDecimal",
			run: func(t *testing.T) {
				type Source struct {
					Value decimal.Decimal
				}
				type Dest struct {
					Value null.Decimal
				}

				testDecimal := decimal.NewFromFloat(123.45)
				src := Source{Value: testDecimal}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.True(t, testDecimal.Equal(dst.Value.Decimal))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.run)
	}
}

func TestCopyPointerConverters(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "NullStringToStringPtr",
			run: func(t *testing.T) {
				type Source struct {
					Value null.String
				}
				type Dest struct {
					Value *string
				}

				src := Source{Value: null.StringFrom("pointer")}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				require.NotNil(t, dst.Value)
				assert.Equal(t, "pointer", *dst.Value)
			},
		},
		{
			name: "StringPtrToNullString",
			run: func(t *testing.T) {
				type Source struct {
					Value *string
				}
				type Dest struct {
					Value null.String
				}

				value := "pointer"
				src := Source{Value: &value}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.Equal(t, "pointer", dst.Value.String)
			},
		},
		{
			name: "NilStringPtrToNullString",
			run: func(t *testing.T) {
				type Source struct {
					Value *string
				}
				type Dest struct {
					Value null.String
				}

				var src Source
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.False(t, dst.Value.Valid)
			},
		},
		{
			name: "InvalidNullStringToStringPtr",
			run: func(t *testing.T) {
				type Source struct {
					Value null.String
				}
				type Dest struct {
					Value *string
				}

				src := Source{Value: null.NewString("", false)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.Nil(t, dst.Value)
			},
		},
		{
			name: "NullIntToIntPtr",
			run: func(t *testing.T) {
				type Source struct {
					Value null.Int
				}
				type Dest struct {
					Value *int64
				}

				src := Source{Value: null.IntFrom(42)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				require.NotNil(t, dst.Value)
				assert.Equal(t, int64(42), *dst.Value)
			},
		},
		{
			name: "IntPtrToNullInt",
			run: func(t *testing.T) {
				type Source struct {
					Value *int64
				}
				type Dest struct {
					Value null.Int
				}

				value := int64(42)
				src := Source{Value: &value}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.Equal(t, int64(42), dst.Value.Int64)
			},
		},
		{
			name: "NullBoolToBoolPtr",
			run: func(t *testing.T) {
				type Source struct {
					Value null.Bool
				}
				type Dest struct {
					Value *bool
				}

				src := Source{Value: null.BoolFrom(true)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				require.NotNil(t, dst.Value)
				assert.True(t, *dst.Value)
			},
		},
		{
			name: "BoolPtrToNullBool",
			run: func(t *testing.T) {
				type Source struct {
					Value *bool
				}
				type Dest struct {
					Value null.Bool
				}

				value := false
				src := Source{Value: &value}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.False(t, dst.Value.Bool)
			},
		},
		{
			name: "NullInt16ToInt16Ptr",
			run: func(t *testing.T) {
				type Source struct {
					Value null.Int16
				}
				type Dest struct {
					Value *int16
				}

				src := Source{Value: null.Int16From(123)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				require.NotNil(t, dst.Value)
				assert.Equal(t, int16(123), *dst.Value)
			},
		},
		{
			name: "Int16PtrToNullInt16",
			run: func(t *testing.T) {
				type Source struct {
					Value *int16
				}
				type Dest struct {
					Value null.Int16
				}

				value := int16(321)
				src := Source{Value: &value}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.Equal(t, int16(321), dst.Value.Int16)
			},
		},
		{
			name: "NullInt32ToInt32Ptr",
			run: func(t *testing.T) {
				type Source struct {
					Value null.Int32
				}
				type Dest struct {
					Value *int32
				}

				src := Source{Value: null.Int32From(111)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				require.NotNil(t, dst.Value)
				assert.Equal(t, int32(111), *dst.Value)
			},
		},
		{
			name: "Int32PtrToNullInt32",
			run: func(t *testing.T) {
				type Source struct {
					Value *int32
				}
				type Dest struct {
					Value null.Int32
				}

				value := int32(222)
				src := Source{Value: &value}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.Equal(t, int32(222), dst.Value.Int32)
			},
		},
		{
			name: "NullFloatToFloatPtr",
			run: func(t *testing.T) {
				type Source struct {
					Value null.Float
				}
				type Dest struct {
					Value *float64
				}

				src := Source{Value: null.FloatFrom(9.87)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				require.NotNil(t, dst.Value)
				assert.Equal(t, 9.87, *dst.Value)
			},
		},
		{
			name: "FloatPtrToNullFloat",
			run: func(t *testing.T) {
				type Source struct {
					Value *float64
				}
				type Dest struct {
					Value null.Float
				}

				value := 6.54
				src := Source{Value: &value}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.Equal(t, value, dst.Value.Float64)
			},
		},
		{
			name: "NullByteToBytePtr",
			run: func(t *testing.T) {
				type Source struct {
					Value null.Byte
				}
				type Dest struct {
					Value *byte
				}

				src := Source{Value: null.ByteFrom(77)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				require.NotNil(t, dst.Value)
				assert.Equal(t, byte(77), *dst.Value)
			},
		},
		{
			name: "BytePtrToNullByte",
			run: func(t *testing.T) {
				type Source struct {
					Value *byte
				}
				type Dest struct {
					Value null.Byte
				}

				value := byte(88)
				src := Source{Value: &value}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.Equal(t, byte(88), dst.Value.Byte)
			},
		},
		{
			name: "NullDateTimeToDateTimePtr",
			run: func(t *testing.T) {
				type Source struct {
					Value null.DateTime
				}
				type Dest struct {
					Value *datetime.DateTime
				}

				testValue := datetime.Of(time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC))
				src := Source{Value: null.DateTimeFrom(testValue)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				require.NotNil(t, dst.Value)
				assert.Equal(t, testValue, *dst.Value)
			},
		},
		{
			name: "DateTimePtrToNullDateTime",
			run: func(t *testing.T) {
				type Source struct {
					Value *datetime.DateTime
				}
				type Dest struct {
					Value null.DateTime
				}

				testValue := datetime.Of(time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC))
				src := Source{Value: &testValue}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.Equal(t, testValue, dst.Value.V)
			},
		},
		{
			name: "NullDateToDatePtr",
			run: func(t *testing.T) {
				type Source struct {
					Value null.Date
				}
				type Dest struct {
					Value *datetime.Date
				}

				testValue := datetime.DateOf(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC))
				src := Source{Value: null.DateFrom(testValue)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				require.NotNil(t, dst.Value)
				assert.Equal(t, testValue, *dst.Value)
			},
		},
		{
			name: "DatePtrToNullDate",
			run: func(t *testing.T) {
				type Source struct {
					Value *datetime.Date
				}
				type Dest struct {
					Value null.Date
				}

				testValue := datetime.DateOf(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC))
				src := Source{Value: &testValue}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.Equal(t, testValue, dst.Value.V)
			},
		},
		{
			name: "NullTimeToTimePtr",
			run: func(t *testing.T) {
				type Source struct {
					Value null.Time
				}
				type Dest struct {
					Value *datetime.Time
				}

				testValue := datetime.TimeOf(time.Date(0, 1, 1, 10, 20, 30, 0, time.UTC))
				src := Source{Value: null.TimeFrom(testValue)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				require.NotNil(t, dst.Value)
				assert.Equal(t, testValue, *dst.Value)
			},
		},
		{
			name: "TimePtrToNullTime",
			run: func(t *testing.T) {
				type Source struct {
					Value *datetime.Time
				}
				type Dest struct {
					Value null.Time
				}

				testValue := datetime.TimeOf(time.Date(0, 1, 1, 5, 10, 15, 0, time.UTC))
				src := Source{Value: &testValue}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.Equal(t, testValue, dst.Value.V)
			},
		},
		{
			name: "NullDecimalToDecimalPtr",
			run: func(t *testing.T) {
				type Source struct {
					Value null.Decimal
				}
				type Dest struct {
					Value *decimal.Decimal
				}

				testValue := decimal.NewFromFloat(456.78)
				src := Source{Value: null.DecimalFrom(testValue)}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				require.NotNil(t, dst.Value)
				assert.True(t, testValue.Equal(*dst.Value))
			},
		},
		{
			name: "DecimalPtrToNullDecimal",
			run: func(t *testing.T) {
				type Source struct {
					Value *decimal.Decimal
				}
				type Dest struct {
					Value null.Decimal
				}

				testValue := decimal.NewFromFloat(654.32)
				src := Source{Value: &testValue}
				var dst Dest

				require.NoError(t, Copy(src, &dst))
				assert.True(t, dst.Value.Valid)
				assert.True(t, testValue.Equal(dst.Value.Decimal))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.run)
	}
}

func TestCopyIntegration(t *testing.T) {
	t.Run("NullToBasic", func(t *testing.T) {
		type Source struct {
			Name   null.String
			Age    null.Int
			Active null.Bool
		}
		type Dest struct {
			Name   string
			Age    int64
			Active bool
		}

		src := Source{
			Name:   null.StringFrom("John Doe"),
			Age:    null.IntFrom(30),
			Active: null.BoolFrom(true),
		}

		var dst Dest

		require.NoError(t, Copy(src, &dst))
		assert.Equal(t, "John Doe", dst.Name)
		assert.Equal(t, int64(30), dst.Age)
		assert.True(t, dst.Active)
	})

	t.Run("BasicToNull", func(t *testing.T) {
		type Source struct {
			Name   string
			Age    int64
			Active bool
		}
		type Dest struct {
			Name   null.String
			Age    null.Int
			Active null.Bool
		}

		src := Source{
			Name:   "Jane Doe",
			Age:    28,
			Active: false,
		}

		var dst Dest

		require.NoError(t, Copy(src, &dst))
		assert.True(t, dst.Name.Valid)
		assert.Equal(t, "Jane Doe", dst.Name.String)
		assert.True(t, dst.Age.Valid)
		assert.Equal(t, int64(28), dst.Age.Int64)
		assert.True(t, dst.Active.Valid)
		assert.False(t, dst.Active.Bool)
	})
}

func TestCopyOptions(t *testing.T) {
	t.Run("IgnoreEmpty", func(t *testing.T) {
		type Source struct {
			Name string
			Age  int
		}

		type Dest struct {
			Name string
			Age  int
		}

		dst := Dest{Name: "Initial Name", Age: 25}
		src := Source{Name: "", Age: 30}

		require.NoError(t, Copy(src, &dst, WithIgnoreEmpty()))
		assert.Equal(t, 30, dst.Age)
	})

	t.Run("CaseInsensitive", func(t *testing.T) {
		type Source struct {
			NAME string
		}

		type Dest struct {
			Name string
		}

		src := Source{NAME: "John Doe"}

		var dst Dest

		require.NoError(t, Copy(src, &dst, WithCaseInsensitive()))
		assert.Equal(t, "John Doe", dst.Name)
	})
}

func TestCopyError(t *testing.T) {
	t.Run("NonPointerDestination", func(t *testing.T) {
		type Source struct {
			Name string
		}

		type Dest struct {
			Name string
		}

		src := Source{Name: "John"}
		dst := Dest{}

		err := Copy(src, dst)
		assert.Error(t, err)
	})
}
