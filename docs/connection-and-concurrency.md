# 数据库连接与并发模型：完整总结

## 一、四个核心概念

在讨论数据库并发之前，必须先厘清这四个概念。它们位于不同层面，容易混淆：

| 概念 | 所在层面 | 含义 |
|------|---------|------|
| **线程 / 协程** | 应用调度层 | 程序内部的执行单元。Java 用 OS 线程，Go 用 goroutine |
| **TCP 连接** | 网络传输层 | 客户端和服务端之间的一条通信管道 |
| **阻塞 / 非阻塞** | I/O 调用方式 | 调用方发起请求后，是干等着（阻塞），还是立刻返回做别的事（非阻塞）|
| **独占 / 共享** | 连接使用模式 | 一个连接同一时刻只能被一个查询用（独占），还是可以同时服务多个查询（共享）|

**关键认知：这四个概念是正交的，可以任意组合。**

---

## 二、三种模型横向对比

### 2.1 对比表

| 维度 | Java JDBC + HikariCP | Java R2DBC | Go database/sql |
|------|---------------------|------------|-----------------|
| **代表** | `ConnectionPool` + `PreparedStatement` | `r2dbc-postgresql` / `r2dbc-mysql` | `sql.DB` + `go-sql-driver/mysql` |
| **调用方阻塞？** | ✅ **线程阻塞**。`executeQuery()` 卡住线程直到数据库响应 | ❌ **线程不阻塞**。返回 `Publisher`，通过回调处理结果 | ✅ **goroutine 阻塞**。`db.Query()` 卡住 goroutine 直到响应 |
| **TCP I/O 模型** | **BIO（阻塞 I/O）**。线程干等着网络数据回来 | **NIO（Netty 异步 I/O）**。I/O 事件由 EventLoop 处理，线程不被网络阻塞 | **阻塞 I/O**。goroutine 干等着网络数据回来 |
| **连接使用模式** | **独占**。一个连接同一时刻只服务一个查询 | **独占**。一个连接同一时刻只服务一个查询 | **独占**。一个连接同一时刻只服务一个查询 |
| **连接池** | 有。`MaxPoolSize` 控制最大连接数 | 有。`r2dbc-pool`，但获取连接也是非阻塞的（返回 Mono） | 有。`MaxOpenConns` 控制最大连接数 |
| **并发扩展性** | 差。线程太重（1MB 栈），线程池容易打满 | 好。线程不阻塞，少量线程可处理大量并发 | 好。goroutine 极轻（2KB 栈），可开大量协程 |
| **代码心智负担** | 低。同步顺序代码，好读好调试 | 高。Reactive 回调链，调试困难 | 低。同步顺序代码，好读好调试 |

### 2.2 一句话总结三种模型

- **Java JDBC**：线程阻塞 + TCP 阻塞 + 连接独占 = 线程池和连接池都是瓶颈
- **Java R2DBC**：线程不阻塞 + TCP 异步 + 连接独占 = 解决了线程瓶颈，但连接数仍要控制
- **Go database/sql**：goroutine 阻塞 + TCP 阻塞 + 连接独占 = 线程侧无瓶颈（goroutine 便宜），但连接数仍是瓶颈

---

## 三、核心问题：为什么数据库连接总是"独占"的？

### 3.1 根本原因：SQL 连接是"有状态"的

HTTP 可以做连接共享（HTTP/2 多路复用），因为 HTTP 是**无状态**的：

```http
GET /api/user/1     ← 请求 A
GET /api/user/2     ← 请求 B（和 A 完全独立，可以共用连接）
```

但 SQL 连接是**有状态**的：

```sql
SET SESSION timezone = 'Asia/Shanghai';  -- 连接级状态
BEGIN;                                     -- 事务状态
SELECT * FROM temp_table;                  -- 临时表
```

如果两个查询共享同一个连接：
- A 设置了 session 变量，B 的查询也会被影响
- A 开了事务，B 的 `SELECT` 可能误进入 A 的事务
- A 建了临时表，B 能读到 A 的数据

所以数据库协议设计之初就是 **"一个连接 = 一个会话上下文"**，独占是必要语义。

### 3.2 有没有例外？

| 场景 | 为什么可以共享 | 实现方式 |
|------|---------------|---------|
| **Redis** | Redis 服务端是**单线程**的，请求严格按顺序处理。TCP 本身的顺序性保证：请求顺序 = 响应顺序 | Lettuce 单连接多路复用，无需 RequestID |
| **HTTP/2** | HTTP 是无状态的，Stream ID 区分不同请求 | 协议内置 Stream ID，一个 TCP 连接上多个 Stream |
| **MySQL X Protocol** | MySQL 8.0 新协议，支持异步操作 | 需要专用驱动（不是 go-sql-driver） |
| **PostgreSQL Pipeline** | 可以在一个连接上排队多个请求 | `pgx` 驱动的 `Batch`，但仍按顺序对应 |

**注意**：Pipeline 不等于多路复用。Pipeline 是"排队顺序执行"，多路复用是"并发交叉执行"。

### 3.3 连接共享需要协议支持什么？

要实现真正的"一个连接服务多个并发查询"，协议必须解决**响应归属**问题：

```
Client ──Query A──┐
                   ├──→ 共享的 TCP 连接 ──→ Server
Client ──Query B──┘
                      ←── Response ??? ──
```

Server 返回一个结果，Client 要知道这是 A 的还是 B 的。解决方案：

1. **Request ID**（如 HTTP/2 Stream ID）：每个请求带唯一 ID，响应带回相同 ID
2. **顺序保证**（如 Redis）：服务端单线程处理，请求顺序 = 响应顺序

MySQL 传统协议（`go-sql-driver/mysql` 用的）**没有 Request ID**，也没有顺序保证（服务端是多线程的），所以无法实现连接共享。

---

## 四、Go 开发者的实际启示

### 4.1 连接池配置不是可选的

```go
sqlDB, err := db.DB()
sqlDB.SetMaxOpenConns(25)      // 最大打开连接数（默认0=无限制，但会耗尽系统资源）
sqlDB.SetMaxIdleConns(5)       // 最大空闲连接数
sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大存活时间
```

**不配置的后果**：并发量高了之后，连接数爆炸，数据库端 `Too many connections`。

### 4.2 Worker Pool 和连接池的关系

```go
// Worker Pool 控制 goroutine 数量
const numWorkers = 50

// 数据库连接池控制连接数量（通过 GORM/DB 配置）
// 如果 numWorkers > MaxOpenConns，多余的 worker 会在取连接时阻塞
```

**黄金法则**：

```
Worker 数量 ≥ MaxOpenConns
```

如果 Worker 太少，连接池利用率低；如果 Worker 太多，大量 goroutine 阻塞在等连接。

### 4.3 为什么 Go 不搞 R2DBC？

| 问题 | Java 的痛点 | Go 的情况 |
|------|------------|----------|
| 线程阻塞 | OS 线程 1MB 栈，开 1w 个线程 = 10GB 内存 | goroutine 2KB 栈，开 1w 个 = 20MB |
| 性能 | 线程上下文切换开销大 | goroutine 调度极快 |
| 复杂度 | 必须用异步才能支撑高并发 | 阻塞 goroutine 就够用了 |

Go 的结论：**阻塞 goroutine 已经足够便宜，没必要引入 Reactive 的复杂度。**

但代价是：数据库连接数和 Java 一样，都是硬瓶颈。

---

## 五、完整流程图

### 5.1 Java JDBC（阻塞线程 + 阻塞 I/O）

```
请求进来 → 线程池分配线程 T1
    T1 → 从连接池取 Conn1
    T1 → 发送 SQL → 【T1 阻塞等网络响应】
    DB 返回结果
    T1 → 处理结果 → 归还 Conn1
    T1 → 返回响应

瓶颈：线程数（1MB/线程）+ 连接数
```

### 5.2 Java R2DBC（非阻塞线程 + 异步 I/O）

```
请求进来 → EventLoop 线程 E1（不阻塞）
    E1 → 从连接池取 Conn1（非阻塞，返回 Mono）
    E1 → 发送 SQL → 【立即返回，E1 处理其他请求】
    DB 返回结果 → Netty 回调 → 另一个线程处理结果
    → 归还 Conn1

瓶颈：连接数（线程侧无瓶颈）
```

### 5.3 Go database/sql（阻塞 goroutine + 阻塞 I/O）

```
请求进来 → 启动 goroutine G1（极轻量）
    G1 → 从连接池取 Conn1
    G1 → 发送 SQL → 【G1 阻塞等网络响应，但 OS 线程被调度器回收】
    DB 返回结果
    G1 → 处理结果 → 归还 Conn1
    G1 → 返回响应

瓶颈：连接数（goroutine 侧无瓶颈）
```

---

## 六、总结

1. **数据库连接在协议层面是独占的**，这是 SQL 有状态性决定的，不是驱动实现能轻易突破的
2. **Go 的 `database/sql` 本质和 Java 传统连接池一样**，都是"取连接→执行→归还"的独占模型
3. **Go 的优势不在 I/O 模型，而在 goroutine 调度**：阻塞 goroutine 比阻塞 OS 线程便宜 500 倍
4. **连接数是所有模型的共同瓶颈**，无论 Java/Go/R2DBC，都要配置 `MaxOpenConns`
5. **Worker Pool 的数量应该 ≥ 连接池大小**，否则连接资源浪费；太多则 goroutine 空等连接
6. **真正的连接多路复用需要协议支持**（Request ID 或顺序保证），MySQL 传统协议不具备这个条件

---

## 七、延伸阅读

- [Go database/sql 连接池源码解析](https://go.dev/doc/database/manage-connections)
- [R2DBC 规范](https://r2dbc.io/spec/1.0.0.RELEASE/spec/html/)
- [HTTP/2 多路复用原理 (RFC 7540)](https://datatracker.ietf.org/doc/html/rfc7540)
- [Lettuce 单连接多路复用实现](https://lettuce.io/core/release/reference/#_connection_pooling)
