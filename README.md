# VEF Framework Go

VEF Framework Go 是一个现代化的 Go Web 开发框架，基于依赖注入和模块化设计，提供开箱即用的 CRUD API、ORM、认证、缓存、事件系统等企业级功能。

## 🚀 核心特性

- **开箱即用的 CRUD API**: 预置一套略带偏见的 CRUD API，快速完成增删改查接口的开发
- **强类型 ORM**: 基于 Bun ORM 的类型安全数据库操作
- **多策略认证体系**: 内置支持 JWT、OpenAPI 和基于密码的认证，具有可扩展的认证器架构
- **灵活的缓存系统**: 支持本地和 Redis 缓存
- **异步事件系统**: 发布订阅模式的事件处理
- **定时任务调度**: Cron 表达式支持的任务系统
- **模块化架构**: 基于 Uber FX 的依赖注入
- **工具函数库**: ID生成、数据转换、密码处理等实用工具

## 📦 快速开始

### 1. 安装和初始化

```bash
# 创建新项目
mkdir my-app && cd my-app
go mod init my-app

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
type = "postgres"  # 或 "sqlite"
host = "localhost"
port = 5432
user = "postgres"
password = "password"
database = "mydb"
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

### 创建 API 资源

框架预置了完整的 CRUD API，支持泛型和类型安全：

```go
// payloads/user.go
package payloads

import (
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/null"
	"github.com/ilxqx/vef-framework-go/orm"
)

// UserSearch 用户搜索参数
type UserSearch struct {
	api.Params
	Keyword string `json:"keyword"` // 关键词
}

// UserParams 用户新增/修改参数
type UserParams struct {
	api.Params
	orm.ModelPK `json:",inline"`

	Username    string      `json:"username" validate:"required,alphanum,max=32" label:"用户账号"`    // 用户账号
	Password    string      `json:"password" validate:"required,min=6,max=128" label:"用户密码"`      // 用户密码
	Name        string      `json:"name" validate:"required,max=16" label:"用户名称"`                 // 用户名称
	IsActive    bool        `json:"isActive"`                                                     // 是否启用
	IsLocked    bool        `json:"isLocked"`                                                     // 是否锁定
	Email       null.String `json:"email" validate:"omitempty,email,max=64" label:"邮箱"`           // 邮箱
	Remark      null.String `json:"remark" validate:"omitempty,max=256" label:"备注"`               // 备注
}
```

```go
// resources/user.go
package resources

import (
	"vef-app-server-starter-go/internal/sys/models"
	"vef-app-server-starter-go/internal/sys/payloads"

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
		Resource: api.NewResource(
			"sys/user",
			api.WithAPIs(
				api.Config{
					Action: "findAll",
					Public: true,
				},
				api.Config{
					Action: "create",
					Public: true,
				},
				api.Config{
					Action: "update",
					Public: true,
				},
				api.Config{
					Action: "delete",
					Public: true,
				},
			),
		),
		FindAllAPI: apis.NewFindAllAPI[models.User, payloads.UserSearch](),
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
    "my-app/resources"
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

## 🔗 相关资源

- [Fiber Web Framework](https://gofiber.io/) - 底层 HTTP 框架
- [Bun ORM](https://bun.uptrace.dev/) - 数据库 ORM
- [Go Playground Validator](https://github.com/go-playground/validator) - 数据验证

---

**VEF Framework Go** - 让企业级 Go Web 开发更简单高效！