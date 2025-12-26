package tools

import (
	"context"
	"encoding/json"
	"unicode/utf8"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/encoding"
	"github.com/ilxqx/vef-framework-go/mcp"
	"github.com/ilxqx/vef-framework-go/orm"
)

// QueryArgs defines the expected arguments for the database_query tool.
type QueryArgs struct {
	Sql    string `json:"sql" jsonschema:"required,description=The SQL query with placeholders (?) for parameters"`
	Params []any  `json:"params,omitempty" jsonschema:"description=Parameters for the SQL query placeholders"`
}

// QueryTool provides MCP tool for executing parameterized SQL queries.
type QueryTool struct {
	db orm.Db
}

// NewQueryTool creates a new QueryTool instance.
func NewQueryTool(db orm.Db) mcp.ToolProvider {
	return &QueryTool{db: db}
}

// Tools implements mcp.ToolProvider.
func (t *QueryTool) Tools() []mcp.ToolDefinition {
	return []mcp.ToolDefinition{
		{
			Tool: &mcp.Tool{
				Name:        "database_query",
				Description: "Execute a parameterized SQL query against the database. Returns query results as JSON array.",
				InputSchema: mcp.MustSchemaFor[QueryArgs](),
			},
			Handler: t.handleQuery,
		},
	}
}

// handleQuery executes a parameterized SQL query.
func (t *QueryTool) handleQuery(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var (
		db         = mcp.DbWithOperator(ctx, t.db)
		args       QueryArgs
		results    []map[string]any
		jsonResult string
		err        error
	)

	if err = json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return mcp.NewToolResultError("Failed to parse arguments: " + err.Error()), nil
	}
	if args.Sql == constants.Empty {
		return mcp.NewToolResultError("Sql parameter is required and must not be empty"), nil
	}
	if err = db.NewRaw(args.Sql, args.Params...).Scan(ctx, &results); err != nil {
		return mcp.NewToolResultError("Query execution failed: " + err.Error()), nil
	}

	// Convert []byte to string to avoid Base64 encoding in JSON
	for _, result := range results {
		convertByteSlicesToStrings(result)
	}
	if jsonResult, err = encoding.ToJson(results); err != nil {
		return mcp.NewToolResultError("Failed to encode results: " + err.Error()), nil
	}

	return mcp.NewToolResultText(jsonResult), nil
}

// convertByteSlicesToStrings converts []byte values to strings in a map.
// Only converts if the byte slice is valid UTF-8 text (e.g., PostgreSQL char/varchar fields).
// Binary data (e.g., BYTEA/BLOB) remains as []byte and will be Base64-encoded in JSON.
func convertByteSlicesToStrings(m map[string]any) {
	for k, v := range m {
		switch val := v.(type) {
		case []byte:
			if utf8.Valid(val) {
				m[k] = string(val)
			}
		case map[string]any:
			convertByteSlicesToStrings(val)
		}
	}
}
