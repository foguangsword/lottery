# Backend

Go 后端服务，"前端-后端-链上合约"架构中的后端层。

## 一、技术方案

### 1.1 技术选型与理由

| 功能 | 选型 | 理由 |
|------|------|------|
| Web 框架 | **Gin** | 最流行、轻量、学习成本极低，类似 Java 的 Spring Boot 但远比它轻 |
| ORM | **GORM** | 最流行、AutoMigrate 自动建表、中文文档丰富，类似 JPA |
| DB 驱动 | **go-sql-driver/mysql** | MySQL 官方 Go 驱动 |
| 链交互 | **go-ethereum (ethclient)** | 以太坊官方 Go 客户端，区块链领域事实标准 |
| 定时任务 | **robfig/cron v3** | 社区事实标准，支持 cron 表达式（类似 Linux crontab） |
| 配置管理 | **Viper + godotenv** | Viper 支持 YAML/JSON/环境变量，godotenv 开发时加载 `.env` 文件 |
| 日志 | **Zap** | Uber 出品，高性能结构化日志（JSON 格式），接近事实标准 |

**为什么没有选依赖注入框架（如 Wire/fx）？**
小项目手动 `new()` 更轻量、更直观，没有额外的学习成本和代码生成步骤。

### 1.2 分层架构

```
HTTP请求
   |
   v
Handler（控制器）--- 解析请求参数、调用 Service、返回统一响应
   |
   v
Service（业务层）--- 处理业务逻辑、调用 Repository 或 Blockchain
   |
   v
Repository（DAO）--- 数据库 CRUD 操作
```

**为什么要分三层？**
- **Handler** 只负责"接请求、调服务、回响应"，不写业务逻辑
- **Service** 只负责业务逻辑，不直接碰数据库
- **Repository** 只负责数据存取，不处理业务规则

这样每层职责单一，代码好维护。比如以后换数据库（MySQL 改 PostgreSQL），只改 Repository 层就行。

---

## 二、项目搭建全过程（从零开始）

### 2.1 初始化 Go 模块

```bash
cd backend
go mod init backend
```

**解释**：`go mod init 模块名` 创建 `go.mod` 文件，相当于 Java 的 `pom.xml` 或 `build.gradle`，用来管理依赖和模块路径。`backend` 是模块名，也就是代码里 `import "backend/xxx"` 的前缀。

### 2.2 创建目录结构

```bash
mkdir -p cmd/api
mkdir -p internal/{config,model,repository,service,handler,router,middleware,blockchain,scheduler}
mkdir -p pkg/{db,logger,response}
mkdir -p configs
```

**目录约定**：
- `cmd/api/` — 可执行程序的入口，每个独立程序放一个子目录（Go 惯例）
- `internal/` — 私有代码，不允许被外部模块导入（Go 编译器强制）
- `pkg/` — 公共代码包，可以被外部复用
- `configs/` — 配置文件

### 2.3 安装依赖

```bash
go get github.com/gin-gonic/gin
go get gorm.io/gorm gorm.io/driver/mysql
go get github.com/ethereum/go-ethereum
go get github.com/robfig/cron/v3
go get github.com/spf13/viper
go get github.com/joho/godotenv
go get go.uber.org/zap
```

**解释**：`go get` 下载第三方库并写入 `go.mod`。相当于 Java 的 `Maven Central` 或 `npm install`。

### 2.4 整理依赖

```bash
go mod tidy
```

**解释**：`go mod tidy` 自动分析代码里实际用到的依赖，删除没用的、补充缺少的，并生成 `go.sum`（依赖校验文件）。**每次新增/删除 import 后都应该执行一次。**

### 2.5 按顺序写代码

搭建顺序（有依赖关系，不能乱）：

1. **基础设施层**（无依赖）
   - `pkg/response/` — 统一 JSON 响应格式
   - `pkg/logger/` — Zap 日志初始化
   - `pkg/db/` — GORM 数据库连接
   - `internal/config/` — Viper 配置加载
   - `internal/middleware/` — Gin 中间件

2. **业务层**（依赖基础设施）
   - `internal/model/` — GORM 数据模型
   - `internal/repository/` — DAO（数据库操作）
   - `internal/service/` — 业务逻辑
   - `internal/handler/` — HTTP 处理器
   - `internal/router/` — 路由注册

3. **特色模块**
   - `internal/blockchain/` — 链交互封装
   - `internal/scheduler/` — 定时任务

4. **入口文件**
   - `cmd/api/main.go` — 组装所有组件并启动

### 2.6 编译与运行

```bash
# 编译为可执行文件
go build -o backend-api ./cmd/api

# 直接运行（开发常用）
go run ./cmd/api

# 运行编译好的文件
./backend-api
```

**解释**：
- `go build` 编译为二进制文件，不运行
- `go run` 编译+运行，不保留二进制文件
- `-o backend-api` 指定输出文件名（Windows 会自动加 `.exe`）

---

## 三、Go 命令速查

| 命令 | 作用 | 类比 Java |
|------|------|----------|
| `go mod init 模块名` | 初始化模块，创建 go.mod | `mvn init` |
| `go get 包路径` | 下载并安装依赖 | `mvn dependency:add` |
| `go mod tidy` | 清理无用依赖，补充缺失依赖 | `mvn clean install` 的一部分 |
| `go build` | 编译为可执行文件 | `mvn package` |
| `go run` | 编译并立即运行 | `java -jar` |
| `go test` | 运行测试 | `mvn test` |
| `go fmt` | 自动格式化代码 | IDEA 的 Reformat Code |
| `go vet` | 静态检查常见错误 | SpotBugs / FindBugs |

---

## 四、项目结构

```
backend/
├── cmd/api/main.go              # 入口：读取配置→初始化→启动HTTP
├── internal/
│   ├── config/config.go         # Viper 加载 configs/config.yaml + .env
│   ├── model/user.go            # GORM 数据模型（struct + tag）
│   ├── repository/user_repo.go  # DAO：Create/Find/Update/Delete
│   ├── service/user_service.go  # 业务逻辑层
│   ├── handler/user_handler.go  # HTTP Handler：解析请求、调用 Service、返回响应
│   ├── router/router.go         # 路由注册：URL → Handler
│   ├── middleware/middleware.go # 日志记录、错误恢复、CORS
│   ├── blockchain/client.go     # ethclient 封装：查余额、查区块号
│   └── scheduler/scheduler.go   # cron 定时任务注册与执行
├── pkg/
│   ├── db/mysql.go              # GORM 连接初始化
│   ├── logger/logger.go         # Zap 日志初始化
│   └── response/response.go     # 统一 JSON 响应格式 {code, message, data}
├── configs/config.yaml          # 配置文件（端口、数据库、区块链RPC）
├── .env                         # 环境变量（密码、私钥等敏感信息）
├── Makefile                     # 常用命令封装
├── README.md
└── go.mod                       # 模块定义 + 依赖列表
```

---

## 五、快速开始

### 5.1 配置

编辑 `configs/config.yaml`：

```yaml
server:
  port: 8080
  mode: debug

database:
  host: localhost
  port: 3306
  username: root
  password: ""
  dbname: backend_db
  charset: utf8mb4

blockchain:
  rpc_url: "https://sepolia.infura.io/v3/YOUR_INFURA_KEY"
  chain_id: 11155111
```

创建 `.env` 文件覆写敏感配置（**不上传到 Git**）：

```
DB_PASSWORD=your_mysql_password
BLOCKCHAIN_PRIVATE_KEY=your_private_key
```

### 5.2 启动

```bash
make run
```

或：

```bash
go run ./cmd/api
```

### 5.3 测试

```bash
curl http://localhost:8080/api/health
```

预期返回：

```json
{"code":0,"message":"ok"}
```

### 5.4 API 列表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/health | 健康检查 |
| POST | /api/users | 创建用户 |
| GET | /api/users | 用户列表 |
| GET | /api/users/:id | 获取用户 |
| PUT | /api/users/:id | 更新用户 |
| DELETE | /api/users/:id | 删除用户 |

---

## 六、开发命令（Makefile）

```bash
make build    # 编译为 backend-api 可执行文件
make run      # 直接运行（go run ./cmd/api）
make tidy     # 整理依赖（go mod tidy）
make clean    # 删除编译产物
```

**Makefile 是什么？** 一个文本文件，定义了一系列快捷命令。比如 `make run` 实际上就是执行 `go run ./cmd/api`，不用你每次手打长命令。

---

## 七、核心代码讲解

### 7.1 统一响应格式

```go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

所有 HTTP 接口统一返回这个结构：
- `code: 0` 表示成功，非 0 表示业务错误
- `message` 给前端显示
- `data` 放具体数据

### 7.2 三层组装（main.go 里的核心逻辑）

```go
userRepo := repository.NewUserRepository()        // 创建 DAO
userService := service.NewUserService(userRepo)    // DAO 注入 Service
userHandler := handler.NewUserHandler(userService) // Service 注入 Handler
r := router.Setup(userHandler)                     // Handler 注册到路由
```

这就是 Go 里最简单的"依赖注入"——手动 `new()` 然后传进去。没有 Spring 的自动装配，但意图非常清晰：谁依赖谁，一眼就能看到。

### 7.3 配置加载顺序

Viper 的加载优先级（从高到低）：
1. 环境变量（如 `DB_PASSWORD`）
2. `.env` 文件
3. `configs/config.yaml` 文件

这样开发时配置文件放默认值，敏感信息（密码、私钥）放 `.env`，生产环境直接读系统环境变量。

---

## 八、后续扩展

### 8.1 调用智能合约

当前 `internal/blockchain/client.go` 只有查询功能（查余额、查区块号）。要调用合约方法，需要：

1. 获取合约 ABI 和地址
2. 用 `abigen` 工具生成 Go 绑定代码（`go-ethereum` 自带）
3. 在 `blockchain/client.go` 里添加调用方法

示例：

```bash
# 根据 ABI 生成 Go 代码
abigen --abi=contract.abi --pkg=contract --out=contract.go
```

### 8.2 接入真实数据库

1. 安装 MySQL
2. 创建数据库：`CREATE DATABASE backend_db;`
3. 在 `.env` 里填密码：`DB_PASSWORD=你的密码`
4. 重启服务，GORM 会自动建表

### 8.3 添加 JWT 认证

在 `internal/middleware/` 里添加 JWT 中间件，在 `router.go` 里对需要保护的接口使用：

```go
api.Use(middleware.JWTAuth())
```

---

## 九、常见问题

**Q: `go mod tidy` 很慢或报错？**
A: 设置 Go 代理：`go env -w GOPROXY=https://goproxy.cn,direct`

**Q: 启动时报数据库连接失败？**
A: 检查 MySQL 是否运行、`configs/config.yaml` 里的用户名密码是否正确、`.env` 里的 `DB_PASSWORD` 是否已设置。

**Q: 如何调试 HTTP 接口？**
A: 推荐安装 [Postman](https://www.postman.com/) 或 [Apifox](https://www.apifox.cn/)，比 curl 更直观。

**Q: 如何查看日志？**
A: 日志默认输出到终端（stdout），是 JSON 格式。生产环境可以配置输出到文件或接入 ELK。
