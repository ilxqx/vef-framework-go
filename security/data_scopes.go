package security

import (
	"fmt"

	"github.com/ilxqx/vef-framework-go/orm"
)

// AllDataScope grants access to all data without any restrictions.
// This is typically used for system administrators or users with full data access.
type AllDataScope struct{}

// NewAllDataScope creates a new AllDataScope instance.
func NewAllDataScope() DataScope {
	return &AllDataScope{}
}

func (*AllDataScope) Key() string {
	return "all"
}

func (*AllDataScope) Supports(_ *Principal, _ *orm.Table) bool {
	// Always supports any table
	return true
}

func (*AllDataScope) Apply(_ *Principal, _ orm.SelectQuery) error {
	// No filtering - allow all data
	return nil
}

// SelfDataScope restricts access to data created by the user themselves.
// This is commonly used for personal data access where users can only see their own records.
type SelfDataScope struct {
	createdByField string // Database column name for the creator, defaults to "created_by"
}

// NewSelfDataScope creates a new SelfDataScope instance.
// The createdByField parameter specifies the database column name for the creator.
// If empty, it defaults to "created_by".
func NewSelfDataScope(createdByField string) DataScope {
	if createdByField == "" {
		createdByField = "created_by"
	}

	return &SelfDataScope{
		createdByField: createdByField,
	}
}

func (*SelfDataScope) Key() string {
	return "self"
}

func (s *SelfDataScope) Supports(_ *Principal, table *orm.Table) bool {
	// Check if the table has the created_by field
	field, _ := table.Field(s.createdByField)

	return field != nil
}

func (s *SelfDataScope) Apply(principal *Principal, query orm.SelectQuery) error {
	if principal == nil {
		return fmt.Errorf("principal is required for self data scope")
	}

	query.Where(func(cb orm.ConditionBuilder) {
		cb.Equals(s.createdByField, principal.Id)
	})

	return nil
}

// DepartmentDataScope restricts access to data within the user's department.
// This is commonly used in organizational structures where users can see data from their department.
type DepartmentDataScope struct {
	deptIdField      string // Database column name for department ID, defaults to "dept_id"
	principalDeptKey string // Key in principal.Details to get department ID, defaults to "deptId"
}

// NewDepartmentDataScope creates a new DepartmentDataScope instance.
// The deptIdField parameter specifies the database column name for department ID.
// The principalDeptKey parameter specifies the key in principal.Details to get department ID.
// If empty, they default to "dept_id" and "deptId" respectively.
func NewDepartmentDataScope(deptIdField, principalDeptKey string) DataScope {
	if deptIdField == "" {
		deptIdField = "dept_id"
	}

	if principalDeptKey == "" {
		principalDeptKey = "deptId"
	}

	return &DepartmentDataScope{
		deptIdField:      deptIdField,
		principalDeptKey: principalDeptKey,
	}
}

func (*DepartmentDataScope) Key() string {
	return "dept"
}

func (d *DepartmentDataScope) Supports(_ *Principal, table *orm.Table) bool {
	// Check if the table has the dept_id field
	field, _ := table.Field(d.deptIdField)

	return field != nil
}

func (d *DepartmentDataScope) Apply(principal *Principal, query orm.SelectQuery) error {
	if principal == nil {
		return fmt.Errorf("principal is required for department data scope")
	}

	// Get department ID from principal details
	details, ok := principal.Details.(map[string]any)
	if !ok {
		return fmt.Errorf("principal details must be a map[string]any for department data scope")
	}

	deptId, ok := details[d.principalDeptKey]
	if !ok {
		return fmt.Errorf("missing '%s' in principal details for department data scope", d.principalDeptKey)
	}

	query.Where(func(cb orm.ConditionBuilder) {
		cb.Equals(d.deptIdField, deptId)
	})

	return nil
}

// DepartmentAndSubDataScope restricts access to data within the user's department and all sub-departments.
// This requires a department hierarchy table to resolve sub-departments.
type DepartmentAndSubDataScope struct {
	deptIdField      string // Database column name for department ID, defaults to "dept_id"
	principalDeptKey string // Key in principal.Details to get department ID, defaults to "deptId"
	// Note: In a real implementation, you would need to query the department hierarchy
	// to get all sub-department IDs. For simplicity, this example assumes the principal
	// details contain a "subDeptIds" array.
	principalSubDeptIdsKey string // Key in principal.Details to get sub-department IDs, defaults to "subDeptIds"
}

// NewDepartmentAndSubDataScope creates a new DepartmentAndSubDataScope instance.
func NewDepartmentAndSubDataScope(deptIdField, principalDeptKey, principalSubDeptIdsKey string) DataScope {
	if deptIdField == "" {
		deptIdField = "dept_id"
	}

	if principalDeptKey == "" {
		principalDeptKey = "deptId"
	}

	if principalSubDeptIdsKey == "" {
		principalSubDeptIdsKey = "subDeptIds"
	}

	return &DepartmentAndSubDataScope{
		deptIdField:            deptIdField,
		principalDeptKey:       principalDeptKey,
		principalSubDeptIdsKey: principalSubDeptIdsKey,
	}
}

func (*DepartmentAndSubDataScope) Key() string {
	return "dept_and_sub"
}

func (d *DepartmentAndSubDataScope) Supports(_ *Principal, table *orm.Table) bool {
	// Check if the table has the dept_id field
	field, _ := table.Field(d.deptIdField)

	return field != nil
}

func (d *DepartmentAndSubDataScope) Apply(principal *Principal, query orm.SelectQuery) error {
	if principal == nil {
		return fmt.Errorf("principal is required for department and sub data scope")
	}

	// Get department ID from principal details
	details, ok := principal.Details.(map[string]any)
	if !ok {
		return fmt.Errorf("principal details must be a map[string]any for department and sub data scope")
	}

	deptId, ok := details[d.principalDeptKey]
	if !ok {
		return fmt.Errorf("missing '%s' in principal details for department and sub data scope", d.principalDeptKey)
	}

	// Get sub-department IDs from principal details
	subDeptIds, ok := details[d.principalSubDeptIdsKey]
	if !ok {
		// If no sub-departments, just filter by current department
		query.Where(func(cb orm.ConditionBuilder) {
			cb.Equals(d.deptIdField, deptId)
		})

		return nil
	}

	// Convert sub-department IDs to slice
	var deptIds []any
	switch v := subDeptIds.(type) {
	case []any:
		deptIds = append([]any{deptId}, v...)
	case []string:
		deptIds = append([]any{deptId}, convertToAnySlice(v)...)
	default:
		// Fallback to just current department
		query.Where(func(cb orm.ConditionBuilder) {
			cb.Equals(d.deptIdField, deptId)
		})

		return nil
	}

	// Filter by current department or any sub-departments
	query.Where(func(cb orm.ConditionBuilder) {
		cb.In(d.deptIdField, deptIds)
	})

	return nil
}

// convertToAnySlice converts a string slice to an any slice.
func convertToAnySlice(strs []string) []any {
	result := make([]any, len(strs))
	for i, s := range strs {
		result[i] = s
	}

	return result
}

// CustomFieldDataScope provides a generic data scope that filters by a custom field.
// This is useful for ad-hoc data permission requirements.
type CustomFieldDataScope struct {
	key              string                            // Unique identifier for this scope
	fieldName        string                            // Database column name to filter
	principalKey     string                            // Key in principal.Details to get the filter value
	supportsFunc     func(*Principal, *orm.Table) bool // Custom supports function
	applyFunc        func(*Principal, orm.SelectQuery) // Custom apply function
	requireFieldFunc func(*orm.Table) bool             // Function to check if table has required field
}

// CustomFieldDataScopeOption defines options for CustomFieldDataScope.
type CustomFieldDataScopeOption func(*CustomFieldDataScope)

// WithSupportsFunc sets a custom supports function.
func WithSupportsFunc(fn func(*Principal, *orm.Table) bool) CustomFieldDataScopeOption {
	return func(s *CustomFieldDataScope) {
		s.supportsFunc = fn
	}
}

// WithApplyFunc sets a custom apply function.
func WithApplyFunc(fn func(*Principal, orm.SelectQuery)) CustomFieldDataScopeOption {
	return func(s *CustomFieldDataScope) {
		s.applyFunc = fn
	}
}

// NewCustomFieldDataScope creates a new CustomFieldDataScope instance.
// This allows users to define custom data scopes without implementing the DataScope interface.
func NewCustomFieldDataScope(
	key string,
	fieldName string,
	principalKey string,
	opts ...CustomFieldDataScopeOption,
) DataScope {
	scope := &CustomFieldDataScope{
		key:          key,
		fieldName:    fieldName,
		principalKey: principalKey,
		requireFieldFunc: func(table *orm.Table) bool {
			field, _ := table.Field(fieldName)

			return field != nil
		},
	}

	for _, opt := range opts {
		opt(scope)
	}

	return scope
}

func (c *CustomFieldDataScope) Key() string {
	return c.key
}

func (c *CustomFieldDataScope) Supports(principal *Principal, table *orm.Table) bool {
	if c.supportsFunc != nil {
		return c.supportsFunc(principal, table)
	}

	// Default: check if table has the required field
	if c.requireFieldFunc != nil {
		return c.requireFieldFunc(table)
	}

	return true
}

func (c *CustomFieldDataScope) Apply(principal *Principal, query orm.SelectQuery) error {
	if c.applyFunc != nil {
		c.applyFunc(principal, query)

		return nil
	}

	// Default: equals filter
	if principal == nil {
		return fmt.Errorf("principal is required for custom field data scope")
	}

	details, ok := principal.Details.(map[string]any)
	if !ok {
		return fmt.Errorf("principal details must be a map[string]any for custom field data scope")
	}

	value, ok := details[c.principalKey]
	if !ok {
		return fmt.Errorf("missing '%s' in principal details for custom field data scope", c.principalKey)
	}

	query.Where(func(cb orm.ConditionBuilder) {
		cb.Equals(c.fieldName, value)
	})

	return nil
}
