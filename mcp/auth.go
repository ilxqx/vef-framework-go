package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/auth"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/security"
)

// GetPrincipalFromContext extracts the Principal from MCP request context.
func GetPrincipalFromContext(ctx context.Context) *security.Principal {
	tokenInfo := auth.TokenInfoFromContext(ctx)
	if tokenInfo != nil && tokenInfo.Extra != nil {
		if principal, ok := tokenInfo.Extra["principal"].(*security.Principal); ok {
			return principal
		}
	}

	return security.PrincipalAnonymous
}

// DbWithOperator returns a database connection with the operator ID bound from the MCP context.
func DbWithOperator(ctx context.Context, db orm.Db) orm.Db {
	if principal := GetPrincipalFromContext(ctx); principal != nil {
		return db.WithNamedArg(constants.PlaceholderKeyOperator, principal.Id)
	}

	return db
}
