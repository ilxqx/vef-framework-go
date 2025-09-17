package search

import "github.com/ilxqx/vef-framework-go/orm"

// Applier is a factory function that creates a function which accepts a value of generic type
// and returns an orm.ApplyFunc function. This ApplyFunc can be applied to an orm.ConditionBuilder
// to build query conditions.
func Applier[T any]() func(T) orm.ApplyFunc[orm.ConditionBuilder] {
	f := New[T]() // f creates a new search instance for type T

	return func(value T) orm.ApplyFunc[orm.ConditionBuilder] { // Returns a function that takes a value and returns an ApplyFunc
		return func(cb orm.ConditionBuilder) orm.ConditionBuilder { // ApplyFunc applies search conditions to condition builder
			f.Apply(cb, value) // Apply applies the search conditions with the given value

			return cb // Return the modified condition builder
		}
	}
}
