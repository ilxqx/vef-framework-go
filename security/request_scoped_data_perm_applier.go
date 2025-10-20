package security

import (
	"fmt"

	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/orm"
)

// RequestScopedDataPermApplier is the default implementation of DataPermissionApplier.
// It applies data permission filtering using a single DataScope instance.
//
// IMPORTANT: This struct is request-scoped and should NOT be stored beyond request lifecycle.
type RequestScopedDataPermApplier struct {
	principal *Principal
	dataScope DataScope
	logger    log.Logger
}

// NewRequestScopedDataPermApplier creates a new request-scoped data permission applier.
// This function is typically called by the data permission middleware for each request.
func NewRequestScopedDataPermApplier(
	principal *Principal,
	dataScope DataScope,
	logger log.Logger,
) DataPermissionApplier {
	return &RequestScopedDataPermApplier{
		principal: principal,
		dataScope: dataScope,
		logger:    logger,
	}
}

// Apply implements security.DataPermissionApplier.Apply.
func (a *RequestScopedDataPermApplier) Apply(query orm.SelectQuery) error {
	// No data scope means no restrictions
	if a.dataScope == nil {
		a.logger.Debugf("No data scope configured, skipping data permission")

		return nil
	}

	// Get the main table from the query
	// The query MUST have called Model() before this point
	queryBuilder, ok := query.(orm.QueryBuilder)
	if !ok {
		return ErrQueryNotQueryBuilder
	}

	table := queryBuilder.GetTable()
	if table == nil {
		return ErrQueryModelNotSet
	}

	// Check if the data scope supports this table
	if !a.dataScope.Supports(a.principal, table) {
		a.logger.Debugf(
			"Data scope %q is not applicable to table %q, skipping data permission",
			a.dataScope.Key(), table.TypeName)

		return nil
	}

	// Apply the data scope
	if err := a.dataScope.Apply(a.principal, query); err != nil {
		return fmt.Errorf("failed to apply data scope %q: %w", a.dataScope.Key(), err)
	}

	a.logger.Debugf("Applied data scope %q to table %q", a.dataScope.Key(), table.TypeName)

	return nil
}
