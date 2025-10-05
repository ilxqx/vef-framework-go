package dbhelpers

import (
	"testing"

	"github.com/ilxqx/vef-framework-go/constants"
)

// benchResult is used to avoid compiler optimizations eliminating the result.
var benchResult string

// columnWithAliasPlus is the `+` concatenation variant for comparison in benchmarks.
func columnWithAliasPlus(column string, alias ...string) string {
	// Early return when alias is not provided or empty
	if len(alias) == 0 || alias[0] == constants.Empty {
		return column
	}

	// Use direct string concatenation
	return alias[0] + constants.Dot + column
}

func BenchmarkColumnWithAlias(b *testing.B) {
	alias := "su"
	column := "username"

	// With alias: compare `+` vs strings.Builder implementation
	b.Run("plus/withAlias", func(b *testing.B) {
		b.ReportAllocs()

		var r string
		for b.Loop() {
			r = columnWithAliasPlus(column, alias)
		}

		benchResult = r
	})

	b.Run("builder/withAlias", func(b *testing.B) {
		b.ReportAllocs()

		var r string
		for b.Loop() {
			r = ColumnWithAlias(column, alias)
		}

		benchResult = r
	})

	// No alias: compare `+` vs strings.Builder implementation
	b.Run("plus/noAlias", func(b *testing.B) {
		b.ReportAllocs()

		var r string
		for b.Loop() {
			r = columnWithAliasPlus(column)
		}

		benchResult = r
	})

	b.Run("builder/noAlias", func(b *testing.B) {
		b.ReportAllocs()

		var r string
		for b.Loop() {
			r = ColumnWithAlias(column)
		}

		benchResult = r
	})
}
