# VEF Framework Go

VEF Framework Go 是一个现代化的 Go Web 开发框架，基于依赖注入和模块化设计，提供开箱即用的 CRUD API、ORM、认证、缓存、事件系统等企业级功能。

## 🚀 核心特性

- **开箱即用的 CRUD API**: 预置一套略带偏见的 CRUD API，快速完成增删改查接口的开发
- **强类型 ORM**: 类型安全数据库操作
- **多策略认证体系**: 内置支持 JWT、OpenAPI 和基于密码的认证，具有可扩展的认证架构
- **灵活的缓存系统**: 支持本地和 Redis 缓存
- **异步事件系统**: 发布订阅模式的事件处理
- **定时任务调度**: Cron表达式支持的任务系统
- **模块化架构**: 依赖注入和模块化设计

## 📦 快速开始

### 1. 安装和初始化

```bash
# 创建新项目
mkdir myapp && cd myapp
go mod init myapp

# 安装框架
go get -u github.com/ilxqx/vef-framework-go
```

### 2. 基础配置

创建 `application.toml` 配置文件：

```toml
[vef.app]
name = "my-app"
port = 8080

[vef.security]
token_expires = "2h"

[vef.datasource]
type = "postgres"  # 目前支持 postgres、mysql、sqlite
host = "localhost"
port = 5432
user = "postgres"
password = "password"
database = "postgres"
schema = "public"
```

### 3. 创建主程序

```go
// main.go
package main

import "github.com/ilxqx/vef-framework-go"

func main() {
    vef.Run()
}
```

这样就完成了一个最基础的 Web 服务器，监听在 8080 端口。

## 🏗️ 项目结构建议

```
my-app/
├── cmd/                 
│   └── main.go          # 应用入口
├── config/              
│   └── application.toml # 配置文件
├── internal/
│   └── models/          # 数据模型定义
│       └── user.go
│   └── payloads/        # API参数定义
│       └── user.go
│   └── resources/       # API资源定义
│       └── user.go
│   └── services/        # 业务共享逻辑定义
        └── user.go
```

## 📊 数据模型

### 定义模型

所有模型都应继承 `orm.Model`，它提供了基础的审计字段：

```go
// models/user.go
package models

import (
	"github.com/ilxqx/vef-framework-go/null"
	"github.com/ilxqx/vef-framework-go/orm"
)

type User struct {
	orm.BaseModel `bun:"table:sys_user,alias:su"`
	orm.Model     `bun:"extend"`

	Username          string        `json:"username" validate:"required,alphanum,max=32" label:"用户账号"` // 用户账号
	Password          string        `json:"-" validate:"required,min=6,max=128" label:"用户密码"`          // 用户密码
	Name              string        `json:"name" validate:"required,max=16" label:"用户名称"`              // 用户名称
	IsActive          null.Bool     `json:"isActive"`                                                  // 是否启用
	IsLocked          null.Bool     `json:"isLocked"`                                                  // 是否锁定
	Email             null.String   `json:"email" validate:"omitempty,email,max=64" label:"邮箱"`        // 邮箱
	Remark            null.String   `json:"remark" validate:"omitempty,max=256" label:"备注"`            // 备注
}
```

### 模型字段标签

具体请参考 [Bun ORM](https://bun.uptrace.dev/guide/models.html) 的文档。

## 🔌 CRUD API

### 1. 定义参数结构

框架预置了完整的 CRUD API，支持泛型和类型安全。首先需要定义参数结构体：

```go
// payloads/user.go
package payloads

import (
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/null"
	"github.com/ilxqx/vef-framework-go/orm"
)

// UserSearch 用户搜索参数
// 嵌入 api.In 来标识这是一个API参数结构体
type UserSearch struct {
	api.In
	Keyword string `json:"keyword" search:"contains,column=username|name|email"` // 关键词搜索
}

// UserParams 用户新增/修改参数
type UserParams struct {
	api.In
	orm.ModelPK `json:",inline"` // 嵌入主键字段（用于更新操作）

	Username string      `json:"username" validate:"required,alphanum,max=32" label:"用户账号"`    // 用户账号
	Password string      `json:"password" validate:"required,min=6,max=128" label:"用户密码"`      // 用户密码
	Name     string      `json:"name" validate:"required,max=16" label:"用户名称"`                 // 用户名称
	IsActive bool        `json:"isActive"`                                                     // 是否启用
	IsLocked bool        `json:"isLocked"`                                                     // 是否锁定
	Email    null.String `json:"email" validate:"omitempty,email,max=64" label:"邮箱"`           // 邮箱
	Remark   null.String `json:"remark" validate:"omitempty,max=256" label:"备注"`               // 备注
}
```

#### 参数验证规则

框架使用 `validate` 标签进行参数验证，支持以下规则：

- `required`: 必填字段
- `omitempty`: 空值时跳过验证
- `min=6,max=128`: 字符串长度限制
- `email`: 邮箱格式验证
- `alphanum`: 仅允许字母和数字
- `label`: 错误信息中显示的字段名称

更多验证规则请参考 [Go Playground Validator](https://github.com/go-playground/validator) 文档。

### 2. 创建 API 资源

使用框架预置的 CRUD API 创建资源：

```go
// resources/user.go
package resources

import (
	"myapp/internal/sys/models"
	"myapp/internal/sys/payloads"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/apis"
)

type userResource struct {
	api.Resource
	*apis.FindAllAPI[models.User, payloads.UserSearch]
	*apis.FindPageAPI[models.User, payloads.UserSearch]
	*apis.CreateAPI[models.User, payloads.UserParams]
	*apis.UpdateAPI[models.User, payloads.UserParams]
	*apis.DeleteAPI[models.User]
}

func NewUserResource() api.Resource {
	return &userResource{
		Resource: api.NewResource("sys/user"),
		FindAllAPI: apis.NewFindAllAPI[models.User, payloads.UserSearch]().Public(),
		FindPageAPI: apis.NewFindPageAPI[models.User, payloads.UserSearch]().Public(),
		FindOneAPI: apis.NewFindOneAPI[models.User, payloads.UserSearch]().Public(),
		CreateAPI:  apis.NewCreateAPI[models.User, payloads.UserParams]().Public(),
		UpdateAPI:  apis.NewUpdateAPI[models.User, payloads.UserParams]().Public(),
		DeleteAPI:  apis.NewDeleteAPI[models.User]().Public(),
	}
}
```

### 3. 高级功能

#### Pre/Post 处理器

框架支持在 CRUD 操作前后执行自定义逻辑：

```go
// 创建用户资源时的业务逻辑处理
func NewUserResource() api.Resource {
	return &userResource{
		Resource: api.NewResource("sys/user"),
		CreateAPI: apis.NewCreateAPI[models.User, payloads.UserParams]().
			PreCreate(func(model *models.User, params *payloads.UserParams, ctx fiber.Ctx, db orm.Db) error {
				// 创建前的业务逻辑：密码加密
				hashed, err := security.HashPassword(params.Password)
				if err != nil {
					return err
				}
				model.Password = hashed
				return nil
			}).
			PostCreate(func(model *models.User, params *payloads.UserParams, ctx fiber.Ctx, tx orm.Db) error {
				// 创建后的业务逻辑：发送欢迎邮件
				return sendWelcomeEmail(model.Email)
			}),
		UpdateAPI: apis.NewUpdateAPI[models.User, payloads.UserParams]().
			PreUpdate(func(oldModel, newModel *models.User, params *payloads.UserParams, ctx fiber.Ctx, db orm.Db) error {
				// 更新前的业务逻辑：检查权限
				if oldModel.IsLocked && !hasAdminPermission(ctx) {
					return result.ErrWithCode(result.ErrCodeForbidden, "无法修改已锁定的用户")
				}
				return nil
			}),
		DeleteAPI: apis.NewDeleteAPI[models.User]().
			PreDelete(func(model *models.User, ctx fiber.Ctx, db orm.Db) (err error) {
				// 删除前的业务逻辑：检查依赖关系
				var count int64
				if count, err = db.NewSelect().Model((*models.Order)(nil)).
					Where(func(cb orm.ConditionBuilder) {
						cb.Equals("user_id", model.Id)
					}).Count(ctx); err != nil {
					return
				}
			
				if count > 0 {
					return result.Err("用户存在关联订单，无法删除")
				}
				return nil
			}),
	}
}
```

#### 查询定制

FindAPI 支持自定义查询逻辑：

```go
// 自定义用户查询
FindAllAPI: apis.NewFindAllAPI[models.User, payloads.UserSearch]().
	// 自定义查询条件
	QueryApplier(func(query orm.SelectQuery, search payloads.UserSearch, ctx fiber.Ctx) {
		query.Where(func(cb orm.ConditionBuilder) {
			cb.IsTrue("is_active")
		})
	}).
	// 包含关联关系
	// Relations(...).
	// 结果后处理
	Processor(func(users []models.User, search payloads.UserSearch, ctx fiber.Ctx) any {
		// 可以对结果进行转换或过滤
		return users
	}),
```

#### 服务注入

框架支持在 API 资源中注入服务，并自动传递给处理器：

```go
// services/user_service.go
type UserService struct {
	logger log.Logger
	db     orm.Db
}

func (s *UserService) WithLogger(logger log.Logger) *UserService {
	// 框架会自动调用此方法注入 Logger
	return &UserService{
		logger: logger,
		db:     s.db,
	}
}

// 业务方法
func (s *UserService) ValidateUser(user *models.User) error {
	s.logger.Infof("Validating user: %s", user.Username)
	// 业务逻辑...
	return nil
}

// resources/user.go
type userResource struct {
	api.Resource
	*apis.FindAllAPI[models.User, payloads.UserSearch]
	*apis.CreateAPI[models.User, payloads.UserParams]
	*apis.UpdateAPI[models.User, payloads.UserParams]
	*apis.DeleteAPI[models.User]
	
	// 注入的服务
	UserService *UserService
}

// 自定义处理器可以接收注入的服务
func (r *userResource) ValidateUser(ctx fiber.Ctx, userService *UserService, params ValidateUserParams) error {
	// userService 会被自动注入
	user := &models.User{Username: params.Username}
	return userService.ValidateUser(user)
}
```

#### 参数验证规则

框架使用 `validate` 标签进行参数验证，支持以下规则：

- `required`: 必填字段
- `omitempty`: 空值时跳过验证
- `min=6,max=128`: 字符串长度限制
- `email`: 邮箱格式验证
- `alphanum`: 仅允许字母和数字
- `label`: 错误信息中显示的字段名称

更多验证规则请参考 [Go Playground Validator](https://github.com/go-playground/validator) 文档。

### 2. 创建 API 资源

```go
// resources/user.go
package resources

import (
	"myapp/internal/sys/models"
	"myapp/internal/sys/payloads"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/apis"
)

type userResource struct {
	api.Resource
	*apis.FindAllAPI[models.User, payloads.UserSearch]
	*apis.CreateAPI[models.User, payloads.UserParams]
	*apis.UpdateAPI[models.User, payloads.UserParams]
	*apis.DeleteAPI[models.User]
}

func NewUserResource() api.Resource {
	return &userResource{
		Resource: api.NewResource("sys/user"),
		FindAllAPI: apis.NewFindAllAPI[models.User, payloads.UserSearch](),
		FindPageAPI: apis.NewFindPageAPI[models.User, payloads.UserSearch](),
		FindOneAPI: apis.NewFindOneAPI[models.User, payloads.UserSearch](),
		CreateAPI:  apis.NewCreateAPI[models.User, payloads.UserParams](),
		UpdateAPI:  apis.NewUpdateAPI[models.User, payloads.UserParams](),
		DeleteAPI:  apis.NewDeleteAPI[models.User](),
	}
}
```

### 注册资源

在 `main.go` 中注册 API 资源：

```go
package main

import (
    "github.com/ilxqx/vef-framework-go"
    "myapp/resources"
)

func main() {
    vef.Run(
        vef.ProvideAPIResource(resources.NewUserResource),
        // 可以注册多个资源
        // vef.ProvideAPIResource(resources.NewOrderResource),
    )
}
```

### API 请求规范

整个应用只存在一个端点 `POST /api`，请求体为 JSON 格式，请求体规范：

```json
{
    "resource": "sys/user",
    "action": "findAll",
    "version": "v1",
    "params": {
        "keyword": "test"
    },
    "meta": {}
}
```

`meta` 字段为可选字段，用于传递一些元数据，一般不会使用。

## 🗄️ 数据库操作

### 基础查询

框架基于 Bun ORM 提供类型安全的数据库操作：

```go
// 简单查询
func (r *userResource) GetUser(ctx fiber.Ctx, db orm.Db, params GetUserParams) error {
	var user models.User
	err := db.NewSelect().
		Model(&user).
		Where(func(cb orm.ConditionBuilder) {
			cb.Equals("id", params.Id)
		}).
		Scan(ctx)
	if err != nil {
		return err
	}
	return result.Ok(user).Response(ctx)
}

// 复杂查询示例
func (r *userResource) SearchUsers(ctx fiber.Ctx, db orm.Db, params UserSearchParams) error {
	var users []models.User
	query := db.NewSelect().Model(&users)
	
	// 动态条件构建
	if params.Name != constants.Empty {
		query.Where(func(cb orm.ConditionBuilder) {
			cb.Contains("name", params.Name)
		})
	}
	
	if params.IsActive != nil {
		query.Where(func(cb orm.ConditionBuilder) {
			cb.Equals("is_active", *params.IsActive)
		})
	}
	
	// 关联查询
	query.Relation("Profile").Relation("Orders")
	
	// 排序和分页
	query.OrderByDesc("created_at").Limit(params.Limit).Offset(params.Offset)
	
	if err := query.Scan(ctx); err != nil {
		return err
	}
	return result.Ok(users).Response(ctx)
}
```

### 事务操作

```go
func (r *userResource) TransferUser(ctx fiber.Ctx, db orm.Db, params TransferParams) error {
	return db.RunInTx(ctx, func(txCtx context.Context, tx orm.Db) error {
		// 查询源用户
		var fromUser models.User
		err := tx.NewSelect().Model(&fromUser).
			Where(func(cb orm.ConditionBuilder) {
				cb.Equals("id", params.FromUserId)
			}).Scan(txCtx)
		if err != nil {
			return err
		}
		
		// 更新用户状态
		_, err = tx.NewUpdate().Model(&fromUser).
			Set("status", "transferred").
			WherePK().Exec(txCtx)
		if err != nil {
			return err
		}
		
		// 创建转移记录
		transferRecord := &models.TransferRecord{
			FromUserId: params.FromUserId,
			ToUserId:   params.ToUserId,
			Reason:     params.Reason,
		}
		_, err = tx.NewInsert().Model(transferRecord).Exec(txCtx)
		return err
	})
}
```

### 条件构建器

```go
// 使用条件构建器进行复杂查询
query := db.NewSelect().Model(&users).
	Where(func(cb orm.ConditionBuilder) {
		// AND 条件组合
		cb.Equals("department_id", departmentId)
		cb.GreaterThan("salary", 5000)
		
		// OR 条件组合
		cb.Or(
			func(cb orm.ConditionBuilder) {
				cb.Equals("level", "senior")
			},
			func(cb orm.ConditionBuilder) {
				cb.Equals("level", "expert")
			},
		)
		
		// IN 查询
		cb.In("status", []string{"active", "pending"})
		
		// 模糊查询
		cb.Contains("email", "@company.com")
		cb.StartsWith("name", "张")
		cb.EndsWith("phone", "1234")
		
		// 空值检查
		cb.IsNotNull("avatar")
		cb.IsNull("deleted_at")
		
		// 日期范围
		cb.Between("created_at", startDate, endDate)
	})
```

## 🔗 依赖注入

### 注册 API 资源

在 `main.go` 中注册 API 资源到依赖注入容器：

```go
package main

import (
    "github.com/ilxqx/vef-framework-go"
    "my-app/resources"
    "my-app/services"
)

func main() {
    vef.Run(
        // 注册 API 资源
        vef.ProvideAPIResource(resources.NewUserResource),
        vef.ProvideAPIResource(resources.NewOrderResource),
        vef.ProvideAPIResource(resources.NewProductResource),
        
        // 注册服务
        fx.Provide(services.NewUserService),
        fx.Provide(services.NewEmailService),
        
        // 注册中间件
        vef.ProvideMiddleware(middleware.NewAuthMiddleware),
    )
}
```

### 可注入的内置类型

框架内置支持以下类型的自动注入：

```go
// API 处理器可以接收这些内置类型
func (r *userResource) MyHandler(
	ctx fiber.Ctx,                    // HTTP 上下文
	db orm.Db,                        // 数据库连接
	logger log.Logger,                // 日志记录器
	transformer trans.Transformer,    // 数据转换器
	principal *security.Principal,    // 当前用户信息（需要认证）
	params MyParams,                  // 请求参数（嵌入 api.In）
) error {
	logger.Infof("Processing request for user: %s", principal.Name)
	// 处理逻辑...
	return result.Ok("success").Response(ctx)
}
```

### 自定义参数解析器

可以注册自定义的参数解析器来注入特定类型：

```go
// 自定义解析器
type CustomServiceResolver struct{}

func (*CustomServiceResolver) Type() reflect.Type {
	return reflect.TypeFor[*services.CustomService]()
}

func (*CustomServiceResolver) Resolve(ctx fiber.Ctx) (reflect.Value, error) {
	// 从上下文中获取或创建服务实例
	service := getCustomServiceFromContext(ctx)
	return reflect.ValueOf(service), nil
}

// 在 main.go 中注册
func main() {
    vef.Run(
        fx.Provide(func() api.HandlerParamResolver {
            return &CustomServiceResolver{}
        }),
        // 其他配置...
    )
}
```

## 📜 API 调用规范

### 请求格式

整个应用只存在一个端点 `POST /api`，请求体为 JSON 格式：

```json
{
    "resource": "sys/user",     // 资源名称
    "action": "findAll",       // 操作名称
    "version": "v1",           // API 版本（可选，默认 v1）
    "params": {                 // 请求参数
        "keyword": "test",
        "pageSize": 20
    },
    "meta": {}                  // 元数据（可选）
}
```

### 响应格式

所有 API 响应都遵循统一的格式：

```json
{
    "code": 0,                  // 状态码（0 表示成功）
    "message": "成功",         // 状态信息
    "data": {                   // 响应数据
        // 具体数据内容
    }
}
```

### 分页响应

使用 `findPage` 动作时的响应格式：

```json
{
    "code": 0,
    "message": "成功",
    "data": {
        "content": [...],           // 数据列表
        "totalElements": 100,       // 总记录数
        "totalPages": 10,           // 总页数
        "page": 0,                  // 当前页码（从 0 开始）
        "size": 10,                 // 每页大小
        "first": true,              // 是否第一页
        "last": false               // 是否最后一页
    }
}
```

### CRUD 操作映射

| 操作     | 动作        | 说明           | 参数类型      |
|----------|-------------|----------------|---------------|
| 查询全部 | findAll     | 查询所有记录     | Search 类型    |
| 分页查询 | findPage    | 分页查询记录     | Search 类型    |
| 查询单个 | findOne     | 根据条件查询单个 | Search 类型    |
| 新增     | create      | 创建新记录       | Params 类型    |
| 修改     | update      | 更新记录         | Params 类型    |
| 删除     | delete      | 删除记录         | 包含 ID 参数 |

## 🛠️ 最佳实践

### 1. 项目结构建议

```
my-app/
├── cmd/
│   └── main.go              # 应用入口
├── config/
│   └── application.toml     # 配置文件
├── internal/
│   ├── models/              # 数据模型定义
│   │   ├── user.go
│   │   └── order.go
│   ├── payloads/            # API参数定义
│   │   ├── user.go
│   │   └── order.go
│   ├── resources/           # API资源定义
│   │   ├── user.go
│   │   └── order.go
│   └── services/            # 业务逻辑定义
│       ├── user_service.go
│       └── email_service.go
└── docs/                   # 文档
```

### 2. 命名约定

- **模型命名**: 使用单数大驼峰命名，如 `User`、`Order`
- **资源命名**: 使用斜杠分隔的小写名称，如 `sys/user`、`shop/order`
- **参数结构**: 使用复数大驼峰，如 `UserParams`、`UserSearch`
- **服务命名**: 使用 `Service` 后缀，如 `UserService`

### 3. 错误处理

```go
// 使用框架提供的错误类型
func (r *userResource) CreateUser(ctx fiber.Ctx, db orm.Db, params UserParams) error {
    // 参数验证错误会自动处理
    
    // 业务逻辑错误
    if existsUser(params.Email) {
        return result.ErrWithCode(result.ErrCodeBadRequest, "邮箱已存在")
    }
    
    // 数据库错误会自动转换
    // 成功响应
    return result.Ok(user.Id).Response(ctx)
}
```

### 4. 日志记录

```go
// 在处理器中使用日志
func (r *userResource) UpdateUser(
    ctx fiber.Ctx, 
    logger log.Logger, 
    db orm.Db, 
    params UserParams,
) error {
    logger.Infof("开始更新用户: %s", params.Id)
    
    // 业务逻辑...
    
    logger.Infof("用户更新成功: %s", params.Id)
    return result.Ok().Response(ctx)
}
```

## 🔗 相关资源

- [Fiber Web Framework](https://gofiber.io/) - 底层 HTTP 框架
- [Bun ORM](https://bun.uptrace.dev/) - 数据库 ORM
- [Go Playground Validator](https://github.com/go-playground/validator) - 数据验证

---

**VEF Framework Go** - 让企业级 Go Web 开发更简单高效！