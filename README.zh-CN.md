# VEF Framework Go

📖 [English](./README.md) | [简体中文](./README.zh-CN.md)

[![GitHub Release](https://img.shields.io/github/v/release/ilxqx/vef-framework-go)](https://github.com/ilxqx/vef-framework-go/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/ilxqx/vef-framework-go/test.yml?branch=main)](https://github.com/ilxqx/vef-framework-go/actions/workflows/test.yml)
[![Coverage](https://img.shields.io/codecov/c/github/ilxqx/vef-framework-go)](https://codecov.io/gh/ilxqx/vef-framework-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/ilxqx/vef-framework-go.svg)](https://pkg.go.dev/github.com/ilxqx/vef-framework-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/ilxqx/vef-framework-go)](https://goreportcard.com/report/github.com/ilxqx/vef-framework-go)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/ilxqx/vef-framework-go)
[![License](https://img.shields.io/github/license/ilxqx/vef-framework-go)](https://github.com/ilxqx/vef-framework-go/blob/main/LICENSE)

一个基于 Uber FX 依赖注入和 Fiber 构建的现代化 Go Web 开发框架，采用约定优于配置的设计理念，为企业级应用快速开发提供开箱即用的完整功能。

## ⚠️ 开发状态与稳定性说明

> **重要提示**：VEF Framework Go 正处于积极开发阶段，尚未发布稳定的 1.0 版本。虽然框架目前在功能上已基本稳定，但在不断完善最佳实践和改进规范的过程中可能会出现破坏性更新。我们会尽力减少此类更新，但为了架构优化和更好的开发体验，有时不可避免需要引入不向后兼容的变更。因此，在生产环境中使用本框架时请务必谨慎，并做好应对重大版本升级时的迁移工作准备。

## 核心特性

- **单一端点 Api 架构** - 所有 Api 请求通过 `POST /api` 统一处理，请求响应格式一致
- **泛型 CRUD Api** - 预置类型安全的增删改查操作，极少样板代码
- **类型安全的 ORM** - 基于 Bun 的流式查询构建器，自动审计字段维护
- **多策略认证** - 内置 Jwt、OpenApi 签名、密码认证，开箱即用
- **模块化设计** - Uber FX 依赖注入，可插拔模块化架构
- **内置功能齐全** - 缓存、事件总线、定时任务、对象存储、数据验证、国际化
- **RBAC 与数据权限** - 行级安全控制，可自定义数据范围

## 快速开始

### 安装

```bash
go get github.com/ilxqx/vef-framework-go
```

**环境要求：** Go 1.25.0 或更高版本

**问题排查：** 如果在执行 `go mod tidy` 时遇到 `google.golang.org/genproto` 的模糊依赖错误，请运行：

```bash
go get google.golang.org/genproto@latest
go mod tidy
```

### 最小示例

创建 `main.go`：

```go
package main

import "github.com/ilxqx/vef-framework-go"

func main() {
    vef.Run()
}
```

创建 `configs/application.toml`：

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

运行应用：

```bash
go run main.go
```

您的 Api 服务现已运行在 `http://localhost:8080`。

## 项目结构

### 推荐的模块组织方式

VEF Framework 应用程序遵循模块化架构模式，将业务领域组织成独立的模块。这种模式在生产应用中得到验证，提供了清晰的关注点分离。

**目录结构：**

```
my-app/
├── cmd/
│   └── server/
│       └── main.go           # 应用入口 - 组合所有模块
├── configs/
│   └── application.toml       # 配置文件
└── internal/
    ├── auth/                  # 认证提供者
    │   ├── module.go          # 认证模块定义
    │   ├── user_loader.go     # UserLoader 实现
    │   └── user_info_loader.go
    ├── sys/                   # 系统/管理功能
    │   ├── models/            # 数据模型
    │   ├── payloads/          # API 参数
    │   ├── resources/         # API 资源
    │   ├── schemas/           # 从模型生成（通过 vef-cli）
    │   └── module.go          # 系统模块定义
    ├── [domain]/              # 业务领域（如 order、inventory）
    │   ├── models/
    │   ├── payloads/
    │   ├── resources/
    │   ├── schemas/
    │   └── module.go
    ├── vef/                   # VEF 框架集成
    │   ├── module.go
    │   ├── build_info.go      # 生成的构建元数据
    │   ├── *_subscriber.go    # 事件订阅者
    │   └── *_loader.go        # 数据加载器
    └── web/                   # SPA 前端集成（可选）
        ├── dist/              # 静态资源
        └── module.go
```

### 模块组合

每个模块导出一个 `vef.Module()`，封装其依赖和资源。main.go 按依赖顺序组合这些模块：

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
        ivef.Module,     // 框架集成（应用内 vef 模块）
        web.Module,      // SPA 服务（可选）
        auth.Module,     // 认证提供者
        sys.Module,      // 系统资源
        // 在此添加您的业务领域模块
    )
}
```

**模块定义示例：**

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
    // 注册其他资源和服务
)
```

**此模式的优势：**
- **清晰边界**：每个模块拥有自己的模型、API 和业务逻辑
- **可测试性**：模块可以独立测试
- **可扩展性**：易于添加新领域而不影响现有代码
- **可维护性**：变更局限于特定模块

## 架构设计

### 单一端点设计

VEF 采用单一端点方式，所有 Api 请求通过 `POST /api`（或 `POST /openapi` 用于外部集成）。

**请求格式：**

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

**响应格式：**

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "page": 1,
    "size": 20,
    "total": 100,
    "items": [...]
  }
}
```

参数与元数据：
- `params`：业务参数（如查询筛选、创建/更新字段）。定义的结构体需嵌入 `api.P`。
- `meta`：请求级控制信息（如 `find_page` 的分页、导入导出的格式等）。定义的结构体需嵌入 `api.M`（例如 `page.Pageable`）。

### 依赖注入

VEF 使用 Uber FX 进行依赖注入。通过辅助函数注册组件：

```go
vef.Run(
    vef.ProvideApiResource(NewUserResource),
    vef.Provide(NewUserService),
)
```

## 定义数据模型

所有模型应嵌入 `orm.Model` 以获得自动审计字段管理：

```go
package models

import (
    "github.com/ilxqx/vef-framework-go/null"
    "github.com/ilxqx/vef-framework-go/orm"
)

type User struct {
    orm.BaseModel `bun:"table:sys_user,alias:su"`
    orm.Model     
    
    Username string      `json:"username" validate:"required,alphanum,max=32" label:"用户名"`
    Email    null.String `json:"email" validate:"omitempty,email,max=64" label:"邮箱"`
    IsActive bool        `json:"isActive"`
}
```

**字段标签说明：**

- `bun` - Bun ORM 配置（表名、列映射、关联关系）
- `json` - JSON 序列化字段名
- `validate` - 验证规则（[go-playground/validator](https://github.com/go-playground/validator)）
- `label` - 错误消息中显示的字段名

**审计字段**（`orm.Model` 自动维护）：

- `id` - 主键（20 字符的 XID，base32 编码）
- `created_at`, `created_by` - 创建时间戳和用户 ID
- `created_by_name` - 创建者名称（仅扫描，不存储到数据库）
- `updated_at`, `updated_by` - 最后更新时间戳和用户 ID
- `updated_by_name` - 更新者名称（仅扫描，不存储到数据库）

说明：数据库列名使用下划线命名（如 `created_at`），JSON 字段使用驼峰命名（如 `createdAt`），以模型中的标签为准。
**可空类型：** 使用 `null.String`、`null.Int`、`null.Bool` 等处理可空字段。

### 布尔列的字段类型

是否使用 `bool`、`sql.Bool` 或 `null.Bool` 取决于目标数据库与是否需要三态（NULL）。

核心建议：
- 大多数场景推荐使用原生 `bool`。主流数据库已原生支持布尔类型，直接映射最简洁。
- 当需要将布尔值以数值型（0/1）存储（如 tinyint/smallint），或者目标数据库不支持原生布尔类型时，使用 `sql.Bool`（非空）或 `null.Bool`（可空）。
- 需要三态（NULL/false/true）时使用 `null.Bool`，其数据库序列化为 NULL/1/0。

决策指南：

| 使用场景 | 首选类型 | 数据库列类型 |
|---------|----------|--------------|
| 数据库原生布尔、非空列 | `bool` | boolean/布尔原生类型 |
| 可空布尔（三态） | `null.Bool` | boolean 或 数值型（常见为 smallint/tinyint） |
| 兼容无布尔数据库，或强制数值存储 0/1 | `sql.Bool`（非空）/ `null.Bool`（可空） | smallint/tinyint（0/1） |
| 仅 Go 计算字段（不入库） | `bool` 且 `bun:"-"` | N/A |

类型说明与示例：

1) 原生 `bool` —— 推荐用于原生布尔列
```go
type User struct {
    orm.Model
    // 数据库：布尔原生类型；是否 NOT NULL 由列定义决定
    IsActive bool `json:"isActive"` // 使用原生布尔时通常无需额外 bun 标签
}
```

2) `sql.Bool` —— 数值化存储（0/1），用于兼容性
```go
import "github.com/ilxqx/vef-framework-go/sql"

type User struct {
    orm.Model
    // 数据库：以数值 0/1 存储；适用于无原生布尔或需统一数值化存储的场景
    IsActive sql.Bool `json:"isActive" bun:"type:smallint,notnull,default:0"`
    IsLocked sql.Bool `json:"isLocked" bun:"type:smallint,notnull,default:0"`
}
```
如果项目不需要兼容无布尔数据库，直接使用 `bool` 更简单。

3) `null.Bool` —— 三态（NULL/false/true）
```go
import "github.com/ilxqx/vef-framework-go/null"

type User struct {
    orm.Model
    // 数据库：允许为 NULL；序列化为 NULL/0/1（为最大兼容性建议列类型使用数值型）
    IsVerified null.Bool `json:"isVerified" bun:"type:smallint"`
}
```
三态语义：
- `null.Bool{Valid: false}` → 数据库为 NULL
- `null.Bool{Valid: true, Bool: false}` → 0/false
- `null.Bool{Valid: true, Bool: true}` → 1/true

4) 仅 Go 字段（不入库）
```go
type User struct {
    orm.Model
    Username string `json:"username"`

    // 计算字段 —— 不入库
    HasPermissions bool `json:"hasPermissions" bun:"-"`
}
```

常见模式：
```go
// 使用原生布尔（推荐）
type UserNative struct {
    orm.Model
    IsActive bool        `json:"isActive"`
    IsLocked bool        `json:"isLocked"`
    IsEmailVerified null.Bool `json:"isEmailVerified"` // 需要 NULL 时使用
}

// 为兼容性使用数值化存储
type UserNumeric struct {
    orm.Model
    IsActive sql.Bool        `json:"isActive" bun:"type:smallint,notnull,default:0"`
    IsLocked sql.Bool        `json:"isLocked" bun:"type:smallint,notnull,default:0"`
    IsEmailVerified null.Bool `json:"isEmailVerified" bun:"type:smallint"`
}
```

## 构建 CRUD Api

### 资源命名最佳实践

在定义 API 资源时，遵循一致的命名约定以避免冲突并明确 API 的所有权。

**推荐模式：`{app}/{domain}/{entity}`**

这种三级命名空间模式在生产应用中广泛使用，提供了多项优势：

```go
// 带应用命名空间的良好示例
api.NewResource("smp/sys/user")           // 系统用户资源
api.NewResource("smp/md/organization")    // 主数据组织
api.NewResource("erp/order/item")         // 清晰的领域分离

// 单应用项目中可接受
api.NewResource("sys/user")               // 无应用命名空间

// 避免使用 - 过于泛化，存在冲突风险
api.NewResource("user")                   // ❌ 无命名空间
```

**应用命名空间的优势：**

- **防止冲突**：避免在共享部署或合并代码库时出现 API 资源冲突
- **明确所有权**：立即识别哪个应用拥有该资源
- **模块化**：支持多个应用或微服务使用同一框架
- **迁移安全**：在重构时易于识别和迁移资源

**框架保留的命名空间：**

以下资源命名空间保留给系统 API，不得用于自定义 API 定义：

- `security/auth` - 认证 API
- `sys/storage` - 存储 API
- `sys/monitor` - 监控 API

使用这些保留名称会因 API 定义重复而导致应用启动失败。

### 第一步：定义参数结构

**查询参数：**

```go
package payloads

import "github.com/ilxqx/vef-framework-go/api"

type UserSearch struct {
    api.P
    Keyword string `json:"keyword" search:"contains,column=username|email"`
    IsActive *bool `json:"isActive" search:"eq"`
}
```

**创建/更新参数：**

```go
type UserParams struct {
    api.P
    Id       string      `json:"id"` // 更新操作时必需

    Username string      `json:"username" validate:"required,alphanum,max=32" label:"用户名"`
    Email    null.String `json:"email" validate:"omitempty,email,max=64" label:"邮箱"`
    IsActive bool        `json:"isActive"`
}
```

**分离创建和更新参数：**

当创建和更新操作具有不同的验证要求时，使用结构体嵌入来共享公共字段，同时允许特定于操作的验证：

```go
// 共享字段
type UserParams struct {
    api.P
    Id       string
    Username string      `json:"username" validate:"required,alphanum,max=32" label:"用户名"`
    Email    null.String `json:"email" validate:"omitempty,email,max=64" label:"邮箱"`
    IsActive bool        `json:"isActive"`
}

// 创建需要密码
type UserCreateParams struct {
    UserParams      `json:",inline"`
    Password        string `json:"password" validate:"required,min=6,max=16" label:"密码"`
    PasswordConfirm string `json:"passwordConfirm" validate:"required,eqfield=Password" label:"确认密码"`
}

// 更新有可选密码
type UserUpdateParams struct {
    UserParams      `json:",inline"`
    Password        null.String `json:"password" validate:"omitempty,min=6,max=16" label:"密码"`
    PasswordConfirm null.String `json:"passwordConfirm" validate:"omitempty,eqfield=Password" label:"确认密码"`
}
```

然后在您的资源中使用特定参数：

```go
CreateApi: apis.NewCreateApi[models.User, payloads.UserCreateParams](),
UpdateApi: apis.NewUpdateApi[models.User, payloads.UserUpdateParams](),
```

**优势：**
- **类型安全的验证**：创建和更新的不同规则（必需与可选密码）
- **清晰的契约**：API 要求在代码中是明确的
- **更好的错误消息**：验证错误与操作的实际要求匹配
- **代码重用**：公共字段仅定义一次并嵌入

### 第二步：创建 Api 资源

> **⚠️ 重要：系统保留的 API 命名空间**
>
> 框架为系统 API 保留了以下资源命名空间。**请勿**在自定义 API 定义中使用这些资源名称，否则会与内置框架功能冲突，导致应用启动失败:
>
> - `security/auth` - 认证 API（login, logout, refresh, get_user_info）
> - `sys/storage` - 存储 API（upload, get_presigned_url, delete_temp, stat, list）
> - `sys/monitor` - 监控 API（get_overview, get_cpu, get_memory, get_disk 等）
>
> 框架会自动检测重复的 API 定义，如果发现冲突将拒绝启动。请使用自定义的资源命名空间，如 `app/`、`custom/` 或您自己的领域特定前缀，以避免冲突。

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
        Resource: api.NewResource("smp/sys/user"),  // ✓ 使用 应用/领域/实体 命名避免冲突
        FindAllApi: apis.NewFindAllApi[models.User, payloads.UserSearch](),
        FindPageApi: apis.NewFindPageApi[models.User, payloads.UserSearch](),
        CreateApi: apis.NewCreateApi[models.User, payloads.UserParams](),
        UpdateApi: apis.NewUpdateApi[models.User, payloads.UserParams](),
        DeleteApi: apis.NewDeleteApi[models.User](),
    }
}
```

### 第三步：注册资源

```go
func main() {
    vef.Run(
        vef.ProvideApiResource(resources.NewUserResource),
    )
}
```

### 预置 Api 列表

| 接口 | 描述 | Action |
|-----|------|--------|
| FindOneApi | 查询单条记录 | find_one |
| FindAllApi | 查询全部记录 | find_all |
| FindPageApi | 分页查询 | find_page |
| CreateApi | 创建记录 | create |
| UpdateApi | 更新记录 | update |
| DeleteApi | 删除记录 | delete |
| CreateManyApi | 批量创建 | create_many |
| UpdateManyApi | 批量更新 | update_many |
| DeleteManyApi | 批量删除 | delete_many |
| FindTreeApi | 树形查询 | find_tree |
| FindOptionsApi | 选项列表(label/value) | find_options |
| FindTreeOptionsApi | 树形选项 | find_tree_options |
| ImportApi | 导入 Excel/CSV | import |
| ExportApi | 导出 Excel/CSV | export |

### Api Builder 方法

使用流式构建器方法配置 Api 行为：

```go
CreateApi: apis.NewCreateApi[User, UserParams]().
    Action("create_user").             // 自定义操作名
    Public().                          // 无需认证
    PermToken("sys.user.create").      // 权限令牌
    EnableAudit().                     // 启用审计日志
    Timeout(10 * time.Second).         // 请求超时
    RateLimit(10, 1*time.Minute).      // 每分钟 10 次请求
```

**注意：** FindApi 类型（FindOneApi、FindAllApi、FindPageApi、FindTreeApi、FindOptionsApi、FindTreeOptionsApi、ExportApi）具有额外的配置方法。详见 [FindApi 配置方法](#findapi-配置方法)。

### FindApi 配置方法

所有 FindApi 类型（FindOneApi、FindAllApi、FindPageApi、FindTreeApi、FindOptionsApi、FindTreeOptionsApi、ExportApi）都支持使用流式方法的统一查询配置系统。这些方法允许您自定义查询行为、添加条件、配置排序和处理结果。

#### 通用配置方法

| 方法 | 说明 | 默认 QueryPart | 适用 API |
|------|------|---------------|----------|
| `WithProcessor` | 设置查询结果的后处理函数 | N/A | 所有 FindApi |
| `WithOptions` | 添加多个 FindApiOptions | N/A | 所有 FindApi |
| `WithSelect` | 添加列到 SELECT 子句 | QueryRoot | 所有 FindApi |
| `WithSelectAs` | 添加带别名的列到 SELECT 子句 | QueryRoot | 所有 FindApi |
| `WithDefaultSort` | 设置默认排序规范 | QueryRoot | 所有 FindApi |
| `WithCondition` | 使用 ConditionBuilder 添加 WHERE 条件 | QueryRoot | 所有 FindApi |
| `WithRelation` | 添加关联查询 | QueryRoot | 所有 FindApi |
| `WithAuditUserNames` | 获取审计用户名（created_by_name、updated_by_name） | QueryRoot | 所有 FindApi |
| `WithQueryApplier` | 添加自定义查询应用函数 | QueryRoot | 所有 FindApi |
| `DisableDataPerm` | 禁用数据权限过滤 | N/A | 所有 FindApi |

**WithProcessor 示例：**

`Processor` 函数在数据库查询完成后、将结果返回给客户端之前执行。这允许您转换、丰富或过滤查询结果。

常见用例：
- **数据脱敏**：隐藏敏感信息（密码、令牌）
- **计算字段**：基于现有数据添加计算值
- **嵌套结构转换**：将扁平数据转换为层次结构
- **聚合计算**：计算统计信息或摘要

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithProcessor(func(users []User, search UserSearch, ctx fiber.Ctx) any {
        // 数据脱敏
        for i := range users {
            users[i].Password = "***"
            users[i].ApiToken = ""
        }
        return users
    }),

// 示例：分页结果中添加计算字段（处理器接收 items 切片）
FindPageApi: apis.NewFindPageApi[Order, OrderSearch]().
    WithProcessor(func(items []Order, search OrderSearch, ctx fiber.Ctx) any {
        for i := range items {
            // 计算总金额
            items[i].TotalAmount = items[i].Quantity * items[i].UnitPrice
        }
        return items
    }),

// 示例：嵌套结构转换
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithProcessor(func(users []User, search UserSearch, ctx fiber.Ctx) any {
        // 按部门分组用户
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

**WithSelect / WithSelectAs 示例：**

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithSelect("username").
    WithSelectAs("email_address", "email"),
```

**WithDefaultSort 示例：**

```go
FindPageApi: apis.NewFindPageApi[User, UserSearch]().
    WithDefaultSort(&sort.OrderSpec{
        Column:    "created_at",
        Direction: sort.OrderDesc,
    }),

// 生产模式：使用 schema 生成的列名以实现类型安全
import "my-app/internal/sys/schemas"

FindPageApi: apis.NewFindPageApi[User, UserSearch]().
    WithDefaultSort(&sort.OrderSpec{
        Column:    schemas.User.CreatedAt(true), // 类型安全的列名，带表前缀
        Direction: sort.OrderDesc,
    }),

// 对于树形结构，使用 sort_order 字段
FindTreeApi: apis.NewFindTreeApi[Menu, MenuSearch](buildMenuTree).
    WithDefaultSort(&sort.OrderSpec{
        Column:    schemas.Menu.SortOrder(true),
        Direction: sort.OrderAsc,
    }),
```

传入空参数可禁用默认排序：

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithDefaultSort(), // 禁用默认排序
```

**WithCondition 示例：**

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithCondition(func(cb orm.ConditionBuilder) {
        cb.Equals("is_deleted", false)
        cb.Equals("is_active", true)
    }),
```

**WithRelation 示例：**

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithRelation(&orm.RelationSpec{
        // 关联 Profile 模型；外键/主键按约定自动解析
        Model: (*Profile)(nil),
        // 可选：自定义别名/选择列
        // Alias: "p",
        SelectedColumns: []orm.ColumnInfo{
            {Name: "name", AutoAlias: true},
            {Name: "email", AutoAlias: true},
        },
    }),
```

**WithAuditUserNames 示例：**

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithAuditUserNames(&User{}), // 默认使用 "name" 列

// 或指定自定义列名
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithAuditUserNames(&User{}, "username"),

// 生产模式：使用包级别的模型实例
// 在 models 包中：var UserModel = &User{}
FindPageApi: apis.NewFindPageApi[User, UserSearch]().
    WithAuditUserNames(models.UserModel), // 推荐用于一致性
```

**WithQueryApplier 示例：**

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithQueryApplier(func(query orm.SelectQuery, search UserSearch, ctx fiber.Ctx) error {
        // 自定义查询逻辑
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

**DisableDataPerm 示例：**

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    DisableDataPerm(), // 必须在 API 注册前调用
```

**重要提示：** `DisableDataPerm()` 必须在 API 注册之前调用（在 `Setup` 方法执行之前）。它应该在 `NewFindXxxApi()` 之后立即链式调用。默认情况下，数据权限过滤是启用的，并在 `Setup` 期间自动应用。

#### QueryPart 系统

配置方法中的 `parts` 参数指定选项应用于查询的哪个部分。这对于使用递归 CTE（公用表表达式）的树形 API 尤为重要。

| QueryPart | 说明 | 使用场景 |
|-----------|------|----------|
| `QueryRoot` | 外层/根查询 | 排序、限制、最终过滤 |
| `QueryBase` | 基础查询（在 CTE 中） | 初始条件、起始节点 |
| `QueryRecursive` | 递归查询（在 CTE 中） | 递归遍历配置 |
| `QueryAll` | 所有查询部分 | 列选择、关联 |

**默认行为：**

- `WithSelect`、`WithSelectAs`、`WithRelation`：默认为 `QueryRoot`（应用于主/根查询）
- `WithCondition`、`WithQueryApplier`、`WithDefaultSort`：默认为 `QueryRoot`（仅应用于根查询）

**普通查询示例：**

```go
FindAllApi: apis.NewFindAllApi[User, UserSearch]().
    WithSelect("username").              // 应用于 QueryRoot（主查询）
    WithCondition(func(cb orm.ConditionBuilder) {
        cb.Equals("is_active", true)     // 应用于 QueryRoot（主查询）
    }),
```

**树形查询示例：**

```go
FindTreeApi: apis.NewFindTreeApi[Category, CategorySearch](buildTree).
    // 为基础查询和递归查询选择列
    WithSelect("sort", apis.QueryBase, apis.QueryRecursive).
    
    // 仅过滤起始节点
    WithCondition(func(cb orm.ConditionBuilder) {
        cb.IsNull("parent_id")           // 仅应用于 QueryBase
    }, apis.QueryBase).
    
    // 向递归遍历添加条件
    WithCondition(func(cb orm.ConditionBuilder) {
        cb.Equals("is_active", true)     // 应用于 QueryRecursive
    }, apis.QueryRecursive),
```

#### 树形查询配置

`FindTreeApi` 和 `FindTreeOptionsApi` 使用递归 CTE（公用表表达式）查询层次数据。理解 QueryPart 如何应用于递归查询的不同部分对于正确配置至关重要。

**递归 CTE 结构：**

```sql
WITH RECURSIVE tree AS (
    -- QueryBase：根节点的初始查询
    SELECT * FROM categories WHERE parent_id IS NULL
    
    UNION ALL
    
    -- QueryRecursive：与 CTE 连接的递归查询
    SELECT c.* FROM categories c
    INNER JOIN tree t ON c.parent_id = t.id
)
-- QueryRoot：从 CTE 的最终 SELECT
SELECT * FROM tree ORDER BY sort
```

**树形查询中的 QueryPart 行为：**

- `WithSelect` / `WithSelectAs`：默认为 `QueryBase` 和 `QueryRecursive`（UNION 两部分的列必须一致）
- `WithCondition` / `WithQueryApplier`：默认仅为 `QueryBase`（过滤起始节点）
- `WithRelation`：默认为 `QueryBase` 和 `QueryRecursive`（两部分都需要连接）
- `WithDefaultSort`：应用于 `QueryRoot`（排序最终结果）

**完整的树形查询示例：**

```go
FindTreeApi: apis.NewFindTreeApi[Category, CategorySearch](
    func(categories []Category) []Category {
        // 从扁平列表构建树结构
        return buildCategoryTree(categories)
    },
).
    // 向基础查询和递归查询添加自定义列
    WithSelect("sort", apis.QueryBase, apis.QueryRecursive).
    WithSelect("icon", apis.QueryBase, apis.QueryRecursive).
    
    // 过滤起始节点（仅活动的根分类）
    WithCondition(func(cb orm.ConditionBuilder) {
        cb.Equals("is_active", true)
        cb.IsNull("parent_id")
    }, apis.QueryBase).
    
    // 向两个查询添加关联
    WithRelation(&orm.RelationSpec{
        Model: (*Metadata)(nil),
        SelectedColumns: []orm.ColumnInfo{
            {Name: "icon", AutoAlias: true},
            {Name: "sort_order", Alias: "sortOrder"},
        },
    }, apis.QueryBase, apis.QueryRecursive).
    
    // 获取审计用户名
    WithAuditUserNames(&User{}).
    
    // 排序最终结果
    WithDefaultSort(&sort.OrderSpec{
        Column:    "sort",
        Direction: sort.OrderAsc,
    }),
```

**FindTreeOptionsApi 配置：**

`FindTreeOptionsApi` 遵循与 `FindTreeApi` 相同的配置模式：

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

#### API 特定配置方法

**FindPageApi：**

```go
FindPageApi: apis.NewFindPageApi[User, UserSearch]().
    WithDefaultPageSize(20), // 设置默认分页大小（当请求未指定或无效时使用）
```

**FindOptionsApi：**

```go
FindOptionsApi: apis.NewFindOptionsApi[User, UserSearch]().
    WithDefaultColumnMapping(&apis.DataOptionColumnMapping{
        LabelColumn:       "name",        // 选项标签列（默认："name"）
        ValueColumn:       "id",          // 选项值列（默认："id"）
        DescriptionColumn: "description", // 可选描述列
    }),

// 高级用法：在选项中包含额外的元数据
FindOptionsApi: apis.NewFindOptionsApi[Menu, MenuSearch]().
    WithDefaultColumnMapping(&apis.DataOptionColumnMapping{
        LabelColumn:       "name",
        ValueColumn:       "id",
        DescriptionColumn: "remark",
        MetaColumns: []string{
            "type",                    // 菜单类型（D=目录，M=菜单，B=按钮）
            "icon",                    // 图标标识
            "sort_order AS sortOrder", // 显示顺序（带别名）
        },
    }),
```

**FindTreeApi：**

对于层次数据结构，使用 `FindTreeApi` 配合 `treebuilder` 包将扁平数据库结果转换为嵌套树结构：

```go
import "github.com/ilxqx/vef-framework-go/treebuilder"

FindTreeApi: apis.NewFindTreeApi[models.Organization, payloads.OrganizationSearch](
    buildOrganizationTree,
).
    WithIdColumn("id").              // ID 列名（默认："id"）
    WithParentIdColumn("parent_id"). // 父 ID 列名（默认："parent_id"）
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

**模型要求：**

您的模型必须具有：
- 父 ID 字段（通常为 `null.String` 以支持根节点）
- 子节点字段（同类型模型的切片，标记为 `bun:"-"` 因为它是计算的）

```go
type Organization struct {
    orm.Model
    Name     string          `json:"name"`
    ParentId null.String     `json:"parentId" bun:"type:varchar(20)"` // 根节点为 NULL
    Children []Organization  `json:"children" bun:"-"`                // 计算字段，不在数据库中
}
```

`treebuilder.Build` 函数处理从扁平列表到层次结构的转换，正确地将子节点嵌套在其父节点下。

**FindTreeOptionsApi：**

结合选项和树形配置以返回层次选项列表：

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

树形选项 API 自动使用内部树构建器将扁平结果转换为嵌套选项结构，非常适合级联选择器或层次菜单。

**ExportApi：**

```go
ExportApi: apis.NewExportApi[User, UserSearch]().
    WithDefaultFormat("excel").                   // 默认导出格式："excel" 或 "csv"
    WithExcelOptions(&excel.ExportOptions{        // Excel 特定选项
        SheetName: "Users",
    }).
    WithCsvOptions(&csv.ExportOptions{            // CSV 特定选项
        Delimiter: ',',
    }).
    WithPreExport(func(users []User, search UserSearch, ctx fiber.Ctx, db orm.Db) error {
        // 导出前修改数据（例如数据脱敏）
        for i := range users {
            users[i].Password = "***"
        }
        return nil
    }).
    WithFilenameBuilder(func(search UserSearch, ctx fiber.Ctx) string {
        // 生成动态文件名
        return fmt.Sprintf("users_%s", time.Now().Format("20060102"))
    }),
```

### Pre/Post 钩子

在 CRUD 操作前后添加自定义业务逻辑：

```go
CreateApi: apis.NewCreateApi[User, UserParams]().
    WithPreCreate(func(model *User, params *UserParams, ctx fiber.Ctx, db orm.Db) error {
        // 创建用户前对密码进行哈希
        hashed, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
        if err != nil {
            return err
        }
        model.Password = string(hashed)
        return nil
    }).
    WithPostCreate(func(model *User, params *UserParams, ctx fiber.Ctx, tx orm.Db) error {
        // 用户创建后发送欢迎邮件（在事务内执行）
        return sendWelcomeEmail(model.Email)
    }),
```

可用的钩子：

**单条记录操作：**
- `WithPreCreate`、`WithPostCreate` - 创建前/后（`WithPostCreate` 在事务内运行）
- `WithPreUpdate`、`WithPostUpdate` - 更新前/后（接收旧模型和新模型，`WithPostUpdate` 在事务内运行）
- `WithPreDelete`、`WithPostDelete` - 删除前/后（`WithPostDelete` 在事务内运行）

**批量操作：**
- `WithPreCreateMany`、`WithPostCreateMany` - 批量创建前/后（`WithPostCreateMany` 在事务内运行）
- `WithPreUpdateMany`、`WithPostUpdateMany` - 批量更新前/后（接收旧模型数组和新模型数组，`WithPostUpdateMany` 在事务内运行）
- `WithPreDeleteMany`、`WithPostDeleteMany` - 批量删除前/后（`WithPostDeleteMany` 在事务内运行）

**导入导出操作：**
- `WithPreImport`、`WithPostImport` - 导入前/后（`WithPreImport` 用于验证，`WithPostImport` 在事务内运行）
- `WithPreExport` - 导出前（用于数据格式化）

**生产模式：**

```go
// 系统用户保护 - 防止删除关键系统用户
DeleteApi: apis.NewDeleteApi[User]().
    WithPreDelete(func(model *User, ctx fiber.Ctx, db orm.Db) error {
        // 保护系统内部用户不被删除
        switch model.Username {
        case "system", "anonymous", "cron":
            return result.Err("禁止删除系统内部用户")
        }
        return nil
    }),

// 条件密码哈希 - 仅在密码被修改时进行哈希
UpdateApi: apis.NewUpdateApi[User, UserUpdateParams]().
    WithPreUpdate(func(oldModel *User, newModel *User, params *UserUpdateParams, ctx fiber.Ctx, db orm.Db) error {
        // 仅在密码被更新时进行哈希
        if params.Password.Valid && params.Password.String != "" {
            hashed, err := bcrypt.GenerateFromPassword([]byte(params.Password.String), bcrypt.DefaultCost)
            if err != nil {
                return err
            }
            newModel.Password = string(hashed)
        } else {
            // 保留现有密码
            newModel.Password = oldModel.Password
        }
        return nil
    }),

// 业务验证 - 在操作前验证业务规则
CreateApi: apis.NewCreateApi[Order, OrderParams]().
    WithPreCreate(func(model *Order, params *OrderParams, ctx fiber.Ctx, db orm.Db) error {
        // 验证订单总额是否匹配项目总额
        if model.TotalAmount <= 0 {
            return result.Err("订单总额必须大于零")
        }

        // 检查库存可用性
        if !checkInventoryAvailable(model.Items) {
            return result.Err("一个或多个商品库存不足")
        }

        return nil
    }),
```

### 自定义处理器

#### 混合生成和自定义 API

您可以使用 `api.WithApis()` 将预构建的 CRUD API 与自定义操作结合。这允许您使用特定领域的操作扩展资源，同时保持框架的约定。

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
                    EnableAudit: true,  // 为此操作启用审计日志
                },
            ),
        ),
        FindPageApi: apis.NewFindPageApi[models.Role, payloads.RoleSearch](),
        CreateApi:   apis.NewCreateApi[models.Role, payloads.RoleParams](),
        UpdateApi:   apis.NewUpdateApi[models.Role, payloads.RoleParams](),
        DeleteApi:   apis.NewDeleteApi[models.Role](),
    }
}

// find_role_permissions 操作的自定义处理器方法
func (r *RoleResource) FindRolePermissions(
    ctx fiber.Ctx,
    db orm.Db,
    params payloads.RolePermissionQuery,
) error {
    // 自定义业务逻辑
    // ...
    return result.Ok(permissions).Response(ctx)
}

// save_role_permissions 操作的自定义处理器方法
func (r *RoleResource) SaveRolePermissions(
    ctx fiber.Ctx,
    db orm.Db,
    params payloads.RolePermissionParams,
) error {
    // 基于事务的自定义逻辑
    return db.RunInTx(ctx.Context(), func(txCtx context.Context, tx orm.Db) error {
        // 在事务中保存权限
        // ...
        return nil
    })
}
```

**关键要点：**

- **方法命名**：处理器方法名必须为 PascalCase，与 snake_case 操作名匹配（例如 `find_role_permissions` → `FindRolePermissions`）
- **API Spec 配置**：每个自定义操作都可以有自己的配置（权限、审计、速率限制）
- **注入规则**：自定义处理器方法遵循与生成的处理器相同的参数注入规则
- **混合 API**：您可以在同一资源中自由混合生成的 CRUD API 和自定义操作

#### 简单自定义处理器

通过在资源上定义方法添加自定义端点：

```go
func (r *UserResource) ResetPassword(
    ctx fiber.Ctx,
    db orm.Db,
    logger log.Logger,
    principal *security.Principal,
    params ResetPasswordParams,
) error {
    logger.Infof("用户 %s 正在重置密码", principal.Id)
    
    // 自定义业务逻辑
    var user models.User
    if err := db.NewSelect().
        Model(&user).
        Where(func(cb orm.ConditionBuilder) {
            cb.Equals("id", principal.Id)
        }).
        Scan(ctx.Context()); err != nil {
        return err
    }
    
    // 更新密码
    // ...
    
    return result.Ok().Response(ctx)
}
```

**可注入参数类型：**

- `fiber.Ctx` - HTTP 上下文
- `orm.Db` - 数据库连接
- `log.Logger` - 日志记录器
- `mold.Transformer` - 数据转换器
- `*security.Principal` - 当前认证用户
- `page.Pageable` - 分页参数
- 嵌入 `api.P` 的自定义结构体
- 嵌入 `api.M` 的自定义结构体（请求元数据）
- Resource 结构体字段（直接字段、带 `api:"in"` 标签的字段或嵌入的结构体）

**Resource 字段注入示例：**

```go
type UserResource struct {
    api.Resource
    userService *UserService  // Resource 字段
}

func NewUserResource(userService *UserService) api.Resource {
    return &UserResource{
        Resource: api.NewResource("sys/user"),
        userService: userService,
    }
}

// Handler 可以直接注入 userService
func (r *UserResource) SendNotification(
    ctx fiber.Ctx,
    service *UserService,  // 从 r.userService 注入
    params NotificationParams,
) error {
    return service.SendEmail(params.Email, params.Message)
}
```

**为什么要使用参数注入而不是直接使用 `r.userService`？**

如果你的服务实现了 `log.LoggerConfigurable[T]` 接口，框架在注入服务时会自动调用 `WithLogger` 方法，提供请求范围的日志记录器。这样每个请求都可以拥有自己的日志上下文，包含请求 ID 等上下文信息。

```go
type UserService struct {
    logger log.Logger
}

// 实现 log.LoggerConfigurable[*UserService] 接口
func (s *UserService) WithLogger(logger log.Logger) *UserService {
    return &UserService{logger: logger}
}

func (s *UserService) SendEmail(email, message string) error {
    s.logger.Infof("发送邮件到 %s", email)  // 请求范围的日志记录器
    // ...
}
```

## 数据库操作

### 查询构建器

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

### 条件构建器方法

构建类型安全的查询条件：

- `Equals(column, value)` - 等于
- `NotEquals(column, value)` - 不等于
- `GreaterThan(column, value)` - 大于
- `GreaterThanOrEquals(column, value)` - 大于等于
- `LessThan(column, value)` - 小于
- `LessThanOrEquals(column, value)` - 小于等于
- `Contains(column, value)` - 包含（LIKE %value%）
- `StartsWith(column, value)` - 开头匹配（LIKE value%）
- `EndsWith(column, value)` - 结尾匹配（LIKE %value）
- `In(column, values)` - IN 子句
- `Between(column, min, max)` - BETWEEN 子句
- `IsNull(column)` - IS NULL
- `IsNotNull(column)` - IS NOT NULL
- `Or(conditions...)` - OR 多个条件

### Search 标签

使用 `search` 标签自动应用查询条件：

```go
type UserSearch struct {
    api.P
    Username string `search:"eq"`                                    // username = ?
    Email    string `search:"contains"`                              // email LIKE ?
    Age      int    `search:"gte"`                                   // age >= ?
    Status   string `search:"in"`                                    // status IN (?)
    Keyword  string `search:"contains,column=username|email|name"`   // 搜索多个列
}
```

**支持的操作符：**

**比较操作符：**
| 标签 | SQL 操作符 | 说明 |
|-----|-----------|------|
| `eq` | = | 等于 |
| `neq` | != | 不等于 |
| `gt` | > | 大于 |
| `gte` | >= | 大于等于 |
| `lt` | < | 小于 |
| `lte` | <= | 小于等于 |

**范围操作符：**
| 标签 | SQL 操作符 | 说明 |
|-----|-----------|------|
| `between` | BETWEEN | 范围内 |
| `notBetween` | NOT BETWEEN | 不在范围内 |

**集合操作符：**
| 标签 | SQL 操作符 | 说明 |
|-----|-----------|------|
| `in` | IN | 在列表中 |
| `notIn` | NOT IN | 不在列表中 |

**空值检查操作符：**
| 标签 | SQL 操作符 | 说明 |
|-----|-----------|------|
| `isNull` | IS NULL | 为空 |
| `isNotNull` | IS NOT NULL | 不为空 |

**字符串匹配（区分大小写）：**
| 标签 | SQL 操作符 | 说明 |
|-----|-----------|------|
| `contains` | LIKE %?% | 包含 |
| `notContains` | NOT LIKE %?% | 不包含 |
| `startsWith` | LIKE ?% | 开头匹配 |
| `notStartsWith` | NOT LIKE ?% | 开头不匹配 |
| `endsWith` | LIKE %? | 结尾匹配 |
| `notEndsWith` | NOT LIKE %? | 结尾不匹配 |

**字符串匹配（不区分大小写）：**
| 标签 | SQL 操作符 | 说明 |
|-----|-----------|------|
| `iContains` | ILIKE %?% | 包含（不区分大小写） |
| `iNotContains` | NOT ILIKE %?% | 不包含（不区分大小写） |
| `iStartsWith` | ILIKE ?% | 开头匹配（不区分大小写） |
| `iNotStartsWith` | NOT ILIKE ?% | 开头不匹配（不区分大小写） |
| `iEndsWith` | ILIKE %? | 结尾匹配（不区分大小写） |
| `iNotEndsWith` | NOT ILIKE %? | 结尾不匹配（不区分大小写） |

### 事务处理

在事务中执行多个操作：

```go
err := db.RunInTx(ctx.Context(), func(txCtx context.Context, tx orm.Db) error {
    // 插入用户
    _, err := tx.NewInsert().Model(&user).Exec(txCtx)
    if err != nil {
        return err // 自动回滚
    }

    // 更新关联记录
    _, err = tx.NewUpdate().Model(&profile).WherePk().Exec(txCtx)
    return err // 返回 nil 自动提交，返回错误自动回滚
})
```

## 认证与授权

### 认证方式

VEF 支持多种认证策略：

1. **Jwt 认证**（默认）- Bearer token 或查询参数 `?__accessToken=xxx`
2. **OpenApi 签名认证** - 用于外部应用，使用 HMAC 签名
3. **密码认证** - 用户名密码登录

### 实现用户加载器

实现 `security.UserLoader` 接口以集成您的用户系统：

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
        Roles: []string{"user"}, // 从数据库加载
    }
    
    return principal, user.Password, nil // 返回哈希后的密码
}

func (l *MyUserLoader) LoadById(ctx context.Context, id string) (*security.Principal, error) {
    // 类似的实现
}

func NewMyUserLoader(db orm.Db) *MyUserLoader {
    return &MyUserLoader{db: db}
}

// 在 main.go 中注册
func main() {
    vef.Run(
        vef.Provide(NewMyUserLoader),
    )
}
```

### 权限控制

在 Api 上设置权限令牌：

```go
CreateApi: apis.NewCreateApi[User, UserParams]().
    PermToken("sys.user.create"),
```

#### 使用内置 RBAC 实现（推荐）

框架已内置基于角色的访问控制（RBAC）实现，只需实现 `security.RolePermissionsLoader` 接口即可：

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

// LoadPermissions 加载指定角色的所有权限
// 返回 map[权限令牌]数据范围
func (l *MyRolePermissionsLoader) LoadPermissions(ctx context.Context, role string) (map[string]security.DataScope, error) {
    // 从数据库加载角色权限
    var permissions []RolePermission
    if err := l.db.NewSelect().
        Model(&permissions).
        Where(func(cb orm.ConditionBuilder) {
            cb.Equals("role_code", role)
        }).
        Scan(ctx); err != nil {
        return nil, err
    }
    
    // 构建权限令牌到数据范围的映射
    result := make(map[string]security.DataScope)
    for _, perm := range permissions {
        // 根据数据范围类型创建对应的 DataScope 实例
        var dataScope security.DataScope
        switch perm.DataScopeType {
        case "all":
            dataScope = security.NewAllDataScope()
        case "self":
            dataScope = security.NewSelfDataScope("")
        case "dept":
            dataScope = NewDepartmentDataScope() // 自定义实现
        // ... 更多自定义数据范围
        }
        
        result[perm.PermissionToken] = dataScope
    }
    
    return result, nil
}

func NewMyRolePermissionsLoader(db orm.Db) security.RolePermissionsLoader {
    return &MyRolePermissionsLoader{db: db}
}

// 在 main.go 中注册
func main() {
    vef.Run(
        vef.Provide(NewMyRolePermissionsLoader),
    )
}
```

**注意：** 框架会自动使用您提供的 `RolePermissionsLoader` 实现来初始化内置的 RBAC 权限检查器和数据权限解析器。

#### 完全自定义权限控制

如果需要实现完全自定义的权限控制逻辑（非 RBAC），可以实现 `security.PermissionChecker` 接口并替换框架的实现：

```go
type MyCustomPermissionChecker struct {
    // 自定义字段
}

func (c *MyCustomPermissionChecker) HasPermission(ctx context.Context, principal *security.Principal, permToken string) (bool, error) {
    // 自定义权限检查逻辑
    // ...
    return true, nil
}

func NewMyCustomPermissionChecker() security.PermissionChecker {
    return &MyCustomPermissionChecker{}
}

// 在 main.go 中替换框架的实现
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

### 数据权限

数据权限用于实现行级数据访问控制，限制用户只能访问特定范围的数据。

#### 内置数据范围

框架提供了两种内置的数据范围实现：

1. **AllDataScope** - 无限制访问所有数据（通常用于管理员）
2. **SelfDataScope** - 只能访问自己创建的数据

```go
import "github.com/ilxqx/vef-framework-go/security"

// 所有数据
allScope := security.NewAllDataScope()

// 仅自己创建的数据（默认使用 created_by 列）
selfScope := security.NewSelfDataScope("")

// 自定义创建者列名
selfScope := security.NewSelfDataScope("creator_id")
```

#### 使用内置 RBAC 数据权限（推荐）

框架的 RBAC 实现会自动处理数据权限。在 `RolePermissionsLoader.LoadPermissions` 中返回权限令牌对应的数据范围即可：

```go
func (l *MyRolePermissionsLoader) LoadPermissions(ctx context.Context, role string) (map[string]security.DataScope, error) {
    result := make(map[string]security.DataScope)
    
    // 为不同权限分配不同的数据范围
    result["sys.user.view"] = security.NewAllDataScope()      // 查看所有用户
    result["sys.user.edit"] = security.NewSelfDataScope("")    // 只能编辑自己创建的用户
    
    return result, nil
}
```

**数据范围优先级：** 当用户拥有多个角色，且这些角色对同一权限令牌配置了不同的数据范围时，框架会选择优先级最高的数据范围。内置优先级常量：

- `security.PrioritySelf` (10) - 仅自己创建的数据
- `security.PriorityDepartment` (20) - 部门数据
- `security.PriorityDepartmentAndSub` (30) - 部门及子部门数据
- `security.PriorityOrganization` (40) - 组织数据
- `security.PriorityOrganizationAndSub` (50) - 组织及子组织数据
- `security.PriorityCustom` (60) - 自定义数据范围
- `security.PriorityAll` (10000) - 所有数据

#### 自定义数据范围

实现 `security.DataScope` 接口来创建自定义的数据访问范围：

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
    return security.PriorityDepartment // 使用框架定义的优先级
}

func (s *DepartmentDataScope) Supports(principal *security.Principal, table *orm.Table) bool {
    // 检查表是否有 department_id 列
    field, _ := table.Field("department_id")
    return field != nil
}

func (s *DepartmentDataScope) Apply(principal *security.Principal, query orm.SelectQuery) error {
    // 从 principal.Details 中获取用户的部门 ID
    type UserDetails struct {
        DepartmentId string `json:"departmentId"`
    }
    
    details, ok := principal.Details.(UserDetails)
    if !ok {
        return nil // 如果没有部门信息，不应用过滤
    }
    
    // 应用过滤条件
    query.Where(func(cb orm.ConditionBuilder) {
        cb.Equals("department_id", details.DepartmentId)
    })
    
    return nil
}
```

然后在 `RolePermissionsLoader` 中使用自定义数据范围：

```go
func (l *MyRolePermissionsLoader) LoadPermissions(ctx context.Context, role string) (map[string]security.DataScope, error) {
    result := make(map[string]security.DataScope)
    
    result["sys.user.view"] = NewDepartmentDataScope() // 只能查看本部门用户
    
    return result, nil
}
```

#### 完全自定义数据权限解析

如果需要实现完全自定义的数据权限解析逻辑（非 RBAC），可以实现 `security.DataPermissionResolver` 接口并替换框架的实现：

```go
type MyCustomDataPermResolver struct {
    // 自定义字段
}

func (r *MyCustomDataPermResolver) ResolveDataScope(ctx context.Context, principal *security.Principal, permToken string) (security.DataScope, error) {
    // 自定义数据权限解析逻辑
    // ...
    return security.NewAllDataScope(), nil
}

func NewMyCustomDataPermResolver() security.DataPermissionResolver {
    return &MyCustomDataPermResolver{}
}

// 在 main.go 中替换框架的实现
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

## 配置说明

### 配置文件

将 `application.toml` 放在 `./configs/` 或 `./` 目录，或通过 `VEF_CONFIG_PATH` 环境变量指定路径。

**完整配置示例：**

```toml
[vef.app]
name = "my-app"          # 应用名称
port = 8080              # HTTP 端口
body_limit = "10MB"      # 请求体大小限制

[vef.datasource]
type = "postgres"        # 数据库类型：postgres、mysql、sqlite
host = "localhost"
port = 5432
user = "postgres"
password = "password"
database = "mydb"
schema = "public"        # PostgreSQL schema
# path = "./data.db"    # SQLite 数据库文件路径

[vef.security]
token_expires = "2h"     # Jwt token 过期时间

[vef.storage]
provider = "minio"       # 存储提供者：memory、filesystem、minio（默认：memory）

[vef.storage.minio]
endpoint = "localhost:9000"
access_key = "minioadmin"
secret_key = "minioadmin"
use_ssl = false
region = "us-east-1"
bucket = "mybucket"

[vef.storage.filesystem]
root = "./storage"       # 当 provider = "filesystem" 时的根目录

[vef.redis]
host = "localhost"
port = 6379
user = ""                # 可选
password = ""            # 可选
database = 0             # 0-15
network = "tcp"          # tcp 或 unix

[vef.cors]
enabled = true
allow_origins = ["*"]
```

### 环境变量

使用环境变量覆盖配置：

- `VEF_CONFIG_PATH` - 配置文件路径
- `VEF_LOG_LEVEL` - 日志级别（debug、info、warn、error）
- `VEF_NODE_ID` - XID 节点标识符，用于 ID 生成
- `VEF_I18N_LANGUAGE` - 语言设置（en、zh-CN）

## 高级功能

### 缓存

使用内存或 Redis 缓存：

```go
import (
    "github.com/ilxqx/vef-framework-go/cache"
    "time"
)

// 内存缓存
memCache := cache.NewMemory[models.User](
    cache.WithMemMaxSize(1000),
    cache.WithMemDefaultTtl(5 * time.Minute),
)

// Redis 缓存
redisCache := cache.NewRedis[models.User](
    redisClient,
    "users",
    cache.WithRdsDefaultTtl(10 * time.Minute),
)

// 使用方式
user, err := memCache.GetOrLoad(ctx, "user:123", func(ctx context.Context) (models.User, error) {
    // 缓存未命中时的回退加载器
    return loadUserFromDB(ctx, "123")
})
```

### 事件总线

发布和订阅事件：

```go
import "github.com/ilxqx/vef-framework-go/event"

// 发布事件
func (r *UserResource) CreateUser(ctx fiber.Ctx, bus event.Bus, ...) error {
    // 创建用户逻辑
    
    bus.Publish(event.NewBaseEvent(
        "user.created",
        event.WithSource("user-service"),
        event.WithMeta("userId", user.Id),
    ))
    
    return result.Ok().Response(ctx)
}

// 订阅事件
func main() {
    vef.Run(
        vef.Invoke(func(bus event.Bus, logger log.Logger) {
            unsubscribe := bus.Subscribe("user.created", func(ctx context.Context, e event.Event) {
                // 处理事件
                logger.Infof("用户已创建: %s", e.Meta()["userId"])
            })
            
            // 可选：稍后取消订阅
            _ = unsubscribe
        }),
    )
}
```

### 生命周期钩子

框架通过 `vef.Lifecycle` 提供生命周期管理，允许您注册在应用启动和关闭期间执行的钩子。这对于正确的资源清理至关重要，特别是对于事件订阅者。

#### 事件订阅者清理

注册事件订阅者时，应在关闭时清理订阅以防止资源泄漏：

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
            // 创建并注册审计事件订阅者
            auditSub := NewAuditEventSubscriber(db, subscriber)

            // 注册清理钩子
            lc.Append(vef.StopHook(func() {
                auditSub.Unsubscribe()  // 关闭时清理
            }))

            // 创建并注册登录事件订阅者
            loginSub := NewLoginEventSubscriber(db, subscriber)

            // 注册清理钩子
            lc.Append(vef.StopHook(func() {
                loginSub.Unsubscribe()  // 关闭时清理
            }))
        },
    ),
)
```

**关键模式：**

1. **存储取消订阅函数**：事件订阅者构造函数在调用 `bus.Subscribe()` 时应返回 `UnsubscribeFunc`
2. **注册停止钩子**：使用 `lc.Append(vef.StopHook(...))` 注册清理函数
3. **在钩子中调用取消订阅**：在关闭期间调用存储的 `Unsubscribe()` 函数

**事件订阅者实现示例：**

```go
type AuditEventSubscriber struct {
    db           orm.Db
    unsubscribe  event.UnsubscribeFunc
}

func NewAuditEventSubscriber(db orm.Db, subscriber event.Subscriber) *AuditEventSubscriber {
    sub := &AuditEventSubscriber{db: db}

    // 订阅并存储取消订阅函数
    sub.unsubscribe = subscriber.Subscribe("*.created", sub.handleAuditEvent)

    return sub
}

func (s *AuditEventSubscriber) handleAuditEvent(ctx context.Context, e event.Event) {
    // 处理审计日志
}

func (s *AuditEventSubscriber) Unsubscribe() {
    if s.unsubscribe != nil {
        s.unsubscribe()
    }
}
```

这种模式确保优雅关闭，不会出现资源泄漏或孤立订阅。

### 上下文助手

`contextx` 包提供实用函数，用于在依赖注入不可用时访问请求范围的资源。这些助手在自定义处理器、钩子或其他需要从 Fiber 上下文访问框架提供的资源的场景中很有用。

```go
import "github.com/ilxqx/vef-framework-go/contextx"

func (r *RoleResource) CustomMethod(ctx fiber.Ctx) error {
    // 获取请求范围的数据库（已预配置 operator）
    db := contextx.Db(ctx)

    // 获取当前认证用户
    principal := contextx.Principal(ctx)

    // 获取请求范围的日志记录器（包含请求 ID）
    logger := contextx.Logger(ctx)

    // 获取 mold 转换器用于数据清理
    transformer := contextx.Transformer(ctx)

    // 使用这些资源
    logger.Infof("用户 %s 正在执行自定义操作", principal.Id)

    var model models.SomeModel
    if err := db.NewSelect().Model(&model).Scan(ctx.Context()); err != nil {
        return err
    }

    return result.Ok(model).Response(ctx)
}
```

**可用助手：**

- **`contextx.Db(ctx)`** - 返回请求范围的 `orm.Db`，已预配置审计字段（如 `operator`）
- **`contextx.Principal(ctx)`** - 返回当前 `*security.Principal`（认证用户或匿名用户）
- **`contextx.Logger(ctx)`** - 返回请求范围的 `log.Logger`，包含请求 ID 用于关联
- **`contextx.Transformer(ctx)`** - 返回 `mold.Transformer` 用于数据转换和清理

**何时使用：**

- **使用 contextx 助手**：在无法使用参数注入的自定义处理器中，或在仅接收 `fiber.Ctx` 的实用函数中
- **优先使用参数注入**：在定义 API 处理器方法时，让框架直接注入依赖作为参数，以获得更好的可测试性和清晰度

**示例 - 使用两种模式：**

```go
// 优先使用：处理器中的参数注入
func (r *UserResource) UpdateProfile(
    ctx fiber.Ctx,
    db orm.Db,           // 由框架注入
    logger log.Logger,   // 由框架注入
    params ProfileParams,
) error {
    logger.Infof("正在更新配置文件")
    // ...
}

// 在注入不可用时使用 contextx
func helperFunction(ctx fiber.Ctx) error {
    db := contextx.Db(ctx)       // 从上下文提取
    logger := contextx.Logger(ctx)
    logger.Infof("助手函数")
    // ...
}
```

### 定时任务

框架基于 [gocron](https://github.com/go-co-op/gocron) 提供定时任务调度功能。

#### 基本用法

通过 DI 注入 `cron.Scheduler` 并创建任务：

```go
import (
    "context"
    "time"
    "github.com/ilxqx/vef-framework-go/cron"
)

func main() {
    vef.Run(
        vef.Invoke(func(scheduler cron.Scheduler) {
            // Cron 表达式任务（5 字段格式）
            scheduler.NewJob(
                cron.NewCronJob(
                    "0 0 * * *",  // 表达式：每天午夜执行
                    false,         // withSeconds: 使用 5 字段格式
                    cron.WithName("daily-cleanup"),
                    cron.WithTags("maintenance"),
                    cron.WithTask(func(ctx context.Context) {
                        // 任务逻辑
                    }),
                ),
            )
            
            // 固定间隔任务
            scheduler.NewJob(
                cron.NewDurationJob(
                    5*time.Minute,
                    cron.WithName("health-check"),
                    cron.WithTask(func() {
                        // 每 5 分钟执行一次
                    }),
                ),
            )
        }),
    )
}
```

#### 任务类型

框架支持多种任务调度方式：

**1. Cron 表达式任务**

```go
// 5 字段格式：分 时 日 月 周
scheduler.NewJob(
    cron.NewCronJob(
        "30 * * * *",  // 每小时的第 30 分钟执行
        false,          // 不包含秒字段
        cron.WithName("hourly-report"),
        cron.WithTask(func() {
            // 生成报表
        }),
    ),
)

// 6 字段格式：秒 分 时 日 月 周
scheduler.NewJob(
    cron.NewCronJob(
        "0 30 * * * *",  // 每小时的第 30 分 0 秒执行
        true,             // 包含秒字段
        cron.WithName("precise-task"),
        cron.WithTask(func() {
            // 精确到秒的任务
        }),
    ),
)
```

**2. 固定间隔任务**

```go
scheduler.NewJob(
    cron.NewDurationJob(
        10*time.Second,
        cron.WithName("metrics-collector"),
        cron.WithTask(func() {
            // 每 10 秒收集一次指标
        }),
    ),
)
```

**3. 随机间隔任务**

```go
scheduler.NewJob(
    cron.NewDurationRandomJob(
        1*time.Minute,  // 最小间隔
        5*time.Minute,  // 最大间隔
        cron.WithName("random-check"),
        cron.WithTask(func() {
            // 在 1-5 分钟随机间隔执行
        }),
    ),
)
```

**4. 一次性任务**

```go
// 立即执行一次
scheduler.NewJob(
    cron.NewOneTimeJob(
        []time.Time{},  // 空切片表示立即执行
        cron.WithName("init-task"),
        cron.WithTask(func() {
            // 初始化任务
        }),
    ),
)

// 在指定时间执行一次
scheduler.NewJob(
    cron.NewOneTimeJob(
        []time.Time{time.Now().Add(1 * time.Hour)},
        cron.WithName("delayed-task"),
        cron.WithTask(func() {
            // 1 小时后执行
        }),
    ),
)

// 在多个指定时间执行
scheduler.NewJob(
    cron.NewOneTimeJob(
        []time.Time{
            time.Date(2024, 12, 31, 23, 59, 0, 0, time.Local),
            time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local),
        },
        cron.WithName("new-year-task"),
        cron.WithTask(func() {
            // 在指定时间点执行
        }),
    ),
)
```

#### 任务配置选项

```go
scheduler.NewJob(
    cron.NewDurationJob(
        1*time.Hour,
        // 任务名称（必需）
        cron.WithName("backup-task"),
        
        // 标签（用于分组和批量操作）
        cron.WithTags("backup", "critical"),
        
        // 任务处理函数（必需）
        cron.WithTask(func(ctx context.Context) {
            // 如果函数接受 context.Context 参数，框架会自动注入
            // 支持优雅关闭和超时控制
        }),
        
        // 允许并发执行（默认为单例模式）
        cron.WithConcurrent(),
        
        // 设置开始时间
        cron.WithStartAt(time.Now().Add(10 * time.Minute)),
        
        // 立即开始执行
        cron.WithStartImmediately(),
        
        // 设置停止时间
        cron.WithStopAt(time.Now().Add(24 * time.Hour)),
        
        // 限制执行次数
        cron.WithLimitedRuns(100),
        
        // 自定义上下文
        cron.WithContext(context.Background()),
    ),
)
```

#### 任务管理

```go
vef.Invoke(func(scheduler cron.Scheduler) {
    // 创建任务
    job, _ := scheduler.NewJob(
        cron.NewDurationJob(
            1*time.Minute,
            cron.WithName("my-task"),
            cron.WithTags("tag1", "tag2"),
            cron.WithTask(func() {}),
        ),
    )
    
    // 获取所有任务
    allJobs := scheduler.Jobs()
    
    // 按标签删除任务
    scheduler.RemoveByTags("tag1", "tag2")
    
    // 按 ID 删除任务
    scheduler.RemoveJob(job.Id())
    
    // 更新任务定义
    scheduler.Update(job.Id(), cron.NewDurationJob(
        2*time.Minute,
        cron.WithName("my-task-updated"),
        cron.WithTask(func() {}),
    ))
    
    // 立即运行任务（不影响调度）
    job.RunNow()
    
    // 查看下次运行时间
    nextRun, _ := job.NextRun()
    
    // 查看最后运行时间
    lastRun, _ := job.LastRun()
    
    // 停止所有任务
    scheduler.StopJobs()
})
```

### 文件存储

框架内置了文件存储功能，支持 MinIO、文件系统以及内存存储。

#### 内置存储资源

框架自动注册了 `sys/storage` 资源，提供以下 Api 端点：

| Action | 说明 |
|--------|------|
| `upload` | 上传文件（自动生成唯一文件名） |
| `get_presigned_url` | 获取预签名 URL（用于直接访问或上传） |
| `delete_temp` | 删除临时文件（仅允许 `temp/` 前缀） |
| `stat` | 获取文件元数据 |
| `list` | 列出文件 |

**上传文件示例：**

```bash
# 使用内置的 upload Api
curl -X POST http://localhost:8080/api \
  -H "Authorization: Bearer <token>" \
  -F "resource=sys/storage" \
  -F "action=upload" \
  -F "version=v1" \
  -F "params[file]=@/path/to/file.jpg" \
  -F "params[contentType]=image/jpeg" \
  -F "params[metadata][key1]=value1"
```

**上传响应：**

```json
{
  "code": 0,
  "message": "成功",
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

#### 文件密钥规则

框架对上传文件使用以下命名规则：

- **临时文件**：`temp/YYYY/MM/DD/{uuid}{extension}`
  - 例如：`temp/2025/01/15/550e8400-e29b-41d4-a716-446655440000.jpg`
  - 原始文件名保存在元数据 `Original-Filename` 中

- **永久文件**：通过 `PromoteObject` 提升临时文件
  - 从临时路径移除 `temp/` 前缀
  - 例如：`temp/2025/01/15/xxx.jpg` → `2025/01/15/xxx.jpg`

#### 自定义文件上传

在自定义资源中注入 `storage.Service` 实现文件上传：

```go
import (
    "mime/multipart"

    "github.com/gofiber/fiber/v3"
    "github.com/ilxqx/vef-framework-go/api"
    "github.com/ilxqx/vef-framework-go/result"
    "github.com/ilxqx/vef-framework-go/storage"
)

// 定义上传参数结构
type UploadAvatarParams struct {
    api.P

    File *multipart.FileHeader `json:"file"`
}

func (r *UserResource) UploadAvatar(
    ctx fiber.Ctx,
    service storage.Service,
    params UploadAvatarParams,
) error {
    // 检查文件是否存在
    if params.File == nil {
        return result.Err("文件不能为空")
    }

    // 打开上传的文件
    reader, err := params.File.Open()
    if err != nil {
        return err
    }
    defer reader.Close()

    // 自定义文件路径
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

#### 临时文件提升

使用 `PromoteObject` 将临时上传的文件转为永久文件：

```go
// 业务逻辑确认后，提升临时文件
info, err := provider.PromoteObject(ctx.Context(), "temp/2025/01/15/xxx.jpg")
// info.Key 变为: "2025/01/15/xxx.jpg"
```

#### 配置存储提供者

将 `vef.storage.provider` 设置为 `minio`、`filesystem` 或 `memory`（默认）并在 `application.toml` 中进行配置：

```toml
[vef.storage]
provider = "minio"  # 选项：minio、filesystem、memory

[vef.storage.minio]
endpoint = "localhost:9000"
access_key = "minioadmin"
secret_key = "minioadmin"
use_ssl = false
region = "us-east-1"
bucket = "mybucket"

[vef.storage.filesystem]
root = "./storage"       # 当 provider = "filesystem" 时的根目录
```

### 数据验证

使用 [go-playground/validator](https://github.com/go-playground/validator) 标签：

```go
type UserParams struct {
    Username string `validate:"required,alphanum,min=3,max=32" label:"用户名"`
    Email    string `validate:"required,email" label:"邮箱"`
    Age      int    `validate:"min=18,max=120" label:"年龄"`
    Website  string `validate:"omitempty,url" label:"网站"`
    Password string `validate:"required,min=8,containsany=!@#$%^&*" label:"密码"`
}
```

**常用规则：**

| 规则 | 说明 |
|------|------|
| `required` | 必填字段 |
| `omitempty` | 可选字段（值为空时跳过验证） |
| `min` | 最小值（数字）或最小长度（字符串） |
| `max` | 最大值（数字）或最大长度（字符串） |
| `len` | 精确长度 |
| `eq` | 等于 |
| `ne` | 不等于 |
| `gt` | 大于 |
| `gte` | 大于等于 |
| `lt` | 小于 |
| `lte` | 小于等于 |
| `alpha` | 仅字母 |
| `alphanum` | 字母和数字 |
| `ascii` | ASCII 字符 |
| `numeric` | 数字字符串 |
| `email` | 邮箱地址 |
| `url` | URL 网址 |
| `uuid` | UUID 格式 |
| `ip` | IP 地址 |
| `json` | JSON 格式 |
| `contains` | 包含指定子串 |
| `startswith` | 以指定字符串开头 |
| `endswith` | 以指定字符串结尾 |

### CLI 工具

VEF Framework 提供 `vef-cli` 命令行工具用于代码生成和项目脚手架任务。

#### 生成构建信息

`generate-build-info` 命令创建包含版本、提交哈希和构建时间戳的构建信息文件：

```bash
go run github.com/ilxqx/vef-framework-go/cmd/vef-cli@latest generate-build-info -o internal/vef/build_info.go -p vef
```

**选项：**
- `-o, --output` - 输出文件路径（默认：`build_info.go`）
- `-p, --package` - 包名（默认：当前目录名）

**在 go:generate 中使用：**

```go
//go:generate go run github.com/ilxqx/vef-framework-go/cmd/vef-cli@latest generate-build-info -o internal/vef/build_info.go -p vef
```

生成的文件提供与监控模块兼容的 `BuildInfo` 变量：

```go
package vef

import "github.com/ilxqx/vef-framework-go/monitor"

// BuildInfo 指向构建元数据，用于 monitor 模块
var BuildInfo = &monitor.BuildInfo{
    AppVersion: "v1.0.0",               // 来自 git tags（或 "dev"）
    BuildTime:  "2025-01-15T10:30:00Z", // 构建时间戳
    GitCommit:  "abc123...",            // Git 提交 SHA
}
```

**生成的字段：**
- **Version**：从 git tags 提取（例如 `v1.0.0`）。如果没有 tags 则回退到 `"dev"`。
- **Commit**：当前 HEAD 的完整 git 提交 SHA。
- **BuildTime**：文件生成时的 UTC 时间戳。

#### 生成模型 Schema

`generate-model-schema` 命令为模型生成类型安全的字段访问器函数：

```bash
go run github.com/ilxqx/vef-framework-go/cmd/vef-cli@latest generate-model-schema -i ./models -o ./schemas -p schemas
```

**选项：**
- `-i, --input` - 包含模型文件的输入目录（必需）
- `-o, --output` - 生成 schema 文件的输出目录（必需）
- `-p, --package` - 生成文件的包名（必需）

**在 go:generate 中使用：**

```go
//go:generate go run github.com/ilxqx/vef-framework-go/cmd/vef-cli@latest generate-model-schema -i ./models -o ./schemas -p schemas
```

生成的 schema 提供类型安全的字段访问器：

```go
package schemas

var User = struct {
    Id        func(withTablePrefix ...bool) string
    Username  func(withTablePrefix ...bool) string
    Email     func(withTablePrefix ...bool) string
    CreatedAt func(withTablePrefix ...bool) string
    // ... 其他字段
}{
    Id:        field("id", "su"),
    Username:  field("username", "su"),
    Email:     field("email", "su"),
    CreatedAt: field("created_at", "su"),
}
```

**在查询中使用：**

```go
import "my-app/internal/sys/schemas"

// 类型安全的列引用
db.NewSelect().
    Model(&users).
    Where(func(cb orm.ConditionBuilder) {
        cb.Equals(schemas.User.Username(), "admin")
        cb.IsNotNull(schemas.User.Email())
    }).
    OrderBy(schemas.User.CreatedAt(true) + " DESC"). // 带表前缀
    Scan(ctx)
```

**优势：**
- **类型安全**：在编译时捕获拼写错误
- **IDE 自动完成**：字段名可被发现
- **重构支持**：重命名字段会更新所有引用
- **表前缀处理**：可选地在列名中包含表别名

关于 AI 辅助开发指南，请参阅 `cmd/CMD_DEV_GUIDELINES.md`。

## 最佳实践

### 项目结构

```
my-app/
├── cmd/
│   └── main.go                 # 应用入口
├── configs/
│   └── application.toml        # 配置文件
├── internal/
│   ├── models/                 # 数据模型
│   │   ├── user.go
│   │   └── order.go
│   ├── payloads/               # Api 参数
│   │   ├── user.go
│   │   └── order.go
│   ├── resources/              # Api 资源
│   │   ├── user.go
│   │   └── order.go
│   └── services/               # 业务服务
│       ├── user_service.go
│       └── email_service.go
└── go.mod
```

### 命名约定

- **模型：** 单数大驼峰（如 `User`、`Order`）
- **资源：** 小写斜杠分隔（如 `sys/user`、`shop/order`）
- **参数：** `XxxParams`（创建/更新）、`XxxSearch`（查询）
- **Action：** 小写下划线分隔（如 `find_page`、`create_user`）

### 错误处理

使用框架的 Result 类型实现一致的错误响应：

```go
import "github.com/ilxqx/vef-framework-go/result"

// 成功
return result.Ok(data).Response(ctx)

// 错误
return result.Err("操作失败")
return result.ErrWithCode(result.ErrCodeBadRequest, "参数无效")
return result.Errf("用户 %s 不存在", username)
```

### 日志记录

注入日志记录器并使用：

```go
func (r *UserResource) Handler(
    ctx fiber.Ctx,
    logger log.Logger,
) error {
    logger.Infof("处理来自 %s 的请求", ctx.IP())
    logger.Warnf("检测到异常活动")
    logger.Errorf("操作失败: %v", err)
    
    return nil
}
```

## 文档与资源

- [Fiber Web Framework](https://gofiber.io/) - 底层 HTTP 框架
- [Bun ORM](https://bun.uptrace.dev/) - 数据库 ORM
- [Go Playground Validator](https://github.com/go-playground/validator) - 数据验证
- [Uber FX](https://uber-go.github.io/fx/) - 依赖注入

## 许可证

本项目采用 [Apache License 2.0](LICENSE) 许可。
