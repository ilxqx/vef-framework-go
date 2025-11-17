# VEF Framework Go

📖 [English](./README.md) | [简体中文](./README.zh-CN.md)

[![GitHub Release](https://img.shields.io/github/v/release/ilxqx/vef-framework-go)](https://github.com/ilxqx/vef-framework-go/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/ilxqx/vef-framework-go/test.yml?branch=main)](https://github.com/ilxqx/vef-framework-go/actions/workflows/test.yml)
[![Coverage](https://img.shields.io/codecov/c/github/ilxqx/vef-framework-go)](https://codecov.io/gh/ilxqx/vef-framework-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/ilxqx/vef-framework-go.svg)](https://pkg.go.dev/github.com/ilxqx/vef-framework-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/ilxqx/vef-framework-go)](https://goreportcard.com/report/github.com/ilxqx/vef-framework-go)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/ilxqx/vef-framework-go)
[![License](https://img.shields.io/github/license/ilxqx/vef-framework-go)](https://github.com/ilxqx/vef-framework-go/blob/main/LICENSE)

A modern Go web development framework built on Uber FX dependency injection and Fiber, designed for rapid enterprise application development with opinionated conventions and comprehensive built-in features.

## Features

- **Single-Endpoint Api Architecture** - All Api requests through `POST /api` with unified request/response format
- **Generic CRUD Apis** - Pre-built type-safe CRUD operations with minimal boilerplate
- **Type-Safe ORM** - Bun-based ORM with fluent query builder and automatic audit tracking
- **Multi-Strategy Authentication** - Jwt, OpenApi signature, and password authentication out of the box
- **Modular Design** - Uber FX dependency injection with pluggable modules
- **Built-in Features** - Cache, event bus, cron scheduler, object storage, data validation, i18n
- **RBAC & Data Permissions** - Row-level security with customizable data scopes

## Quick Start

### Installation

```bash
go get github.com/ilxqx/vef-framework-go
```

**Requirements:** Go 1.25.0 or higher

**Troubleshooting:** If you encounter ambiguous import errors with `google.golang.org/genproto` during `go mod tidy`, run:

```bash
go get google.golang.org/genproto@latest
go mod tidy
```

### Minimal Example

Create `main.go`:

```go
package main

import "github.com/ilxqx/vef-framework-go"

func main() {
    vef.Run()
}
```

Create `configs/application.toml`:

```toml
[vef.app]
name = "my-app"
port = 8080

[vef.datasource]
type = "postgres"
host = "localhost"
port = 5432
user = "postgres"
password = "password"
database = "mydb"
schema = "public"
```

Run the application:

```bash
go run main.go
```

Your Api server is now running at `http://localhost:8080`.

## Project Structure

### Recommended Module Organization

VEF Framework applications follow a modular architecture pattern where business domains are organized into self-contained modules. This pattern is demonstrated in production applications and provides clear separation of concerns.

**Directory Structure:**

```
my-app/
├── cmd/
│   └── server/
│       └── main.go           # Application entry - composes all modules
├── configs/
│   └── application.toml       # Configuration file
└── internal/
    ├── auth/                  # Authentication providers
    │   ├── module.go          # Auth module definition
    │   ├── user_loader.go     # UserLoader implementation
    │   └── user_info_loader.go
    ├── sys/                   # System/admin features
    │   ├── models/            # Data models
    │   ├── payloads/          # API parameters
    │   ├── resources/         # API resources
    │   ├── schemas/           # Generated from models (via vef-cli)
    │   └── module.go          # System module definition
    ├── [domain]/              # Business domains (e.g., order, inventory)
    │   ├── models/
    │   ├── payloads/
    │   ├── resources/
    │   ├── schemas/
    │   └── module.go
    ├── vef/                   # VEF framework integrations
    │   ├── module.go
    │   ├── build_info.go      # Generated build metadata
    │   ├── *_subscriber.go    # Event subscribers
    │   └── *_loader.go        # Data loaders
    └── web/                   # SPA frontend integration (optional)
        ├── dist/              # Static assets
        └── module.go
```

### Module Composition

Each module exports a `vef.Module()` that encapsulates its dependencies and resources. The main.go composes these modules in dependency order:

```go
package main

import (
    "github.com/ilxqx/vef-framework-go"
    "my-app/internal/auth"
    "my-app/internal/sys"
    ivef "my-app/internal/vef"
    "my-app/internal/web"
)

func main() {
    vef.Run(
        ivef.Module,     // Framework integrations (your app's vef module)
        web.Module,      // SPA serving (optional)
        auth.Module,     // Authentication providers
        sys.Module,      // System resources
        // Add your business domain modules here
    )
}
```

**Module Definition Example:**

```go
// internal/sys/module.go
package sys

import (
    "github.com/ilxqx/vef-framework-go"
    "my-app/internal/sys/resources"
)

var Module = vef.Module(
    "app:sys",
    vef.ProvideApiResource(resources.NewUserResource),
    vef.ProvideApiResource(resources.NewRoleResource),
    // Register other resources and services
)
```

**Benefits of this pattern:**
- **Clear boundaries**: Each module owns its models, APIs, and business logic
- **Testability**: Modules can be tested independently
- **Scalability**: Easy to add new domains without affecting existing code
- **Maintainability**: Changes are localized to specific modules

## Architecture

### Single-Endpoint Design

VEF uses a single-endpoint approach where all Api requests go through `POST /api` (or `POST /openapi` for external integrations).

**Request Format:**

```json
{
  "resource": "sys/user",
  "action": "find_page",
  "version": "v1",
  "params": {
    "keyword": "john"
  },
  "meta": {
    "page": 1,
    "size": 20
  }
}
```

**Response Format:**

```json
{
  "code": 0,
  "message": "Success",
  "data": {
    "page": 1,
    "size": 20,
    "total": 100,
    "items": [...]
  }
}
```

Params vs Meta:
- `params` carries business data (e.g., search filters, create/update fields). Define your structs embedding `api.P`.
- `meta` carries request-level options (e.g., pagination for `find_page`, export/import format). Define your structs embedding `api.M` (e.g., `page.Pageable`).

### Dependency Injection

VEF leverages Uber FX for dependency injection. Register components using helper functions:

```go
vef.Run(
    vef.ProvideApiResource(NewUserResource),
    vef.Provide(NewUserService),
)
```

## Defining Models

All models should embed `orm.Model` for automatic audit field management:

```go
package models

import (
    "github.com/ilxqx/vef-framework-go/null"
    "github.com/ilxqx/vef-framework-go/orm"
)

type User struct {
    orm.BaseModel `bun:"table:sys_user,alias:su"`
    orm.Model     
    
    Username string      `json:"username" validate:"required,alphanum,max=32" label:"Username"`
    Email    null.String `json:"email" validate:"omitempty,email,max=64" label:"Email"`
    IsActive bool        `json:"isActive"`
}
```

**Field Tags:**

- `bun` - Bun ORM configuration (table name, column mapping, relations)
- `json` - JSON serialization name
- `validate` - Validation rules ([go-playground/validator](https://github.com/go-playground/validator))
- `label` - Human-readable field name for error messages

**Audit Fields** (automatically maintained by `orm.Model`):

- `id` - Primary key (20-character XID in base32 encoding)
- `created_at`, `created_by` - Creation timestamp and user ID
- `created_by_name` - Creator name (scan-only, not stored in database)
- `updated_at`, `updated_by` - Last update timestamp and user ID
- `updated_by_name` - Updater name (scan-only, not stored in database)

Note: Database columns use snake_case (e.g., `created_at`), while JSON fields use camelCase (e.g., `createdAt`) as shown in the model tags.

**Null Types:** Use `null.String`, `null.Int`, `null.Bool`, etc. for nullable fields.

### Field Types for Boolean Columns

Choosing the right type depends on your target database and whether you need tri‑state (NULL) semantics.

Key guidance:
- Prefer `bool` in most cases. Modern mainstream databases natively support boolean types, and plain `bool` maps well.
- Use `sql.Bool` when you need to store booleans as numeric types (e.g., tinyint/smallint with 0/1) for databases that lack native boolean or when you explicitly require numeric storage for compatibility.
- Use `null.Bool` when you need tri‑state: NULL, false, true. It serializes to database values as NULL/1/0.

Decision guide:

| Use Case | Preferred Type | Database Column |
|----------|----------------|-----------------|
| Non‑nullable boolean on DBs with native boolean | `bool` | boolean/true native type |
| Nullable boolean (tri‑state) | `null.Bool` | boolean or numeric (often smallint/tinyint) |
| Target DB without native boolean or require numeric 0/1 storage | `sql.Bool` (non‑null) / `null.Bool` (nullable) | smallint/tinyint with 0/1 |
| Go‑only/computed (not stored) | `bool` with `bun:"-"` | N/A |

Type details and examples:

1) Plain `bool` — recommended for native boolean columns
```go
type User struct {
    orm.Model
    // Database: boolean (native), NOT NULL as needed
    IsActive bool `json:"isActive"` // bun tag usually not required when using native boolean
}
```

2) `sql.Bool` — numeric 0/1 storage for compatibility
```go
import "github.com/ilxqx/vef-framework-go/sql"

type User struct {
    orm.Model
    // Database: numeric boolean (0/1), for DBs without native boolean or enforced numeric schema
    IsActive sql.Bool `json:"isActive" bun:"type:smallint,notnull,default:0"`
    IsLocked sql.Bool `json:"isLocked" bun:"type:smallint,notnull,default:0"`
}
```
When you don’t have to support non‑boolean databases, prefer plain `bool` for simplicity.

3) `null.Bool` — tri‑state (NULL/false/true)
```go
import "github.com/ilxqx/vef-framework-go/null"

type User struct {
    orm.Model
    // Database: allows NULL; stored as NULL/0/1 (use numeric column for maximum compatibility)
    IsVerified null.Bool `json:"isVerified" bun:"type:smallint"`
}
```
Three‑state logic:
- `null.Bool{Valid: false}` → NULL in database
- `null.Bool{Valid: true, Bool: false}` → 0/false
- `null.Bool{Valid: true, Bool: true}` → 1/true

4) Go‑only fields (not stored)
```go
type User struct {
    orm.Model
    Username string `json:"username"`

    // Computed field — not stored in database
    HasPermissions bool `json:"hasPermissions" bun:"-"`
}
```

Common patterns:
```go
// Native boolean DBs (recommended)
type UserNative struct {
    orm.Model
    IsActive bool      `json:"isActive"`
    IsLocked bool      `json:"isLocked"`
    IsEmailVerified null.Bool `json:"isEmailVerified"` // use NULL when needed
}

// Numeric storage for compatibility
type UserNumeric struct {
    orm.Model
    IsActive sql.Bool       `json:"isActive" bun:"type:smallint,notnull,default:0"`
    IsLocked sql.Bool       `json:"isLocked" bun:"type:smallint,notnull,default:0"`
    IsEmailVerified null.Bool `json:"isEmailVerified" bun:"type:smallint"`
}
```

## Building CRUD Apis

### Resource Naming Best Practices

When defining API resources, follow a consistent naming convention to avoid conflicts and make API ownership clear.

**Recommended Pattern: `{app}/{domain}/{entity}`**

This three-level namespace pattern is used in production applications and provides several benefits:

```go
// Good examples with application namespace
api.NewResource("smp/sys/user")           // System user resource
api.NewResource("smp/md/organization")    // Master data organization
api.NewResource("erp/order/item")         // Clear domain separation

// Acceptable for single-app projects
api.NewResource("sys/user")               // No app namespace

// Avoid - too generic, risks conflicts
api.NewResource("user")                   // ❌ No namespace
```

**Benefits of Application Namespacing:**

- **Conflict Prevention**: Avoids API resource collisions in shared deployments or when merging codebases
- **Clear Ownership**: Immediately identifies which application owns the resource
- **Modularity**: Supports multiple applications or microservices using the same framework
- **Migration Safety**: Easy to identify and migrate resources when restructuring

**Framework Reserved Namespaces:**

The following resource namespaces are reserved for system APIs and must not be used in custom API definitions:

- `security/auth` - Authentication APIs
- `sys/storage` - Storage APIs
- `sys/monitor` - Monitoring APIs

Using these reserved names will cause application startup failures due to duplicate API definitions.

### Step 1: Define Parameter Structures

**Search Parameters:**

```go
package payloads

import "github.com/ilxqx/vef-framework-go/api"

type UserSearch struct {
    api.P
    Keyword string `json:"keyword" search:"contains,column=username|email"`
    IsActive *bool `json:"isActive" search:"eq"`
}
```

**Create/Update Parameters:**

```go
type UserParams struct {
    api.P
    Id       string      `json:"id"` // Required for updates

    Username string      `json:"username" validate:"required,alphanum,max=32" label:"Username"`
    Email    null.String `json:"email" validate:"omitempty,email,max=64" label:"Email"`
    IsActive bool        `json:"isActive"`
}
```

**Separate Create and Update Parameters:**

When Create and Update operations have different validation requirements, use struct embedding to share common fields while allowing operation-specific validation:

```go
// Shared fields
type UserParams struct {
    api.P
    Id       string
    Username string      `json:"username" validate:"required,alphanum,max=32" label:"Username"`
    Email    null.String `json:"email" validate:"omitempty,email,max=64" label:"Email"`
    IsActive bool        `json:"isActive"`
}

// Create requires password
type UserCreateParams struct {
    UserParams      `json:",inline"`
    Password        string `json:"password" validate:"required,min=6,max=16" label:"Password"`
    PasswordConfirm string `json:"passwordConfirm" validate:"required,eqfield=Password" label:"Confirm Password"`
}

// Update has optional password
type UserUpdateParams struct {
    UserParams      `json:",inline"`
    Password        null.String `json:"password" validate:"omitempty,min=6,max=16" label:"Password"`
    PasswordConfirm null.String `json:"passwordConfirm" validate:"omitempty,eqfield=Password" label:"Confirm Password"`
}
```

Then use the specific params in your resource:

```go
CreateApi: apis.NewCreateApi[models.User, payloads.UserCreateParams](),
UpdateApi: apis.NewUpdateApi[models.User, payloads.UserUpdateParams](),
```

**Benefits:**
- **Type-safe validation**: Different rules for Create vs Update (required vs optional password)
- **Clear contracts**: API requirements are explicit in code
- **Better error messages**: Validation errors match the operation's actual requirements
- **Code reuse**: Common fields are defined once and embedded

### Step 2: Create Api Resource

> **⚠️ IMPORTANT: Reserved System API Namespaces**
>
> The framework reserves the following resource namespaces for system APIs. **DO NOT** use these resource names in your custom API definitions, as they will conflict with built-in framework functionality and cause application startup failures:
>
> - `security/auth` - Authentication APIs (login, logout, refresh, get_user_info)
> - `sys/storage` - Storage APIs (upload, get_presigned_url, delete_temp, stat, list)
> - `sys/monitor` - Monitoring APIs (get_overview, get_cpu, get_memory, get_disk, etc.)
>
> The framework automatically detects duplicate API definitions and will fail to start if conflicts are found. Use custom resource namespaces like `app/`, `custom/`, or your own domain-specific prefixes to avoid conflicts.

```go
package resources

import (
    "github.com/ilxqx/vef-framework-go/api"
    "github.com/ilxqx/vef-framework-go/apis"
)

type UserResource struct {
    api.Resource
    apis.FindAllApi[models.User, payloads.UserSearch]
    apis.FindPageApi[models.User, payloads.UserSearch]
    apis.CreateApi[models.User, payloads.UserParams]
    apis.UpdateApi[models.User, payloads.UserParams]
    apis.DeleteApi[models.User]
}

func NewUserResource() api.Resource {
    return &UserResource{
        Resource: api.NewResource("smp/sys/user"),  // ✓ Use app/domain/entity to avoid conflicts
        FindAllApi: apis.NewFindAllApi[models.User, payloads.UserSearch](),
        FindPageApi: apis.NewFindPageApi[models.User, payloads.UserSearch](),
        CreateApi: apis.NewCreateApi[models.User, payloads.UserParams](),
        UpdateApi: apis.NewUpdateApi[models.User, payloads.UserParams](),
        DeleteApi: apis.NewDeleteApi[models.User](),
    }
}
```

### Step 3: Register Resource

```go
func main() {
    vef.Run(
        vef.ProvideApiResource(resources.NewUserResource),
    )
}
```

### Pre-built Apis

| Api | Description | Action |
|-----|-------------|--------|
| FindOneApi | Find single record | find_one |
| FindAllApi | Find all records | find_all |
| FindPageApi | Paginated query | find_page |
| CreateApi | Create record | create |
| UpdateApi | Update record | update |
| DeleteApi | Delete record | delete |
| CreateManyApi | Batch create | create_many |
| UpdateManyApi | Batch update | update_many |
| DeleteManyApi | Batch delete | delete_many |
| FindTreeApi | Hierarchical query | find_tree |
| FindOptionsApi | Options list (label/value) | find_options |
| FindTreeOptionsApi | Tree options | find_tree_options |
| ImportApi | Import from Excel/CSV | import |
| ExportApi | Export to Excel/CSV | export |

### Api Builder Methods

Configure Api behavior with fluent builder methods:

```go
CreateApi: apis.NewCreateApi[User, UserParams]().
    Action("create_user").             // Custom action name
    Public().                          // No authentication required
    PermToken("sys.user.create").      // Permission token
    EnableAudit().                     // Enable audit logging
    Timeout(10 * time.Second).         // Request timeout
    RateLimit(10, 1*time.Minute).      // 10 requests per minute
```

**Note:** FindApi types (FindOneApi, FindAllApi, FindPageApi, FindTreeApi, FindOptionsApi, FindTreeOptionsApi, ExportApi) have additional configuration methods. See [FindApi Configuration Methods](#findapi-configuration-methods) for details.

### FindApi Configuration Methods

All FindApi types (FindOneApi, FindAllApi, FindPageApi, FindTreeApi, FindOptionsApi, FindTreeOptionsApi, ExportApi) support a unified query configuration system using fluent methods. These methods allow you to customize query behavior, add conditions, configure sorting, and process results.

#### Common Configuration Methods

| Method | Description | Default QueryPart | Applicable APIs |
|--------|-------------|-------------------|-----------------|
| `WithProcessor` | Set post-processing function for query results | N/A | All FindApi |
| `WithOptions` | Add multiple FindApiOptions | N/A | All FindApi |
| `WithSelect` | Add column to SELECT clause | QueryRoot | All FindApi |
| `WithSelectAs` | Add column with alias to SELECT clause | QueryRoot | All FindApi |
| `WithDefaultSort` | Set default sorting specifications | QueryRoot | All FindApi |
| `WithCondition` | Add WHERE condition using ConditionBuilder | QueryRoot | All FindApi |
| `WithRelation` | Add relation join | QueryRoot | All FindApi |
| `WithAuditUserNames` | Fetch audit user names (created_by_name, updated_by_name) | QueryRoot | All FindApi |
| `WithQueryApplier` | Add custom query applier function | QueryRoot | All FindApi |
| `DisableDataPerm` | Disable data permission filtering | N/A | All FindApi |

**WithProcessor Example:**

The `Processor` function is executed after the database query completes but before returning results to the client. This allows you to transform, enrich, or filter the query results.

Common use cases:
- **Data masking**: Hide sensitive information (passwords, tokens)
- **Computed fields**: Add calculated values based on existing data
- **Nested structure transformation**: Convert flat data to hierarchical structures
- **Aggregation**: Compute statistics or summaries

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithProcessor(func(users []User, search UserSearch, ctx fiber.Ctx) any {
        // Data masking
        for i := range users {
            users[i].Password = "***"
            users[i].ApiToken = ""
        }
        return users
    }),

// Example: Adding computed fields in paged results (processor receives items slice)
FindPageApi: apis.NewFindPageApi[Order, OrderSearch]().
    WithProcessor(func(items []Order, search OrderSearch, ctx fiber.Ctx) any {
        for i := range items {
            // Calculate total amount
            items[i].TotalAmount = items[i].Quantity * items[i].UnitPrice
        }
        return items
    }),

// Example: Nested structure transformation
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithProcessor(func(users []User, search UserSearch, ctx fiber.Ctx) any {
        // Group users by department
        type DepartmentUsers struct {
            DepartmentName string `json:"departmentName"`
            Users          []User `json:"users"`
        }
        
        grouped := make(map[string]*DepartmentUsers)
        for _, user := range users {
            if _, exists := grouped[user.DepartmentId]; !exists {
                grouped[user.DepartmentId] = &DepartmentUsers{
                    DepartmentName: user.DepartmentName,
                    Users:          []User{},
                }
            }
            grouped[user.DepartmentId].Users = append(grouped[user.DepartmentId].Users, user)
        }
        
        result := make([]DepartmentUsers, 0, len(grouped))
        for _, dept := range grouped {
            result = append(result, *dept)
        }
        return result
    }),
```

**WithSelect / WithSelectAs Example:**

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithSelect("username").
    WithSelectAs("email_address", "email"),
```

**WithDefaultSort Example:**

```go
FindPageApi: apis.NewFindPageApi[User, UserSearch]().
    WithDefaultSort(&sort.OrderSpec{
        Column:    "created_at",
        Direction: sort.OrderDesc,
    }),

// Production pattern: Use schema-generated column names for type safety
import "my-app/internal/sys/schemas"

FindPageApi: apis.NewFindPageApi[User, UserSearch]().
    WithDefaultSort(&sort.OrderSpec{
        Column:    schemas.User.CreatedAt(true), // Type-safe column with table prefix
        Direction: sort.OrderDesc,
    }),

// For tree structures, use sort_order field
FindTreeApi: apis.NewFindTreeApi[Menu, MenuSearch](buildMenuTree).
    WithDefaultSort(&sort.OrderSpec{
        Column:    schemas.Menu.SortOrder(true),
        Direction: sort.OrderAsc,
    }),
```

Pass empty arguments to disable default sorting:

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithDefaultSort(), // Disable default sorting
```

**WithCondition Example:**

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithCondition(func(cb orm.ConditionBuilder) {
        cb.Equals("is_deleted", false)
        cb.Equals("is_active", true)
    }),
```

**WithRelation Example:**

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithRelation(&orm.RelationSpec{
        // Join the Profile model; foreign/referenced keys are auto-resolved
        Model: (*Profile)(nil),
        // Optional: customize alias/columns
        // Alias: "p",
        SelectedColumns: []orm.ColumnInfo{
            {Name: "name", AutoAlias: true},
            {Name: "email", AutoAlias: true},
        },
    }),
```

**WithAuditUserNames Example:**

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithAuditUserNames(&User{}), // Uses "name" column by default

// Or specify custom column name
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithAuditUserNames(&User{}, "username"),

// Production pattern: Use package-level model instance
// In models package: var UserModel = &User{}
FindPageApi: apis.NewFindPageApi[User, UserSearch]().
    WithAuditUserNames(models.UserModel), // Recommended for consistency
```

**WithQueryApplier Example:**

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithQueryApplier(func(query orm.SelectQuery, search UserSearch, ctx fiber.Ctx) error {
        // Custom query logic
        if search.IncludeInactive {
            query.Where(func(cb orm.ConditionBuilder) {
                cb.Or(
                    cb.Equals("is_active", true),
                    cb.Equals("is_active", false),
                )
            })
        }
        return nil
    }),
```

**DisableDataPerm Example:**

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    DisableDataPerm(), // Must be called before API registration
```

**Important:** `DisableDataPerm()` must be called before the API is registered (before the `Setup` method is executed). It should be chained immediately after `NewFindXxxApi()`. By default, data permission filtering is enabled and automatically applied during `Setup`.

#### QueryPart System

The `parts` parameter in configuration methods specifies which part(s) of the query the option applies to. This is particularly important for tree APIs that use recursive CTEs (Common Table Expressions).

| QueryPart | Description | Use Case |
|-----------|-------------|----------|
| `QueryRoot` | Outer/root query | Sorting, limiting, final filtering |
| `QueryBase` | Base query (in CTE) | Initial conditions, starting nodes |
| `QueryRecursive` | Recursive query (in CTE) | Recursive traversal configuration |
| `QueryAll` | All query parts | Column selection, relations |

**Default Behavior:**

- `WithSelect`, `WithSelectAs`, `WithRelation`: Default to `QueryRoot` (applies to the main/root query)
- `WithCondition`, `WithQueryApplier`, `WithDefaultSort`: Default to `QueryRoot` (applies to root query only)

**Normal Query Example:**

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithSelect("username").              // Applies to QueryRoot (main query)
    WithCondition(func(cb orm.ConditionBuilder) {
        cb.Equals("is_active", true)     // Applies to QueryRoot (main query)
    }),
```

**Tree Query Example:**

```go
FindTreeApi: apis.NewFindTreeApi[Category, CategorySearch](buildTree).
    // Select columns for both base and recursive queries
    WithSelect("sort", apis.QueryBase, apis.QueryRecursive).
    
    // Filter only starting nodes
    WithCondition(func(cb orm.ConditionBuilder) {
        cb.IsNull("parent_id")           // Only applies to QueryBase
    }, apis.QueryBase).
    
    // Add condition to recursive traversal
    WithCondition(func(cb orm.ConditionBuilder) {
        cb.Equals("is_active", true)     // Applies to QueryRecursive
    }, apis.QueryRecursive),
```

#### Tree Query Configuration

`FindTreeApi` and `FindTreeOptionsApi` use recursive CTEs (Common Table Expressions) to query hierarchical data. Understanding how QueryPart applies to different parts of the recursive query is essential for proper configuration.

**Recursive CTE Structure:**

```sql
WITH RECURSIVE tree AS (
    -- QueryBase: Initial query for root nodes
    SELECT * FROM categories WHERE parent_id IS NULL
    
    UNION ALL
    
    -- QueryRecursive: Recursive query joining with CTE
    SELECT c.* FROM categories c
    INNER JOIN tree t ON c.parent_id = t.id
)
-- QueryRoot: Final SELECT from CTE
SELECT * FROM tree ORDER BY sort
```

**QueryPart Behavior in Tree Queries:**

- `WithSelect` / `WithSelectAs`: Default to `QueryBase` and `QueryRecursive` (columns must be consistent in both parts of UNION)
- `WithCondition` / `WithQueryApplier`: Default to `QueryBase` only (filter starting nodes)
- `WithRelation`: Default to `QueryBase` and `QueryRecursive` (joins needed in both parts)
- `WithDefaultSort`: Applies to `QueryRoot` (sort final results)

**Complete Tree Query Example:**

```go
FindTreeApi: apis.NewFindTreeApi[Category, CategorySearch](
    func(categories []Category) []Category {
        // Build tree structure from flat list
        return buildCategoryTree(categories)
    },
).
    // Add custom columns to both base and recursive queries
    WithSelect("sort", apis.QueryBase, apis.QueryRecursive).
    WithSelect("icon", apis.QueryBase, apis.QueryRecursive).
    
    // Filter starting nodes (only active root categories)
    WithCondition(func(cb orm.ConditionBuilder) {
        cb.Equals("is_active", true)
        cb.IsNull("parent_id")
    }, apis.QueryBase).
    
    // Add relation to both queries
    WithRelation(&orm.RelationSpec{
        Model: (*Metadata)(nil),
        SelectedColumns: []orm.ColumnInfo{
            {Name: "icon", AutoAlias: true},
            {Name: "sort_order", Alias: "sortOrder"},
        },
    }, apis.QueryBase, apis.QueryRecursive).
    
    // Fetch audit user names
    WithAuditUserNames(&User{}).
    
    // Sort final results
    WithDefaultSort(&sort.OrderSpec{
        Column:    "sort",
        Direction: sort.OrderAsc,
    }),
```

**FindTreeOptionsApi Configuration:**

`FindTreeOptionsApi` follows the same configuration pattern as `FindTreeApi`:

```go
FindTreeOptionsApi: apis.NewFindTreeOptionsApi[Category, CategorySearch]().
    WithDefaultColumnMapping(&apis.DataOptionColumnMapping{
        LabelColumn: "name",
        ValueColumn: "id",
    }).
    WithIdColumn("id").
    WithParentIdColumn("parent_id").
    WithCondition(func(cb orm.ConditionBuilder) {
        cb.Equals("is_active", true)
    }, apis.QueryBase),
```

#### API-Specific Configuration Methods

**FindPageApi:**

```go
FindPageApi: apis.NewFindPageApi[User, UserSearch]().
    WithDefaultPageSize(20), // Set default page size (used when request doesn't specify or is invalid)
```

**FindOptionsApi:**

```go
FindOptionsApi: apis.NewFindOptionsApi[User, UserSearch]().
    WithDefaultColumnMapping(&apis.DataOptionColumnMapping{
        LabelColumn:       "name",        // Column for option label (default: "name")
        ValueColumn:       "id",          // Column for option value (default: "id")
        DescriptionColumn: "description", // Optional description column
    }),

// Advanced: Include additional metadata in options
FindOptionsApi: apis.NewFindOptionsApi[Menu, MenuSearch]().
    WithDefaultColumnMapping(&apis.DataOptionColumnMapping{
        LabelColumn:       "name",
        ValueColumn:       "id",
        DescriptionColumn: "remark",
        MetaColumns: []string{
            "type",                  // Menu type (D=Directory, M=Menu, B=Button)
            "icon",                  // Icon identifier
            "sort_order AS sortOrder", // Display order with alias
        },
    }),
```

**FindTreeApi:**

For hierarchical data structures, use `FindTreeApi` with the `treebuilder` package to convert flat database results into nested tree structures:

```go
import "github.com/ilxqx/vef-framework-go/treebuilder"

FindTreeApi: apis.NewFindTreeApi[models.Organization, payloads.OrganizationSearch](
    buildOrganizationTree,
).
    WithIdColumn("id").              // ID column name (default: "id")
    WithParentIdColumn("parent_id"). // Parent ID column name (default: "parent_id")
    WithDefaultSort(&sort.OrderSpec{
        Column:    "sort_order",
        Direction: sort.OrderAsc,
    })

func buildOrganizationTree(flatModels []models.Organization) []models.Organization {
    return treebuilder.Build(
        flatModels,
        treebuilder.Adapter[models.Organization]{
            GetId:       func(m models.Organization) string { return m.Id },
            GetParentId: func(m models.Organization) string { return m.ParentId.ValueOrZero() },
            SetChildren: func(m *models.Organization, children []models.Organization) {
                m.Children = children
            },
        },
    )
}
```

**Model Requirements:**

Your model must have:
- A parent ID field (typically `null.String` to support root nodes)
- A children field (slice of same model type, marked with `bun:"-"` since it's computed)

```go
type Organization struct {
    orm.Model
    Name     string          `json:"name"`
    ParentId null.String     `json:"parentId" bun:"type:varchar(20)"` // NULL for root nodes
    Children []Organization  `json:"children" bun:"-"`                // Computed, not in DB
}
```

The `treebuilder.Build` function handles the conversion from flat list to hierarchical structure, properly nesting children under their parents.

**FindTreeOptionsApi:**

Combines both options and tree configuration to return hierarchical option lists:

```go
FindTreeOptionsApi: apis.NewFindTreeOptionsApi[models.Organization, payloads.OrganizationSearch]().
    WithDefaultColumnMapping(&apis.DataOptionColumnMapping{
        LabelColumn: "name",
        ValueColumn: "id",
    }).
    WithIdColumn("id").
    WithParentIdColumn("parent_id").
    WithDefaultSort(&sort.OrderSpec{
        Column:    "sort_order",
        Direction: sort.OrderAsc,
    })
```

The tree options API automatically uses the internal tree builder to convert flat results into nested option structures, perfect for cascading selectors or hierarchical menus.

**ExportApi:**

```go
ExportApi: apis.NewExportApi[User, UserSearch]().
    WithDefaultFormat("excel").                   // Default export format: "excel" or "csv"
    WithExcelOptions(&excel.ExportOptions{        // Excel-specific options
        SheetName: "Users",
    }).
    WithCsvOptions(&csv.ExportOptions{            // CSV-specific options
        Delimiter: ',',
    }).
    WithPreExport(func(users []User, search UserSearch, ctx fiber.Ctx, db orm.Db) error {
        // Modify data before export (e.g., data masking)
        for i := range users {
            users[i].Password = "***"
        }
        return nil
    }).
    WithFilenameBuilder(func(search UserSearch, ctx fiber.Ctx) string {
        // Generate dynamic filename
        return fmt.Sprintf("users_%s", time.Now().Format("20060102"))
    }),
```

### Pre/Post Hooks

Add custom business logic before/after CRUD operations:

```go
CreateApi: apis.NewCreateApi[User, UserParams]().
    WithPreCreate(func(model *User, params *UserParams, ctx fiber.Ctx, db orm.Db) error {
        // Hash password before creating user
        hashed, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
        if err != nil {
            return err
        }
        model.Password = string(hashed)
        return nil
    }).
    WithPostCreate(func(model *User, params *UserParams, ctx fiber.Ctx, tx orm.Db) error {
        // Send welcome email after user creation (within transaction)
        return sendWelcomeEmail(model.Email)
    }),
```

Available hooks:

**Single Record Operations:**

- `WithPreCreate`, `WithPostCreate` - Before/after creation (`WithPostCreate` runs in transaction)
- `WithPreUpdate`, `WithPostUpdate` - Before/after update (receives both old and new model, `WithPostUpdate` runs in transaction)
- `WithPreDelete`, `WithPostDelete` - Before/after deletion (`WithPostDelete` runs in transaction)

**Batch Operations:**

- `WithPreCreateMany`, `WithPostCreateMany` - Before/after batch creation (`WithPostCreateMany` runs in transaction)
- `WithPreUpdateMany`, `WithPostUpdateMany` - Before/after batch update (receives old and new model arrays, `WithPostUpdateMany` runs in transaction)
- `WithPreDeleteMany`, `WithPostDeleteMany` - Before/after batch deletion (`WithPostDeleteMany` runs in transaction)

**Import/Export Operations:**

- `WithPreImport`, `WithPostImport` - Before/after import (`WithPreImport` for validation, `WithPostImport` runs in transaction)
- `WithPreExport` - Before export (for data formatting)

**Production Patterns:**

```go
// System user protection - Prevent deletion of critical system users
DeleteApi: apis.NewDeleteApi[User]().
    WithPreDelete(func(model *User, ctx fiber.Ctx, db orm.Db) error {
        // Protect system-internal users from deletion
        switch model.Username {
        case "system", "anonymous", "cron":
            return result.Err("Cannot delete system internal user")
        }
        return nil
    }),

// Conditional password hashing - Only hash if password is being changed
UpdateApi: apis.NewUpdateApi[User, UserUpdateParams]().
    WithPreUpdate(func(oldModel *User, newModel *User, params *UserUpdateParams, ctx fiber.Ctx, db orm.Db) error {
        // Only hash password if it's being updated
        if params.Password.Valid && params.Password.String != "" {
            hashed, err := bcrypt.GenerateFromPassword([]byte(params.Password.String), bcrypt.DefaultCost)
            if err != nil {
                return err
            }
            newModel.Password = string(hashed)
        } else {
            // Preserve existing password
            newModel.Password = oldModel.Password
        }
        return nil
    }),

// Business validation - Validate business rules before operation
CreateApi: apis.NewCreateApi[Order, OrderParams]().
    WithPreCreate(func(model *Order, params *OrderParams, ctx fiber.Ctx, db orm.Db) error {
        // Validate order total matches item totals
        if model.TotalAmount <= 0 {
            return result.Err("Order total must be greater than zero")
        }

        // Check inventory availability
        if !checkInventoryAvailable(model.Items) {
            return result.Err("Insufficient inventory for one or more items")
        }

        return nil
    }),
```

### Custom Handlers

#### Mixing Generated and Custom APIs

You can combine pre-built CRUD APIs with custom actions using `api.WithApis()`. This allows you to extend resources with domain-specific operations while maintaining the framework's conventions.

```go
package resources

import (
    "github.com/ilxqx/vef-framework-go/api"
    "github.com/ilxqx/vef-framework-go/apis"
)

type RoleResource struct {
    api.Resource
    apis.FindPageApi[models.Role, payloads.RoleSearch]
    apis.CreateApi[models.Role, payloads.RoleParams]
    apis.UpdateApi[models.Role, payloads.RoleParams]
    apis.DeleteApi[models.Role]
}

func NewRoleResource() api.Resource {
    return &RoleResource{
        Resource: api.NewResource(
            "app/sys/role",
            api.WithApis(
                api.Spec{
                    Action: "find_role_permissions",
                },
                api.Spec{
                    Action:      "save_role_permissions",
                    EnableAudit: true,  // Enable audit logging for this action
                },
            ),
        ),
        FindPageApi: apis.NewFindPageApi[models.Role, payloads.RoleSearch](),
        CreateApi:   apis.NewCreateApi[models.Role, payloads.RoleParams](),
        UpdateApi:   apis.NewUpdateApi[models.Role, payloads.RoleParams](),
        DeleteApi:   apis.NewDeleteApi[models.Role](),
    }
}

// Custom handler method for find_role_permissions action
func (r *RoleResource) FindRolePermissions(
    ctx fiber.Ctx,
    db orm.Db,
    params payloads.RolePermissionQuery,
) error {
    // Custom business logic
    // ...
    return result.Ok(permissions).Response(ctx)
}

// Custom handler method for save_role_permissions action
func (r *RoleResource) SaveRolePermissions(
    ctx fiber.Ctx,
    db orm.Db,
    params payloads.RolePermissionParams,
) error {
    // Transaction-based custom logic
    return db.RunInTx(ctx.Context(), func(txCtx context.Context, tx orm.Db) error {
        // Save permissions in transaction
        // ...
        return nil
    })
}
```

**Key Points:**

- **Method Naming**: Handler method names must be in PascalCase matching the snake_case action name (e.g., `find_role_permissions` → `FindRolePermissions`)
- **API Spec Configuration**: Each custom action can have its own configuration (permissions, audit, rate limiting)
- **Injection Rules**: Custom handler methods follow the same parameter injection rules as generated handlers
- **Mixed APIs**: You can freely mix generated CRUD APIs with custom actions in the same resource

#### Simple Custom Handlers

Add custom endpoints by defining methods on your resource:

```go
func (r *UserResource) ResetPassword(
    ctx fiber.Ctx,
    db orm.Db,
    logger log.Logger,
    principal *security.Principal,
    params ResetPasswordParams,
) error {
    logger.Infof("User %s resetting password", principal.Id)
    
    // Custom business logic
    var user models.User
    if err := db.NewSelect().
        Model(&user).
        Where(func(cb orm.ConditionBuilder) {
            cb.Equals("id", principal.Id)
        }).
        Scan(ctx.Context()); err != nil {
        return err
    }
    
    // Update password
    // ...
    
    return result.Ok().Response(ctx)
}
```

**Injectable Parameters:**

- `fiber.Ctx` - HTTP context
- `orm.Db` - Database connection
- `log.Logger` - Logger instance
- `mold.Transformer` - Data transformer
- `*security.Principal` - Current authenticated user
- `page.Pageable` - Pagination parameters
- Custom structs embedding `api.P`
- Custom structs embedding `api.M` (request metadata)
- Resource struct fields (direct fields, `api:"in"` tagged fields, or embedded structs)

**Example of Resource Field Injection:**

```go
type UserResource struct {
    api.Resource
    userService *UserService  // Resource field
}

func NewUserResource(userService *UserService) api.Resource {
    return &UserResource{
        Resource: api.NewResource("sys/user"),
        userService: userService,
    }
}

// Handler can inject userService directly
func (r *UserResource) SendNotification(
    ctx fiber.Ctx,
    service *UserService,  // Injected from r.userService
    params NotificationParams,
) error {
    return service.SendEmail(params.Email, params.Message)
}
```

**Why use parameter injection instead of `r.userService` directly?**

If your service implements the `log.LoggerConfigurable[T]` interface, the framework will automatically call the `WithLogger` method when injecting the service, providing a request-scoped logger. This allows each request to have its own logging context with request ID and other contextual information.

```go
type UserService struct {
    logger log.Logger
}

// Implement log.LoggerConfigurable[*UserService] interface
func (s *UserService) WithLogger(logger log.Logger) *UserService {
    return &UserService{logger: logger}
}

func (s *UserService) SendEmail(email, message string) error {
    s.logger.Infof("Sending email to %s", email)  // Request-scoped logger
    // ...
}
```

## Database Operations

### Query Builder

```go
var users []models.User
err := db.NewSelect().
    Model(&users).
    Where(func(cb orm.ConditionBuilder) {
        cb.Equals("is_active", true)
        cb.GreaterThan("age", 18)
        cb.Contains("username", keyword)
    }).
    Relation("Profile").
    OrderByDesc("created_at").
    Limit(10).
    Scan(ctx)
```

### Condition Builder Methods

Build type-safe query conditions:

- `Equals(column, value)` - Equal to
- `NotEquals(column, value)` - Not equal to
- `GreaterThan(column, value)` - Greater than
- `GreaterThanOrEquals(column, value)` - Greater than or equal
- `LessThan(column, value)` - Less than
- `LessThanOrEquals(column, value)` - Less than or equal
- `Contains(column, value)` - LIKE %value%
- `StartsWith(column, value)` - LIKE value%
- `EndsWith(column, value)` - LIKE %value
- `In(column, values)` - IN clause
- `Between(column, min, max)` - BETWEEN clause
- `IsNull(column)` - IS NULL
- `IsNotNull(column)` - IS NOT NULL
- `Or(conditions...)` - OR multiple conditions

### Search Tags

Automatically apply query conditions using `search` tags:

```go
type UserSearch struct {
    api.P
    Username string `search:"eq"`                                    // username = ?
    Email    string `search:"contains"`                              // email LIKE ?
    Age      int    `search:"gte"`                                   // age >= ?
    Status   string `search:"in"`                                    // status IN (?)
    Keyword  string `search:"contains,column=username|email|name"`   // Search multiple columns
}
```

**Supported Operators:**

**Comparison Operators:**

| Tag | SQL Operator | Description |
|-----|--------------|-------------|
| `eq` | = | Equal |
| `neq` | != | Not equal |
| `gt` | > | Greater than |
| `gte` | >= | Greater than or equal |
| `lt` | < | Less than |
| `lte` | <= | Less than or equal |

**Range Operators:**

| Tag | SQL Operator | Description |
|-----|--------------|-------------|
| `between` | BETWEEN | Between range |
| `notBetween` | NOT BETWEEN | Not between range |

**Collection Operators:**

| Tag | SQL Operator | Description |
|-----|--------------|-------------|
| `in` | IN | In list |
| `notIn` | NOT IN | Not in list |

**Null Check Operators:**

| Tag | SQL Operator | Description |
|-----|--------------|-------------|
| `isNull` | IS NULL | Is null |
| `isNotNull` | IS NOT NULL | Is not null |

**String Matching (Case Sensitive):**

| Tag | SQL Operator | Description |
|-----|--------------|-------------|
| `contains` | LIKE %?% | Contains |
| `notContains` | NOT LIKE %?% | Does not contain |
| `startsWith` | LIKE ?% | Starts with |
| `notStartsWith` | NOT LIKE ?% | Does not start with |
| `endsWith` | LIKE %? | Ends with |
| `notEndsWith` | NOT LIKE %? | Does not end with |

**String Matching (Case Insensitive):**

| Tag | SQL Operator | Description |
|-----|--------------|-------------|
| `iContains` | ILIKE %?% | Contains (case insensitive) |
| `iNotContains` | NOT ILIKE %?% | Does not contain (case insensitive) |
| `iStartsWith` | ILIKE ?% | Starts with (case insensitive) |
| `iNotStartsWith` | NOT ILIKE ?% | Does not start with (case insensitive) |
| `iEndsWith` | ILIKE %? | Ends with (case insensitive) |
| `iNotEndsWith` | NOT ILIKE %? | Does not end with (case insensitive) |

### Transactions

Execute multiple operations in a transaction:

```go
err := db.RunInTx(ctx.Context(), func(txCtx context.Context, tx orm.Db) error {
    // Insert user
    _, err := tx.NewInsert().Model(&user).Exec(txCtx)
    if err != nil {
        return err // Auto-rollback
    }

    // Update related records
    _, err = tx.NewUpdate().Model(&profile).WherePk().Exec(txCtx)
    return err // Auto-commit on nil, rollback on error
})
```

## Authentication & Authorization

### Authentication Methods

VEF supports multiple authentication strategies:

1. **Jwt Authentication** (default) - Bearer token or query parameter `?__accessToken=xxx`
2. **OpenApi Signature** - For external applications using HMAC signature
3. **Password Authentication** - Username/password login

### Implementing User Loader

Implement `security.UserLoader` to integrate with your user system:

```go
package services

import (
    "context"
    "github.com/ilxqx/vef-framework-go/orm"
    "github.com/ilxqx/vef-framework-go/security"
)

type MyUserLoader struct {
    db orm.Db
}

func (l *MyUserLoader) LoadByUsername(ctx context.Context, username string) (*security.Principal, string, error) {
    var user models.User
    if err := l.db.NewSelect().
        Model(&user).
        Where(func(cb orm.ConditionBuilder) {
            cb.Equals("username", username)
        }).
        Scan(ctx); err != nil {
        return nil, "", err
    }

    principal := &security.Principal{
        Type: security.PrincipalTypeUser,
        Id:   user.Id,
        Name: user.Name,
        Roles: []string{"user"}, // Load from database
    }

    return principal, user.Password, nil // Return hashed password
}

func (l *MyUserLoader) LoadById(ctx context.Context, id string) (*security.Principal, error) {
    // Similar implementation
}

func NewMyUserLoader(db orm.Db) *MyUserLoader {
    return &MyUserLoader{db: db}
}

// Register in main.go
func main() {
    vef.Run(
        vef.Provide(NewMyUserLoader),
    )
}
```

### Permission Control

Set permission tokens on Apis:

```go
CreateApi: apis.NewCreateApi[User, UserParams]().
    PermToken("sys.user.create"),
```

#### Using Built-in RBAC Implementation (Recommended)

The framework provides a built-in Role-Based Access Control (RBAC) implementation. You only need to implement the `security.RolePermissionsLoader` interface:

```go
package services

import (
    "context"
    "github.com/ilxqx/vef-framework-go/orm"
    "github.com/ilxqx/vef-framework-go/security"
)

type MyRolePermissionsLoader struct {
    db orm.Db
}

// LoadPermissions loads all permissions for the given role
// Returns map[permission token]data scope
func (l *MyRolePermissionsLoader) LoadPermissions(ctx context.Context, role string) (map[string]security.DataScope, error) {
    // Load role permissions from database
    var permissions []RolePermission
    if err := l.db.NewSelect().
        Model(&permissions).
        Where(func(cb orm.ConditionBuilder) {
            cb.Equals("role_code", role)
        }).
        Scan(ctx); err != nil {
        return nil, err
    }
    
    // Build mapping of permission tokens to data scopes
    result := make(map[string]security.DataScope)
    for _, perm := range permissions {
        // Create corresponding DataScope instance based on scope type
        var dataScope security.DataScope
        switch perm.DataScopeType {
        case "all":
            dataScope = security.NewAllDataScope()
        case "self":
            dataScope = security.NewSelfDataScope("")
        case "dept":
            dataScope = NewDepartmentDataScope() // Custom implementation
        // ... more custom data scopes
        }
        
        result[perm.PermissionToken] = dataScope
    }
    
    return result, nil
}

func NewMyRolePermissionsLoader(db orm.Db) security.RolePermissionsLoader {
    return &MyRolePermissionsLoader{db: db}
}

// Register in main.go
func main() {
    vef.Run(
        vef.Provide(NewMyRolePermissionsLoader),
    )
}
```

**Note:** The framework will automatically use your `RolePermissionsLoader` implementation to initialize the built-in RBAC permission checker and data permission resolver.

#### Fully Custom Permission Control

If you need to implement completely custom permission control logic (non-RBAC), you can implement the `security.PermissionChecker` interface and replace the framework's implementation:

```go
type MyCustomPermissionChecker struct {
    // Custom fields
}

func (c *MyCustomPermissionChecker) HasPermission(ctx context.Context, principal *security.Principal, permToken string) (bool, error) {
    // Custom permission check logic
    // ...
    return true, nil
}

func NewMyCustomPermissionChecker() security.PermissionChecker {
    return &MyCustomPermissionChecker{}
}

// Replace framework implementation in main.go
func main() {
    vef.Run(
        vef.Provide(NewMyCustomPermissionChecker),
        vef.Replace(vef.Annotate(
            NewMyCustomPermissionChecker,
            vef.As(new(security.PermissionChecker)),
        )),
    )
}
```

### Data Permissions

Data permissions implement row-level data access control, restricting users to specific data scopes.

#### Built-in Data Scopes

The framework provides two built-in data scope implementations:

1. **AllDataScope** - Unrestricted access to all data (typically for administrators)
2. **SelfDataScope** - Access only to self-created data

```go
import "github.com/ilxqx/vef-framework-go/security"

// All data
allScope := security.NewAllDataScope()

// Only self-created data (defaults to created_by column)
selfScope := security.NewSelfDataScope("")

// Custom creator column name
selfScope := security.NewSelfDataScope("creator_id")
```

#### Using Built-in RBAC Data Permissions (Recommended)

The framework's RBAC implementation automatically handles data permissions. Simply return the data scope for each permission token in `RolePermissionsLoader.LoadPermissions`:

```go
func (l *MyRolePermissionsLoader) LoadPermissions(ctx context.Context, role string) (map[string]security.DataScope, error) {
    result := make(map[string]security.DataScope)
    
    // Assign different data scopes to different permissions
    result["sys.user.view"] = security.NewAllDataScope()      // View all users
    result["sys.user.edit"] = security.NewSelfDataScope("")    // Edit only self-created users
    
    return result, nil
}
```

**Data Scope Priority:** When a user has multiple roles with different data scopes for the same permission token, the framework selects the scope with the highest priority. Built-in priority constants:

- `security.PrioritySelf` (10) - Self-created data only
- `security.PriorityDepartment` (20) - Department data
- `security.PriorityDepartmentAndSub` (30) - Department and sub-department data
- `security.PriorityOrganization` (40) - Organization data
- `security.PriorityOrganizationAndSub` (50) - Organization and sub-organization data
- `security.PriorityCustom` (60) - Custom data scope
- `security.PriorityAll` (10000) - All data

#### Custom Data Scopes

Implement the `security.DataScope` interface to create custom data access scopes:

```go
package scopes

import (
    "github.com/ilxqx/vef-framework-go/orm"
    "github.com/ilxqx/vef-framework-go/security"
)

type DepartmentDataScope struct{}

func NewDepartmentDataScope() security.DataScope {
    return &DepartmentDataScope{}
}

func (s *DepartmentDataScope) Key() string {
    return "department"
}

func (s *DepartmentDataScope) Priority() int {
    return security.PriorityDepartment // Use framework-defined priority
}

func (s *DepartmentDataScope) Supports(principal *security.Principal, table *orm.Table) bool {
    // Check if table has department_id column
    field, _ := table.Field("department_id")
    return field != nil
}

func (s *DepartmentDataScope) Apply(principal *security.Principal, query orm.SelectQuery) error {
    // Get user's department ID from principal.Details
    type UserDetails struct {
        DepartmentId string `json:"departmentId"`
    }
    
    details, ok := principal.Details.(UserDetails)
    if !ok {
        return nil // If no department info, don't apply filter
    }
    
    // Apply filtering condition
    query.Where(func(cb orm.ConditionBuilder) {
        cb.Equals("department_id", details.DepartmentId)
    })
    
    return nil
}
```

Then use the custom data scope in your `RolePermissionsLoader`:

```go
func (l *MyRolePermissionsLoader) LoadPermissions(ctx context.Context, role string) (map[string]security.DataScope, error) {
    result := make(map[string]security.DataScope)
    
    result["sys.user.view"] = NewDepartmentDataScope() // View only department users
    
    return result, nil
}
```

#### Fully Custom Data Permission Resolution

If you need to implement completely custom data permission resolution logic (non-RBAC), you can implement the `security.DataPermissionResolver` interface and replace the framework's implementation:

```go
type MyCustomDataPermResolver struct {
    // Custom fields
}

func (r *MyCustomDataPermResolver) ResolveDataScope(ctx context.Context, principal *security.Principal, permToken string) (security.DataScope, error) {
    // Custom data permission resolution logic
    // ...
    return security.NewAllDataScope(), nil
}

func NewMyCustomDataPermResolver() security.DataPermissionResolver {
    return &MyCustomDataPermResolver{}
}

// Replace framework implementation in main.go
func main() {
    vef.Run(
        vef.Provide(NewMyCustomDataPermResolver),
        vef.Replace(vef.Annotate(
            NewMyCustomDataPermResolver,
            vef.As(new(security.DataPermissionResolver)),
        )),
    )
}
```

## Configuration

### Configuration File

Place `application.toml` in `./configs/` or `./` directory, or specify via `VEF_CONFIG_PATH` environment variable.

**Complete Configuration Example:**

```toml
[vef.app]
name = "my-app"          # Application name
port = 8080              # HTTP port
body_limit = "10MB"      # Request body size limit

[vef.datasource]
type = "postgres"        # Database type: postgres, mysql, sqlite
host = "localhost"
port = 5432
user = "postgres"
password = "password"
database = "mydb"
schema = "public"        # PostgreSQL schema
# path = "./data.db"    # SQLite database file path

[vef.security]
token_expires = "2h"     # Jwt token expiration time

[vef.storage]
provider = "minio"       # Storage provider: memory, filesystem, minio (default: memory)

[vef.storage.minio]
endpoint = "localhost:9000"
access_key = "minioadmin"
secret_key = "minioadmin"
use_ssl = false
region = "us-east-1"
bucket = "mybucket"

[vef.storage.filesystem]
root = "./storage"       # Used when provider = "filesystem"

[vef.redis]
host = "localhost"
port = 6379
user = ""                # Optional
password = ""            # Optional
database = 0             # 0-15
network = "tcp"          # tcp or unix

[vef.cors]
enabled = true
allow_origins = ["*"]
```

### Environment Variables

Override configuration with environment variables:

- `VEF_CONFIG_PATH` - Configuration file path
- `VEF_LOG_LEVEL` - Log level (debug, info, warn, error)
- `VEF_NODE_ID` - XID node identifier for ID generation
- `VEF_I18N_LANGUAGE` - Language (en, zh-CN)

## Advanced Features

### Cache

Use in-memory or Redis cache:

```go
import (
    "github.com/ilxqx/vef-framework-go/cache"
    "time"
)

// In-memory cache
memCache := cache.NewMemory[models.User](
    cache.WithMemMaxSize(1000),
    cache.WithMemDefaultTtl(5 * time.Minute),
)

// Redis cache
redisCache := cache.NewRedis[models.User](
    redisClient,
    "users",
    cache.WithRdsDefaultTtl(10 * time.Minute),
)

// Usage
user, err := memCache.GetOrLoad(ctx, "user:123", func(ctx context.Context) (models.User, error) {
    // Fallback loader when cache miss
    return loadUserFromDB(ctx, "123")
})
```

### Event Bus

Publish and subscribe to events:

```go
import "github.com/ilxqx/vef-framework-go/event"

// Publishing events
func (r *UserResource) CreateUser(ctx fiber.Ctx, bus event.Bus, ...) error {
    // Create user logic
    
    bus.Publish(event.NewBaseEvent(
        "user.created",
        event.WithSource("user-service"),
        event.WithMeta("userId", user.Id),
    ))
    
    return result.Ok().Response(ctx)
}

// Subscribing to events
func main() {
    vef.Run(
        vef.Invoke(func(bus event.Bus, logger log.Logger) {
            unsubscribe := bus.Subscribe("user.created", func(ctx context.Context, e event.Event) {
                // Handle event
                logger.Infof("User created: %s", e.Meta()["userId"])
            })
            
            // Optionally unsubscribe later
            _ = unsubscribe
        }),
    )
}
```

### Lifecycle Hooks

The framework provides lifecycle management through `vef.Lifecycle`, allowing you to register hooks that execute during application startup and shutdown. This is essential for proper resource cleanup, particularly for event subscribers.

#### Event Subscriber Cleanup

When registering event subscribers, you should clean up subscriptions on shutdown to prevent resource leaks:

```go
import (
    "github.com/ilxqx/vef-framework-go"
    "github.com/ilxqx/vef-framework-go/event"
    "github.com/ilxqx/vef-framework-go/orm"
)

var Module = vef.Module(
    "app:vef",
    vef.Invoke(
        func(lc vef.Lifecycle, db orm.Db, subscriber event.Subscriber) {
            // Create and register audit event subscriber
            auditSub := NewAuditEventSubscriber(db, subscriber)

            // Register cleanup hook
            lc.Append(vef.StopHook(func() {
                auditSub.Unsubscribe()  // Cleanup on shutdown
            }))

            // Create and register login event subscriber
            loginSub := NewLoginEventSubscriber(db, subscriber)

            // Register cleanup hook
            lc.Append(vef.StopHook(func() {
                loginSub.Unsubscribe()  // Cleanup on shutdown
            }))
        },
    ),
)
```

**Key Patterns:**

1. **Store unsubscribe function**: Event subscriber constructors should return an `UnsubscribeFunc` when they call `bus.Subscribe()`
2. **Register stop hooks**: Use `lc.Append(vef.StopHook(...))` to register cleanup functions
3. **Call unsubscribe in hooks**: Invoke the stored `Unsubscribe()` function during shutdown

**Example Event Subscriber Implementation:**

```go
type AuditEventSubscriber struct {
    db           orm.Db
    unsubscribe  event.UnsubscribeFunc
}

func NewAuditEventSubscriber(db orm.Db, subscriber event.Subscriber) *AuditEventSubscriber {
    sub := &AuditEventSubscriber{db: db}

    // Subscribe and store unsubscribe function
    sub.unsubscribe = subscriber.Subscribe("*.created", sub.handleAuditEvent)

    return sub
}

func (s *AuditEventSubscriber) handleAuditEvent(ctx context.Context, e event.Event) {
    // Handle audit logging
}

func (s *AuditEventSubscriber) Unsubscribe() {
    if s.unsubscribe != nil {
        s.unsubscribe()
    }
}
```

This pattern ensures graceful shutdown without resource leaks or orphaned subscriptions.

### Context Helpers

The `contextx` package provides utility functions to access request-scoped resources when dependency injection is not available. These helpers are useful in custom handlers, hooks, or other scenarios where you need to access framework-provided resources from the Fiber context.

```go
import "github.com/ilxqx/vef-framework-go/contextx"

func (r *RoleResource) CustomMethod(ctx fiber.Ctx) error {
    // Get request-scoped database (with operator pre-configured)
    db := contextx.Db(ctx)

    // Get current authenticated user
    principal := contextx.Principal(ctx)

    // Get request-scoped logger (includes request ID)
    logger := contextx.Logger(ctx)

    // Get mold transformer for data sanitization
    transformer := contextx.Transformer(ctx)

    // Use the resources
    logger.Infof("User %s performing custom operation", principal.Id)

    var model models.SomeModel
    if err := db.NewSelect().Model(&model).Scan(ctx.Context()); err != nil {
        return err
    }

    return result.Ok(model).Response(ctx)
}
```

**Available Helpers:**

- **`contextx.Db(ctx)`** - Returns request-scoped `orm.Db` with audit fields (like `operator`) pre-configured
- **`contextx.Principal(ctx)`** - Returns current `*security.Principal` (authenticated user or anonymous)
- **`contextx.Logger(ctx)`** - Returns request-scoped `log.Logger` with request ID for correlation
- **`contextx.Transformer(ctx)`** - Returns `mold.Transformer` for data transformation and sanitization

**When to Use:**

- **Use contextx helpers**: In custom handlers where you cannot use parameter injection, or in utility functions that only receive `fiber.Ctx`
- **Prefer parameter injection**: When defining API handler methods, let the framework inject dependencies directly as parameters for better testability and clarity

**Example - Using Both Patterns:**

```go
// Prefer this: Parameter injection in handler
func (r *UserResource) UpdateProfile(
    ctx fiber.Ctx,
    db orm.Db,           // Injected by framework
    logger log.Logger,   // Injected by framework
    params ProfileParams,
) error {
    logger.Infof("Updating profile")
    // ...
}

// Use contextx when injection not available
func helperFunction(ctx fiber.Ctx) error {
    db := contextx.Db(ctx)       // Extract from context
    logger := contextx.Logger(ctx)
    logger.Infof("Helper function")
    // ...
}
```

### Cron Scheduler

The framework provides cron job scheduling based on [gocron](https://github.com/go-co-op/gocron).

#### Basic Usage

Inject `cron.Scheduler` via DI and create jobs:

```go
import (
    "context"
    "time"
    "github.com/ilxqx/vef-framework-go/cron"
)

func main() {
    vef.Run(
        vef.Invoke(func(scheduler cron.Scheduler) {
            // Cron expression job (5-field format)
            scheduler.NewJob(
                cron.NewCronJob(
                    "0 0 * * *",  // Expression: daily at midnight
                    false,         // withSeconds: use 5-field format
                    cron.WithName("daily-cleanup"),
                    cron.WithTags("maintenance"),
                    cron.WithTask(func(ctx context.Context) {
                        // Task logic
                    }),
                ),
            )
            
            // Fixed interval job
            scheduler.NewJob(
                cron.NewDurationJob(
                    5*time.Minute,
                    cron.WithName("health-check"),
                    cron.WithTask(func() {
                        // Every 5 minutes
                    }),
                ),
            )
        }),
    )
}
```

#### Job Types

The framework supports multiple job scheduling strategies:

**1. Cron Expression Jobs**

```go
// 5-field format: minute hour day month weekday
scheduler.NewJob(
    cron.NewCronJob(
        "30 * * * *",  // Every hour at minute 30
        false,          // No seconds field
        cron.WithName("hourly-report"),
        cron.WithTask(func() {
            // Generate report
        }),
    ),
)

// 6-field format: second minute hour day month weekday
scheduler.NewJob(
    cron.NewCronJob(
        "0 30 * * * *",  // Every hour at minute 30, second 0
        true,             // With seconds field
        cron.WithName("precise-task"),
        cron.WithTask(func() {
            // Precise timing task
        }),
    ),
)
```

**2. Fixed Interval Jobs**

```go
scheduler.NewJob(
    cron.NewDurationJob(
        10*time.Second,
        cron.WithName("metrics-collector"),
        cron.WithTask(func() {
            // Collect metrics every 10 seconds
        }),
    ),
)
```

**3. Random Interval Jobs**

```go
scheduler.NewJob(
    cron.NewDurationRandomJob(
        1*time.Minute,  // Minimum interval
        5*time.Minute,  // Maximum interval
        cron.WithName("random-check"),
        cron.WithTask(func() {
            // Execute at random intervals between 1-5 minutes
        }),
    ),
)
```

**4. One-Time Jobs**

```go
// Execute immediately
scheduler.NewJob(
    cron.NewOneTimeJob(
        []time.Time{},  // Empty slice means immediate execution
        cron.WithName("init-task"),
        cron.WithTask(func() {
            // Initialization task
        }),
    ),
)

// Execute at specific time
scheduler.NewJob(
    cron.NewOneTimeJob(
        []time.Time{time.Now().Add(1 * time.Hour)},
        cron.WithName("delayed-task"),
        cron.WithTask(func() {
            // Execute after 1 hour
        }),
    ),
)

// Execute at multiple specific times
scheduler.NewJob(
    cron.NewOneTimeJob(
        []time.Time{
            time.Date(2024, 12, 31, 23, 59, 0, 0, time.Local),
            time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local),
        },
        cron.WithName("new-year-task"),
        cron.WithTask(func() {
            // Execute at specific times
        }),
    ),
)
```

#### Job Configuration Options

```go
scheduler.NewJob(
    cron.NewDurationJob(
        1*time.Hour,
        // Job name (required)
        cron.WithName("backup-task"),
        
        // Tags (for grouping and bulk operations)
        cron.WithTags("backup", "critical"),
        
        // Task handler function (required)
        cron.WithTask(func(ctx context.Context) {
            // If the function accepts context.Context, the framework auto-injects it
            // Supports graceful shutdown and timeout control
        }),
        
        // Allow concurrent execution (default is singleton mode)
        cron.WithConcurrent(),
        
        // Set start time
        cron.WithStartAt(time.Now().Add(10 * time.Minute)),
        
        // Start immediately
        cron.WithStartImmediately(),
        
        // Set stop time
        cron.WithStopAt(time.Now().Add(24 * time.Hour)),
        
        // Limit number of runs
        cron.WithLimitedRuns(100),
        
        // Custom context
        cron.WithContext(context.Background()),
    ),
)
```

#### Job Management

```go
vef.Invoke(func(scheduler cron.Scheduler) {
    // Create job
    job, _ := scheduler.NewJob(
        cron.NewDurationJob(
            1*time.Minute,
            cron.WithName("my-task"),
            cron.WithTags("tag1", "tag2"),
            cron.WithTask(func() {}),
        ),
    )
    
    // Get all jobs
    allJobs := scheduler.Jobs()
    
    // Remove jobs by tags
    scheduler.RemoveByTags("tag1", "tag2")
    
    // Remove job by ID
    scheduler.RemoveJob(job.Id())
    
    // Update job definition
    scheduler.Update(job.Id(), cron.NewDurationJob(
        2*time.Minute,
        cron.WithName("my-task-updated"),
        cron.WithTask(func() {}),
    ))
    
    // Run job immediately (doesn't affect schedule)
    job.RunNow()
    
    // Get next run time
    nextRun, _ := job.NextRun()
    
    // Get last run time
    lastRun, _ := job.LastRun()
    
    // Stop all jobs
    scheduler.StopJobs()
})
```

### File Storage

The framework provides built-in file storage functionality with support for MinIO, filesystem, and in-memory storage.

#### Built-in Storage Resource

The framework automatically registers the `sys/storage` resource with the following Api endpoints:

| Action | Description |
|--------|-------------|
| `upload` | Upload file (auto-generates unique filename) |
| `get_presigned_url` | Get presigned URL (for direct access or upload) |
| `delete_temp` | Delete temporary file (only keys under `temp/`) |
| `stat` | Get file metadata |
| `list` | List files |

**Upload Example:**

```bash
# Using built-in upload Api
curl -X POST http://localhost:8080/api \
  -H "Authorization: Bearer <token>" \
  -F "resource=sys/storage" \
  -F "action=upload" \
  -F "version=v1" \
  -F "params[file]=@/path/to/file.jpg" \
  -F "params[contentType]=image/jpeg" \
  -F "params[metadata][key1]=value1"
```

**Upload Response:**

```json
{
  "code": 0,
  "message": "Success",
  "data": {
    "key": "temp/2025/01/15/550e8400-e29b-41d4-a716-446655440000.jpg",
    "size": 1024000,
    "contentType": "image/jpeg",
    "etag": "\"d41d8cd98f00b204e9800998ecf8427e\"",
    "lastModified": "2025-01-15T10:30:00Z",
    "metadata": {
      "Original-Filename": "file.jpg",
      "key1": "value1"
    }
  }
}
```

#### File Key Conventions

The framework uses the following naming convention for uploaded files:

- **Temporary files**: `temp/YYYY/MM/DD/{uuid}{extension}`
  - Example: `temp/2025/01/15/550e8400-e29b-41d4-a716-446655440000.jpg`
  - Original filename is preserved in `Original-Filename` metadata

- **Permanent files**: Promote temporary files via `PromoteObject`
  - Removes `temp/` prefix from the path
  - Example: `temp/2025/01/15/xxx.jpg` → `2025/01/15/xxx.jpg`

#### Custom File Upload

Inject `storage.Service` in custom resources for file uploads:

```go
import (
    "mime/multipart"

    "github.com/gofiber/fiber/v3"
    "github.com/ilxqx/vef-framework-go/api"
    "github.com/ilxqx/vef-framework-go/result"
    "github.com/ilxqx/vef-framework-go/storage"
)

// Define upload parameter struct
type UploadAvatarParams struct {
    api.P

    File *multipart.FileHeader `json:"file"`
}

func (r *UserResource) UploadAvatar(
    ctx fiber.Ctx,
    service storage.Service,
    params UploadAvatarParams,
) error {
    // Check if file exists
    if params.File == nil {
        return result.Err("File is required")
    }

    // Open uploaded file
    reader, err := params.File.Open()
    if err != nil {
        return err
    }
    defer reader.Close()

    // Custom file path
    info, err := service.PutObject(ctx.Context(), storage.PutObjectOptions{
        Key:         "avatars/" + params.File.Filename,
        Reader:      reader,
        Size:        params.File.Size,
        ContentType: params.File.Header.Get("Content-Type"),
        Metadata: map[string]string{
            "userId": "12345",
        },
    })
    if err != nil {
        return err
    }
    
    return result.Ok(info).Response(ctx)
}
```

#### Promoting Temporary Files

Use `PromoteObject` to convert temporary uploads to permanent files:

```go
// After business logic confirms, promote temporary file
info, err := provider.PromoteObject(ctx.Context(), "temp/2025/01/15/xxx.jpg")
// info.Key becomes: "2025/01/15/xxx.jpg"
```

#### Storage Configuration

Set `vef.storage.provider` to `minio`, `filesystem`, or `memory` (default) and configure the matching section in `application.toml`:

```toml
[vef.storage]
provider = "minio"  # options: minio, filesystem, memory

[vef.storage.minio]
endpoint = "localhost:9000"
access_key = "minioadmin"
secret_key = "minioadmin"
use_ssl = false
region = "us-east-1"
bucket = "mybucket"

[vef.storage.filesystem]
root = "./storage"       # Base directory when provider = "filesystem"
```

### Data Validation

Use [go-playground/validator](https://github.com/go-playground/validator) tags:

```go
type UserParams struct {
    Username string `validate:"required,alphanum,min=3,max=32" label:"Username"`
    Email    string `validate:"required,email" label:"Email"`
    Age      int    `validate:"min=18,max=120" label:"Age"`
    Website  string `validate:"omitempty,url" label:"Website"`
    Password string `validate:"required,min=8,containsany=!@#$%^&*" label:"Password"`
}
```

**Common Rules:**

| Rule | Description |
|------|-------------|
| `required` | Required field |
| `omitempty` | Optional field (skip validation if empty) |
| `min` | Minimum value (number) or minimum length (string) |
| `max` | Maximum value (number) or maximum length (string) |
| `len` | Exact length |
| `eq` | Equal to |
| `ne` | Not equal to |
| `gt` | Greater than |
| `gte` | Greater than or equal to |
| `lt` | Less than |
| `lte` | Less than or equal to |
| `alpha` | Alphabetic characters only |
| `alphanum` | Alphanumeric characters |
| `ascii` | ASCII characters |
| `numeric` | Numeric string |
| `email` | Email address |
| `url` | URL |
| `uuid` | UUID format |
| `ip` | IP address |
| `json` | JSON format |
| `contains` | Contains substring |
| `startswith` | Starts with string |
| `endswith` | Ends with string |

### CLI Tools

VEF Framework provides the `vef-cli` command-line tool for code generation and project scaffolding tasks.

#### Generate Build Info

The `generate-build-info` command creates a build_info.go file with app version, commit hash, and build timestamp:

```bash
go run github.com/ilxqx/vef-framework-go/cmd/vef-cli@latest generate-build-info -o internal/vef/build_info.go -p vef
```

**Options:**
- `-o, --output` - Output file path (default: `build_info.go`)
- `-p, --package` - Package name (default: current directory name)

**Usage in go:generate:**

```go
//go:generate go run github.com/ilxqx/vef-framework-go/cmd/vef-cli@latest generate-build-info -o internal/vef/build_info.go -p vef
```

The generated file provides a `BuildInfo` variable compatible with the monitor module:

```go
package vef

import "github.com/ilxqx/vef-framework-go/monitor"

// BuildInfo is a pointer to build metadata used by the monitor module.
var BuildInfo = &monitor.BuildInfo{
    AppVersion: "v1.0.0",               // From git tags (or "dev")
    BuildTime:  "2025-01-15T10:30:00Z", // Build timestamp
    GitCommit:  "abc123...",            // Git commit SHA
}
```

**Generated Fields:**
- **Version**: Extracted from git tags (e.g., `v1.0.0`). Falls back to `"dev"` if no tags exist.
- **Commit**: Full git commit SHA from current HEAD.
- **BuildTime**: UTC timestamp when the file was generated.

#### Generate Model Schema

The `generate-model-schema` command generates type-safe field accessor functions for your models:

```bash
go run github.com/ilxqx/vef-framework-go/cmd/vef-cli@latest generate-model-schema -i ./models -o ./schemas -p schemas
```

**Options:**
- `-i, --input` - Input directory containing model files (required)
- `-o, --output` - Output directory for generated schema files (required)
- `-p, --package` - Package name for generated files (required)

**Usage in go:generate:**

```go
//go:generate go run github.com/ilxqx/vef-framework-go/cmd/vef-cli@latest generate-model-schema -i ./models -o ./schemas -p schemas
```

The generated schema provides type-safe field accessors:

```go
package schemas

var User = struct {
    Id        func(withTablePrefix ...bool) string
    Username  func(withTablePrefix ...bool) string
    Email     func(withTablePrefix ...bool) string
    CreatedAt func(withTablePrefix ...bool) string
    // ... other fields
}{
    Id:        field("id", "su"),
    Username:  field("username", "su"),
    Email:     field("email", "su"),
    CreatedAt: field("created_at", "su"),
}
```

**Usage in queries:**

```go
import "my-app/internal/sys/schemas"

// Type-safe column references
db.NewSelect().
    Model(&users).
    Where(func(cb orm.ConditionBuilder) {
        cb.Equals(schemas.User.Username(), "admin")
        cb.IsNotNull(schemas.User.Email())
    }).
    OrderBy(schemas.User.CreatedAt(true) + " DESC"). // With table prefix
    Scan(ctx)
```

**Benefits:**
- **Type safety**: Catch typos at compile time
- **IDE autocomplete**: Field names are discoverable
- **Refactoring support**: Renaming fields updates all references
- **Table prefix handling**: Optionally include table alias in column names

For AI-assisted development guidelines, see `cmd/CMD_DEV_GUIDELINES.md`.

## Best Practices

### Project Structure

```txt
my-app/
├── cmd/
│   └── main.go                 # Application entry point
├── configs/
│   └── application.toml        # Configuration file
├── internal/
│   ├── models/                 # Data models
│   │   ├── user.go
│   │   └── order.go
│   ├── payloads/               # Api parameters
│   │   ├── user.go
│   │   └── order.go
│   ├── resources/              # Api resources
│   │   ├── user.go
│   │   └── order.go
│   └── services/               # Business services
│       ├── user_service.go
│       └── email_service.go
└── go.mod
```

### Naming Conventions

- **Models:** Singular PascalCase (e.g., `User`, `Order`)
- **Resources:** Lowercase with slashes (e.g., `sys/user`, `shop/order`, `auth/user_role`)
- **Parameters:** `XxxParams` (Create/Update), `XxxSearch` (Query)
- **Actions:** Lowercase snake_case (e.g., `find_page`, `create_user`)

### Error Handling

Use framework's Result type for consistent error responses:

```go
import "github.com/ilxqx/vef-framework-go/result"

// Success
return result.Ok(data).Response(ctx)

// Error
return result.Err("Operation failed")
return result.ErrWithCode(result.ErrCodeBadRequest, "Invalid parameters")
return result.Errf("User %s not found", username)
```

### Logging

Inject logger and use:

```go
func (r *UserResource) Handler(
    ctx fiber.Ctx,
    logger log.Logger,
) error {
    logger.Infof("Processing request from %s", ctx.IP())
    logger.Warnf("Unusual activity detected")
    logger.Errorf("Operation failed: %v", err)
    
    return nil
}
```

## Documentation & Resources

- [Fiber Web Framework](https://gofiber.io/) - Underlying HTTP framework
- [Bun ORM](https://bun.uptrace.dev/) - Database ORM
- [Go Playground Validator](https://github.com/go-playground/validator) - Data validation
- [Uber FX](https://uber-go.github.io/fx/) - Dependency injection

## License

This project is licensed under the [Apache License 2.0](LICENSE).
