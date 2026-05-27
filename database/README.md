# Database

本地数据库环境，使用 Docker 启动 MySQL。

## 快速开始

### 1. 启动 MySQL

确保已安装 Docker，然后在当前目录执行：

```bash
docker compose up -d
```

这会启动一个 MySQL 8.0 容器：
- 端口映射：`localhost:3306`
- 数据库：`lottery_db`
- 用户：`lottery` / `lottery123`
- root 密码：`rootpass`

**首次启动**时，MySQL 会自动执行 `migrations/` 目录下的 `001_init.sql` 完成建表。

### 2. 查看状态

```bash
docker compose ps
docker logs lottery-mysql
```

### 3. 连接数据库

```bash
# 用容器内的 mysql 客户端
docker exec -it lottery-mysql mysql -u lottery -p lottery123 lottery_db

# 或直接用后端连接（backend/configs/config.yaml 已配好）
```

### 4. 停止 / 重启

```bash
# 停止（数据保留在 volume 中）
docker compose down

# 彻底删除（包括数据！慎用）
docker compose down -v

# 重启
docker compose restart
```

## 目录说明

```
database/
├── docker-compose.yml      # MySQL 容器配置
├── migrations/
│   └── 001_init.sql        # 初始建表脚本
└── README.md
```

## 数据持久化

MySQL 数据通过 Docker Volume `mysql_data` 持久化。即使容器删除重建，数据不会丢失。

```bash
# 查看 volume
docker volume ls | grep mysql_data

# 手动备份数据
docker exec lottery-mysql mysqldump -uroot -prootpass lottery_db > backup.sql

# 从备份恢复
cat backup.sql | docker exec -i lottery-mysql mysql -uroot -prootpass lottery_db
```

## 数据库管理方案

方案工作流

  1. docker compose up -d          # 启动 MySQL
  2. DBeaver 连接 localhost:3306   # 用 lottery/lottery123
  3. 首次：执行 001_init.sql       # 手动点运行，或者让 Docker 自动执行
  4. 后续改表：写新的 SQL 脚本，DBeaver 里执行
  5. Go 后端只管 CRUD，不碰表结构

  这是标准的"DBA 模式"，团队越大越需要这样。



## !-----以下暂不采用-----！

## 手动执行迁移脚本（如需）

如果容器已经启动过，之后新增了 SQL 脚本，可以手动执行：

```bash
docker exec -i lottery-mysql mysql -ulottery -plottery123 lottery_db < migrations/002_add_column.sql
```

> 生产环境建议使用专业的迁移工具（如 golang-migrate），见下方进阶说明。

## 进阶：手动迁移 vs GORM AutoMigrate

### 方式一：GORM AutoMigrate（当前后端默认）

后端 `main.go` 里调用：

```go
db.DB.AutoMigrate(&model.User{}, &model.Activity{})
```

**优点：** 开发方便，改 model 后重启服务自动同步表结构  
**缺点：** 生产环境危险，不会安全地删除列/修改类型，容易误伤数据

### 方式二：手动 SQL 迁移（推荐用于生产）

1. **安装 golang-migrate CLI**
   
   ```bash
   # Windows (PowerShell)
   scoop install golang-migrate
   
   # Mac
   brew install golang-migrate
   ```
   
2. **创建迁移文件**
   ```bash
   migrate create -ext sql -dir migrations -seq add_email_index
   ```
   这会生成 `002_add_email_index.up.sql` 和 `.down.sql`

3. **在 main.go 启动时执行迁移**
   
   ```go
   import "github.com/golang-migrate/migrate/v4"
   
   m, _ := migrate.New("file://database/migrations", "mysql://lottery:lottery123@tcp(localhost:3306)/lottery_db")
   m.Up() // 执行所有待执行的 up 迁移
   ```
   
4. **版本控制**
   
   - `migrations` 目录里的 `.sql` 文件纳入 Git
   - 每次改表结构都新建一个迁移文件，不修改旧的
   - 数据库里会自动创建 `schema_migrations` 表记录当前版本

**总结：** 开发阶段可以用 AutoMigrate 图快，正式上线前切到手动迁移。
