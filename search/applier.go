package search

import "github.com/ilxqx/vef-framework-go/orm"

// Applier is a factory function that creates a function which accepts a value of generic type
// and returns an orm.ApplyFunc function. This ApplyFunc can be applied to an orm.ConditionBuilder
// to build query conditions.
func Applier[T any]() func(T) orm.ApplyFunc[orm.ConditionBuilder] {
	f := New[T]()

	return func(value T) orm.ApplyFunc[orm.ConditionBuilder] {
		return func(cb orm.ConditionBuilder) orm.ConditionBuilder {
			f.Apply(cb, value)

			return cb
		}
	}
}
