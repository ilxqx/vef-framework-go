# VEF Framework Go

📖 [English](./README.md) | [简体中文](./README.zh-CN.md)

A modern Go web development framework built on Uber FX dependency injection and Fiber, designed for rapid enterprise application development with opinionated conventions and comprehensive built-in features.

**Current Version:** v0.6.0

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

**Requirements:** Go 1.25 or higher

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
    "page": 1,
    "size": 20,
    "keyword": "john"
  },
  "meta": {}
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
    orm.Model     `bun:"extend"`
    
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

**Null Types:** Use `null.String`, `null.Int`, `null.Bool`, etc. for nullable fields.

## Building CRUD Apis

### Step 1: Define Parameter Structures

**Search Parameters:**

```go
package payloads

import "github.com/ilxqx/vef-framework-go/api"

type UserSearch struct {
    api.In
    Keyword string `json:"keyword" search:"contains,column=username|email"`
    IsActive *bool `json:"isActive" search:"eq"`
}
```

**Create/Update Parameters:**

```go
type UserParams struct {
    api.In
    Id       string      `json:"id"` // Required for updates

    Username string      `json:"username" validate:"required,alphanum,max=32" label:"Username"`
    Email    null.String `json:"email" validate:"omitempty,email,max=64" label:"Email"`
    IsActive bool        `json:"isActive"`
}
```

### Step 2: Create Api Resource

```go
package resources

import (
    "github.com/ilxqx/vef-framework-go/api"
    "github.com/ilxqx/vef-framework-go/apis"
)

type UserResource struct {
    api.Resource
    *apis.FindAllApi[models.User, payloads.UserSearch]
    *apis.FindPageApi[models.User, payloads.UserSearch]
    *apis.CreateApi[models.User, payloads.UserParams]
    *apis.UpdateApi[models.User, payloads.UserParams]
    *apis.DeleteApi[models.User]
}

func NewUserResource() api.Resource {
    return &UserResource{
        Resource: api.NewResource("sys/user"),
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
| `WithSelect` | Add column to SELECT clause | QueryAll | All FindApi |
| `WithSelectAs` | Add column with alias to SELECT clause | QueryAll | All FindApi |
| `WithDefaultSort` | Set default sorting specifications | QueryRoot | All FindApi |
| `WithCondition` | Add WHERE condition using ConditionBuilder | QueryRoot | All FindApi |
| `WithRelation` | Add relation join | QueryAll | All FindApi |
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

// Example: Adding computed fields
FindPageApi: apis.NewFindPageApi[Order, OrderSearch]().
    WithProcessor(func(page page.Page[Order], search OrderSearch, ctx fiber.Ctx) any {
        for i := range page.Items {
            // Calculate total amount
            page.Items[i].TotalAmount = page.Items[i].Quantity * page.Items[i].UnitPrice
        }
        return page
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
        Name: "Profile",
    }),
```

**WithAuditUserNames Example:**

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithAuditUserNames(&User{}), // Uses "name" column by default
    
// Or specify custom column name
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithAuditUserNames(&User{}, "username"),
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

- `WithSelect`, `WithSelectAs`, `WithRelation`: Default to `QueryAll` (applies to all parts)
- `WithCondition`, `WithQueryApplier`, `WithDefaultSort`: Default to `QueryRoot` (applies to root query only)

**Normal Query Example:**

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithSelect("username").              // Applies to QueryAll (main query)
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
        Name: "Metadata",
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
FindTreeOptionsApi: apis.NewFindTreeOptionsApi[Category, CategorySearch](
    buildCategoryTree,
).
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
```

**FindTreeApi:**

```go
FindTreeApi: apis.NewFindTreeApi[Category, CategorySearch](buildTree).
    WithIdColumn("id").              // ID column name (default: "id")
    WithParentIdColumn("parent_id"), // Parent ID column name (default: "parent_id")
```

**FindTreeOptionsApi:**

Combines both options and tree configuration:

```go
FindTreeOptionsApi: apis.NewFindTreeOptionsApi[Category, CategorySearch](buildTree).
    WithDefaultColumnMapping(&apis.DataOptionColumnMapping{
        LabelColumn: "name",
        ValueColumn: "id",
    }).
    WithIdColumn("id").
    WithParentIdColumn("parent_id"),
```

**ExportApi:**

```go
ExportApi: apis.NewExportApi[User, UserSearch]().
    WithDefaultFormat("xlsx").                    // Default export format: "xlsx" or "csv"
    WithExcelOptions(&excel.ExportOptions{        // Excel-specific options
        SheetName: "Users",
    }).
    WithCsvOptions(&csv.ExportOptions{            // CSV-specific options
        Delimiter: ',',
    }).
    WithPreExport(func(users []User, search UserSearch, ctx fiber.Ctx) ([]User, error) {
        // Modify data before export (e.g., data masking)
        for i := range users {
            users[i].Password = "***"
        }
        return users, nil
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
    PreCreate(func(model *User, params *UserParams, ctx fiber.Ctx, db orm.Db) error {
        // Hash password before creating user
        hashed, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
        if err != nil {
            return err
        }
        model.Password = string(hashed)
        return nil
    }).
    PostCreate(func(model *User, params *UserParams, ctx fiber.Ctx, tx orm.Db) error {
        // Send welcome email after user creation (within transaction)
        return sendWelcomeEmail(model.Email)
    }),
```

Available hooks:

**Single Record Operations:**

- `PreCreate`, `PostCreate` - Before/after creation (PostCreate runs in transaction)
- `PreUpdate`, `PostUpdate` - Before/after update (receives both old and new model, PostUpdate runs in transaction)
- `PreDelete`, `PostDelete` - Before/after deletion (PostDelete runs in transaction)

**Batch Operations:**

- `PreCreateMany`, `PostCreateMany` - Before/after batch creation (PostCreateMany runs in transaction)
- `PreUpdateMany`, `PostUpdateMany` - Before/after batch update (receives old and new model arrays, PostUpdateMany runs in transaction)
- `PreDeleteMany`, `PostDeleteMany` - Before/after batch deletion (PostDeleteMany runs in transaction)

**Import/Export Operations:**

- `PreImport`, `PostImport` - Before/after import (PreImport for validation, PostImport runs in transaction)
- `PreExport` - Before export (for data formatting)

### Custom Handlers

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
    api.In
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
        return nil, constants.Empty, err
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
- `security.PriorityDeptAndSub` (30) - Department and sub-department data
- `security.PriorityOrganization` (40) - Organization data
- `security.PriorityOrgAndSub` (50) - Organization and sub-organization data
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
provider = "minio"       # Storage provider: memory, minio

[vef.storage.minio]
endpoint = "localhost:9000"
access_key = "minioadmin"
secret_key = "minioadmin"
use_ssl = false
region = "us-east-1"
bucket = "mybucket"

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
- `VEF_NODE_ID` - Snowflake node ID for ID generation
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
    cache.WithMemDefaultTTL(5 * time.Minute),
)

// Redis cache
redisCache := cache.NewRedis[models.User](
    redisClient,
    "users",
    cache.WithRdsDefaultTTL(10 * time.Minute),
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
    
    bus.Publish(event.NewBase("user.created", "user-service", map[string]string{
        "userId": user.Id,
    }))
    
    return result.Ok().Response(ctx)
}

// Subscribing to events
func main() {
    vef.Run(
        vef.Invoke(func(bus event.Bus) {
            unsubscribe := bus.Subscribe("user.created", func(ctx context.Context, e event.Event) {
                // Handle event
                log.Infof("User created: %s", e.Meta()["userId"])
            })
            
            // Optionally unsubscribe later
            _ = unsubscribe
        }),
    )
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

The framework provides built-in file storage functionality with support for MinIO and in-memory storage.

#### Built-in Storage Resource

The framework automatically registers the `base/storage` resource with the following Api endpoints:

| Action | Description |
|--------|-------------|
| `upload` | Upload file (auto-generates unique filename) |
| `getPresignedUrl` | Get presigned URL (for direct access or upload) |
| `stat` | Get file metadata |
| `list` | List files |

**Upload Example:**

```bash
# Using built-in upload Api
curl -X POST http://localhost:8080/api \
  -H "Authorization: Bearer <token>" \
  -F "resource=base/storage" \
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

Inject `storage.Provider` in custom resources for file uploads:

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
    api.In
    
    File *multipart.FileHeader `json:"file"`
}

func (r *UserResource) UploadAvatar(
    ctx fiber.Ctx,
    provider storage.Provider,
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
    info, err := provider.PutObject(ctx.Context(), storage.PutObjectOptions{
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

Configure storage in `application.toml`:

```toml
[vef.storage]
provider = "minio"  # or "memory" (for testing)

[vef.storage.minio]
endpoint = "localhost:9000"
access_key = "minioadmin"
secret_key = "minioadmin"
use_ssl = false
region = "us-east-1"
bucket = "mybucket"
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
