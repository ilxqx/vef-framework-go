package apis

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/go-streams"
	"github.com/samber/lo"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/storage"
)

// validateColumnsExist validates that the specified columns exist in the model schema.
func validateColumnsExist(schema *schema.Table, columns ...struct {
	name   string
	column string
},
) error {
	for _, c := range columns {
		if c.column != constants.Empty {
			if !schema.HasField(c.column) {
				return result.Err(i18n.T("field_not_exist_in_model", map[string]any{
					"field": c.column,
					"name":  c.name,
					"model": schema.TypeName,
				}))
			}
		}
	}

	return nil
}

// validateOptionColumns validates columns for DataOptionColumnMapping.
func validateOptionColumns(schema *schema.Table, mapping *DataOptionColumnMapping) error {
	columns := []struct {
		name   string
		column string
	}{
		{"labelColumn", mapping.LabelColumn},
		{"valueColumn", mapping.ValueColumn},
	}

	if mapping.DescriptionColumn != constants.Empty {
		columns = append(columns, struct {
			name   string
			column string
		}{"descriptionColumn", mapping.DescriptionColumn})
	}

	return validateColumnsExist(schema, columns...)
}

// mergeOptionColumnMapping merges the provided mapping with default mapping.
// Uses fallback values for empty columns based on the provided default mapping or system defaults.
func mergeOptionColumnMapping(mapping, defaultMapping *DataOptionColumnMapping) {
	if defaultMapping == nil {
		defaultMapping = defaultDataOptionColumnMapping
	}

	if mapping.LabelColumn == constants.Empty {
		mapping.LabelColumn = lo.CoalesceOrEmpty(defaultMapping.LabelColumn, defaultLabelColumn)
	}

	if mapping.ValueColumn == constants.Empty {
		mapping.ValueColumn = lo.CoalesceOrEmpty(defaultMapping.ValueColumn, defaultValueColumn)
	}

	if mapping.DescriptionColumn == constants.Empty {
		mapping.DescriptionColumn = defaultMapping.DescriptionColumn
	}

	if len(mapping.MetaColumns) == 0 {
		mapping.MetaColumns = defaultMapping.MetaColumns
	}
}

// ApplyDataPermission applies data permission filtering to a SelectQuery.
func ApplyDataPermission(query orm.SelectQuery, ctx fiber.Ctx) error {
	if applier := contextx.DataPermApplier(ctx); applier != nil {
		if err := applier.Apply(query); err != nil {
			return fmt.Errorf("failed to apply data permission: %w", err)
		}
	}

	return nil
}

// GetAuditUserNameRelations returns RelationSpecs for creator and updater joins.
func GetAuditUserNameRelations(userModel any, nameColumn ...string) []*orm.RelationSpec {
	nc := defaultAuditUserNameColumn
	if len(nameColumn) > 0 {
		nc = nameColumn[0]
	}

	relations := []*orm.RelationSpec{
		{
			Model:         userModel,
			Alias:         "creator",
			JoinType:      orm.JoinLeft,
			ForeignColumn: "created_by",
			SelectedColumns: []orm.ColumnInfo{
				{
					Name:  nc,
					Alias: constants.ColumnCreatedByName,
				},
			},
		},
		{
			Model:         userModel,
			Alias:         "updater",
			JoinType:      orm.JoinLeft,
			ForeignColumn: "updated_by",
			SelectedColumns: []orm.ColumnInfo{
				{
					Name:  nc,
					Alias: constants.ColumnUpdatedByName,
				},
			},
		},
	}

	return relations
}

// columnAliasPattern matches "column AS alias" format (case-insensitive AS, flexible spaces).
var columnAliasPattern = regexp.MustCompile(`^\s*(.+?)\s+(?i:as)\s+(.+?)\s*$`)

// parseMetaColumn parses a single meta column specification into (column, alias).
// Supports formats:
//   - "column" -> ("column", "column")
//   - "column AS alias" -> ("column", "alias")
//   - "column as alias" -> ("column", "alias")
func parseMetaColumn(spec string) (column, alias string) {
	if matches := columnAliasPattern.FindStringSubmatch(spec); len(matches) == 3 {
		column = strings.TrimSpace(matches[1])
		alias = strings.TrimSpace(matches[2])

		return column, alias
	}

	// No alias specified, use column name as alias
	trimmed := strings.TrimSpace(spec)

	return trimmed, trimmed
}

// parseMetaColumns parses meta column specifications into structured info.
// This function should be called once to avoid redundant parsing.
func parseMetaColumns(specs []string) []orm.ColumnInfo {
	if len(specs) == 0 {
		return nil
	}

	return streams.MapTo(
		streams.FromSlice(specs),
		func(spec string) orm.ColumnInfo {
			columnName, aliasName := parseMetaColumn(spec)

			return orm.ColumnInfo{
				Name:  columnName,
				Alias: aliasName,
			}
		},
	).Collect()
}

// validateMetaColumns validates that all meta columns exist in the table schema.
func validateMetaColumns(schema *schema.Table, metaColumns []orm.ColumnInfo) error {
	for _, col := range metaColumns {
		if !schema.HasField(col.Name) {
			return result.Err(i18n.T("field_not_exist_in_model", map[string]any{
				"field": col.Name,
				"name":  "metaColumns",
				"model": schema.TypeName,
			}))
		}
	}

	return nil
}

// buildMetaJsonExpr constructs a JSON_OBJECT expression for meta columns.
func buildMetaJsonExpr(eb orm.ExprBuilder, metaColumns []orm.ColumnInfo) schema.QueryAppender {
	// Build JSON_OBJECT arguments: key1, value1, key2, value2, ...
	jsonArgs := streams.FlatMap(
		streams.FromSlice(metaColumns),
		func(col orm.ColumnInfo) streams.Stream[any] {
			return streams.Of[any](col.Alias, eb.Column(col.Name))
		},
	).Collect()

	return eb.JsonObject(jsonArgs...)
}

// batchCleanup cleans up files in batch, collecting all errors.
func batchCleanup[TModel any](
	ctx context.Context,
	promoter storage.Promoter[TModel],
	models []TModel,
) error {
	errs := streams.MapTo(
		streams.Range(0, len(models)),
		func(i int) error {
			if err := promoter.Promote(ctx, nil, &models[i]); err != nil {
				return fmt.Errorf("model %d: %w", i, err)
			}

			return nil
		},
	).Filter(func(err error) bool {
		return err != nil
	}).Collect()

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// batchRollback rolls back files in batch, collecting all errors.
// Used in update operations to rollback to old files when failures occur.
// Count specifies how many models to rollback (typically the index where the error occurred).
func batchRollback[TModel any](
	ctx context.Context,
	promoter storage.Promoter[TModel],
	oldModels, newModels []TModel,
	count int,
) error {
	errs := streams.MapTo(
		streams.Range(0, count),
		func(i int) error {
			if err := promoter.Promote(ctx, &newModels[i], &oldModels[i]); err != nil {
				return fmt.Errorf("model %d: %w", i, err)
			}

			return nil
		},
	).Filter(func(err error) bool {
		return err != nil
	}).Collect()

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
