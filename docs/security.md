# 安全架构设计

## 资产清单

在讨论安全之前，先明确系统需要保护什么：

| 资产 | 存储位置 | 泄露/破坏后果 |
|------|----------|---------------|
| 合约 owner 私钥 | 后端 → KMS | 攻击者可创建假活动、伪造开奖结果 |
| 用户密码 hash | MySQL | 用户账号被盗，但密码本身不可逆 |
| 用户身份信息（邮箱等） | MySQL | 隐私泄露 |
| 报名 / 中奖数据 | 合约 + MySQL | 可被链上数据纠正，但前端展示受影响 |
| seed + 开奖参数 | 合约 input data + MySQL | 公开数据，但需保证可追溯 |

**核心风险是合约 owner 私钥。** 其他数据泄露影响有限（链上数据天然备份），但私钥泄露等同于整个平台作废。

## 威胁模型

```
                      ┌─────────────────────────┐
                      │      外部攻击者           │
                      │  DDoS / 漏洞扫描 / 社工   │
                      └───────────┬──────────────┘
                                  │
    ┌──────────┐    ┌──────────┐  │  ┌──────────┐    ┌──────────┐
    │  前端     │    │  SLB     │──┘  │  ECS     │    │  RDS     │
    │  XSS     │    │  HTTPS   │     │  后端    │    │  MySQL   │
    │  CSRF    │    │  DDoS    │     │  应用漏洞 │    │  拖库    │
    │  钓鱼    │    └──────────┘     │  私钥泄露 │    └──────────┘
    └──────────┘                     └──────────┘
                                          │
                                    ┌──────────┐
                                    │  以太坊   │
                                    │  合约攻击 │
                                    │  前置交易 │
                                    └──────────┘
```

### 按危害程度排序

| 优先级 | 威胁 | 危害 | 场景 | 当前状态 |
|--------|------|------|------|----------|
| P0 | 合约 owner 私钥泄露 | 灾难级 | 服务器沦陷 / 代码泄露 / 日志泄露 | 待接入 KMS |
| P0 | **平台操控开奖区块参数** | 灾难级 | `drawWinner` 的 `blockNumber`/`blockTs` 由后端传入，平台可暴力搜索有利参数 | **✅ 已确认修复方案：改为链上读取** |
| P1 | MySQL 被拖库 | 严重 | SQL 注入 / 弱密码 / 端口暴露 | 待加固 |
| P1 | API 未鉴权 | 严重 | 未登录可调用敏感接口 | 待实现 JWT |
| P2 | XSS / CSRF | 中等 | 窃取用户会话 / 伪造操作 | 待前端排查 |
| P2 | DDoS | 中等 | 服务不可用 | SLB 基础防护 |
| P2 | 报名阶段平台刷号 / 拒签 | 中等 | 后端代签模式下平台可控制链上报名名单 | **产品级 trade-off，需文档诚实声明** |
| P3 | 合约随机数攻击 | 低 | seed 由外部公开数据源提供，不依赖区块随机性 | 攻击面极小 |

## 云基础设施安全（阿里云）

### 最小方案（月费 ~300-500 元）

```
                          ┌─────────────────── VPC ───────────────────┐
                          │                                           │
  Internet ──→ SLB ──→ ECS (Docker) ──→ RDS MySQL (私网)             │
                HTTPS    Nginx + Go     无公网 IP，安全组仅允许 ECS    │
                         私钥不从 ECS 本地读取，走 KMS                 │
                         安全组：仅 443                               │
                         系统盘：加密                                  │
                         漏洞扫描：云安全中心（免费版）                  │
                         日志：SLS 免费额度                            │
```

**为什么这么选：**

| 服务 | 用途 | 选型理由 | 月费 |
|------|------|----------|------|
| SLB | HTTPS 终结 + 基础 DDoS 防护 | 比直接暴露 ECS 安全一层 | ~¥50 |
| ECS 1 台 | Nginx + Go 后端 | Docker 部署，一台够用 | ~¥150 (2c4g) |
| RDS MySQL | 数据库 | 自带备份 + 加密，免运维，比自建安全 | ~¥120 (1c1g 基础版) |
| KMS | 存储合约私钥 | 极低成本，密钥不出 HSM | ~¥0 (免费额度内) |
| 云安全中心 | 漏洞扫描 + 基线检查 | 免费版即可满足 | ¥0 |
| SLS | 操作审计日志 | 免费额度 500MB/天 | ¥0 |

**不买的：**

| 服务 | 为什么不买 |
|------|-----------|
| WAF | 月费 ¥3000+，业务体量小，SLB + Nginx 限流 + 安全组足够。做大再说 |
| Anti-DDoS Pro | 同上，SLB 自带基础防护 |
| Redis | Go 程序内存缓存够用，数据库查询量不大，省掉 |

### VPC 与安全组

```
VPC: 192.168.0.0/16
├── 交换机-公网: 192.168.1.0/24   # SLB
└── 交换机-私网: 192.168.2.0/24   # ECS + RDS

安全组规则（ECS）：
  - 入方向：TCP 443 来自 SLB 地址段
  - 入方向：TCP 22  来自公司 VPN/跳板机 IP（白名单）
  - 入方向：TCP 8080 来自 127.0.0.1（Go 后端健康检查）
  - 出方向：全拒绝，仅允许访问 RDS 端口 + Infura HTTPS

安全组规则（RDS）：
  - 入方向：TCP 3306 来自 ECS 安全组
  - 禁止任何公网访问
```

### SLB 配置

- 协议：HTTPS (TLS 1.2+)，HTTP 自动跳转 HTTPS
- 证书：阿里云免费证书或 Let's Encrypt
- 后端端口：ECS 的 80（Nginx 再反代到 Go）
- 健康检查：HTTP GET /health
- 访问控制：无（公开网站）

## 工程安全

### 合约层

**当前已满足：**
- `onlyOwner` 修饰符覆盖所有写操作，权限边界清晰
- Solidity ^0.8.x 内置溢出检查

**待修复：**

```solidity
// 1. drawWinner 区块参数改为链上读取（关键信任修复）
// 当前 blockNumber 和 blockTs 由后端传入，平台可通过暴力搜索微调参数来操控中奖结果。
// 修复方案：不再传入，直接使用 block.number 和 block.timestamp。
function drawWinner(string memory activityName, string memory seed, uint256 nowtime)
    public onlyOwner returns (uint256[] memory)
{
    uint256 blockNumber = block.number;
    uint256 blockTs = block.timestamp;
    // ... existing logic ...
}

// 2. triggerDraw() 残留代码 —— 直接删除
// 该函数将状态错误地设为 active（应为 drawing），且修改 startTime。
// drawWinner 可直接在活动截止后调用，无需 triggerDraw 过渡。

// 3. drawWinner 增加 seed 存储与事件，降低验证门槛
string public activitySeed;  // 或 mapping(string => string)
event WinnerDrawn(
    string indexed activityName,
    string seed,
    uint256 blockNumber,
    uint256 blockTs,
    uint256[] winnerIndexs
);

// 4. participate 补上状态和时间校验
modifier onlyActive(string memory activityName, uint256 nowtime) { ... }
// 当前 participate 没有使用 onlyActive，导致任何时间都能报名。

// 5. 移除 hardhat/console.sol 和多余的 payable
// 生产部署前必须移除调试导入；不接收 ETH 的函数不应标记 payable。

// 6. 修正状态名与拼写
// drawed -> drawn；stutas -> status
```

**不需要担心的：**
- 重入攻击：合约无 value 流转，无外部调用，天然免疫
- 前置交易 (frontrunning)：seed 来自外部公开数据源，无法通过交易排序影响结果
- 闪电贷攻击：不涉及 DeFi 逻辑

### Go 后端

```go
// 1. 所有输入必须校验
// 正例：使用结构体 + validator
type RegisterRequest struct {
    ActivityID uint   `json:"activity_id" validate:"required,min=1"`
    UserID     uint   `json:"user_id"     validate:"required,min=1"`
}

// 2. SQL 查询统一走 GORM 参数化
db.Where("user_id = ?", userID).Find(&registrations)
// 严禁：db.Raw("SELECT * FROM users WHERE id = " + userID)

// 3. 数据库连接
// DSN 从环境变量读取，禁止硬编码
dsn := os.Getenv("MYSQL_DSN")  // user:pass@tcp(rds-internal:3306)/db?tls=true

// 4. 密码存储
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// cost=10, 约 100ms/次，可接受
```

**中间件链（按顺序）：**

```
请求 → 限流 → CORS → JWT 认证 → 参数校验 → Handler
```

| 中间件 | 实现 | 说明 |
|--------|------|------|
| 限流 | `golang.org/x/time/rate` 或 gin 中间件 | 每 IP 100 req/min，登录接口 10 req/min |
| CORS | `gin-contrib/cors` | 仅允许自己的域名 |
| JWT | `golang-jwt/jwt` | 24h 过期，存 HttpOnly Cookie |
| 参数校验 | `go-playground/validator` | 白名单校验 |

**关键安全配置：**

```go
// Gin 生产配置
gin.SetMode(gin.ReleaseMode)

// JWT Cookie 属性
http.Cookie{
    Name:     "token",
    Value:    jwtString,
    HttpOnly: true,   // JavaScript 不可读，防 XSS 窃取
    Secure:   true,   // 仅 HTTPS 传输
    SameSite: http.SameSiteStrictMode,
    MaxAge:   86400,  // 24h
    Path:     "/",
}

// 请求体大小限制
router.MaxMultipartMemory = 8 << 20  // 8MB

// 超时设置
srv := &http.Server{
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}
```

**需要特别注意的点：**

- **日志脱敏**：绝对不在日志中打印 password、seed、私钥、JWT token
- **错误信息**：对外只返回通用错误（"用户名或密码错误"而非"用户不存在"），防止用户名枚举
- **合约调用重试**：nonce 管理要正确处理，避免双花
- **链上数据回写 MySQL**：交易确认后再写库，避免链上失败但数据库已更新

### 前端

```html
<!-- CSP 头（Nginx 层配置更合适，这里标注内容） -->
<!-- default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' fonts.googleapis.com;
     font-src fonts.gstatic.com; connect-src 'self'; img-src 'self' data:; -->

<!-- 所有用户输入必须转义 -->
<!-- 好：textContent -->
<span id="name"></span>
<script>document.getElementById('name').textContent = userInput;</script>

<!-- 坏：innerHTML 直接拼接用户输入 -->
<span id="name"></span>
<script>document.getElementById('name').innerHTML = userInput;</script>
```

**安全清单：**
- [ ] 无 `innerHTML` 直接拼接用户数据（当前原型有个别地方用了 innerHTML 动态渲染列表，数据来自 mock 而不是用户输入，但需要排查）
- [ ] 调用后端 API 时带上 `credentials: 'same-origin'`
- [ ] 表单提交前做客户端长度校验，减轻服务端压力
- [ ] 无 `eval()`、无 `document.write()`

### MySQL

```sql
-- 最小权限原则：应用账号只给 CRUD，不给 DDL
CREATE USER 'app_user'@'ecs-private-ip' IDENTIFIED BY 'strong-password';
GRANT SELECT, INSERT, UPDATE, DELETE ON lucky_draw.* TO 'app_user'@'ecs-private-ip';
-- 不给 DROP, ALTER, CREATE, GRANT

-- 强制 TLS 连接
ALTER USER 'app_user'@'ecs-private-ip' REQUIRE SSL;

-- 定期备份（RDS 自带，确认开启了自动备份）
```

## 密钥管理

### 合约 owner 私钥：三层防护

```
┌──────────────────────────────────────────────┐
│  第一层：KMS                                   │
│  私钥加密存储在 KMS，ECS 本地不存在明文           │
│  后端通过 SDK 调用 KMS 签名交易                  │
│  密钥从不出 KMS 边界                            │
├──────────────────────────────────────────────┤
│  第二层：专用地址                               │
│  合约 owner 地址仅用于此项目，不混用              │
│  建议定期轮换（changeOwner），旧地址废弃          │
├──────────────────────────────────────────────┤
│  第三层：交易确认                               │
│  关键操作（drawWinner）要求人工复核后再发交易      │
│  可在管理后台加二次确认                          │
└──────────────────────────────────────────────┘
```

**KMS 备选方案（如果觉得 KMS 接入复杂）：**

用环境变量 + Docker secret，私钥只在容器启动时解密读取，不落盘。但 KMS 方案更安全，建议初期就上。

### 其他密钥

| 密钥 | 存储 | 更换周期 |
|------|------|----------|
| JWT signing key | KMS 或环境变量 | 每次发版 |
| MySQL 密码 | 环境变量 | 90 天 |
| Infura API key | 环境变量 | 不主动暴露即可 |

## 生产部署清单

### ECS 加固

```bash
# 1. 创建非 root 用户
useradd -m -s /bin/bash app
usermod -aG docker app

# 2. SSH 加固 /etc/ssh/sshd_config
PermitRootLogin no
PasswordAuthentication no          # 仅密钥登录
PubkeyAuthentication yes
MaxAuthTries 3

# 3. 安装 fail2ban
apt install fail2ban
# 配置：SSH 5 次失败封禁 30 分钟

# 4. 自动安全更新
apt install unattended-upgrades
# 配置：安全更新自动安装

# 5. 防火墙（除安全组外再加一层）
ufw default deny incoming
ufw allow 80
ufw allow 443
ufw allow from <管理IP> to any port 22
ufw enable
```

### Docker 安全

```dockerfile
# 生产 Dockerfile（相比开发用的简单版，加上安全措施）
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/api

FROM alpine:3.20
RUN addgroup -S app && adduser -S app -G app
COPY --from=builder /app/server /app/server

# 非 root 运行
USER app

# 只读文件系统（需要写临时目录的除外）
# 如果程序不需要写本地文件，加: --read-only

EXPOSE 8080
CMD ["/app/server"]
```

### Nginx 加固

```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$host$request_uri;  # HTTP 强制跳转 HTTPS
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    # TLS 配置（仅安全套件）
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # 安全头
    add_header X-Content-Type-Options nosniff;
    add_header X-Frame-Options DENY;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src https://fonts.gstatic.com; connect-src 'self'; img-src 'self' data:;";

    # 隐藏版本号
    server_tokens off;
    proxy_hide_header X-Powered-By;

    # 限流
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req zone=api burst=20 nodelay;

    location / {
        root /usr/share/nginx/html;
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

        # 请求体大小限制
        client_max_body_size 1m;
    }
}
```

### 上线前自检清单

- [ ] `go mod tidy && go vet ./...` 无报错
- [ ] GORM 日志级别设为 `Warn`（生产模式，不打印 SQL）
- [ ] Gin 模式设为 `ReleaseMode`
- [ ] 所有环境变量已配置，无硬编码
- [ ] MySQL 连接开启了 TLS
- [ ] 合约已部署到目标网络，地址已确认
- [ ] 安全组规则已确认，仅开放必要端口
- [ ] RDS 开启了自动备份
- [ ] SLB 证书已配置且未过期
- [ ] 域名 DNS 已指向 SLB
- [ ] 在本地跑过一次完整流程（注册 → 登录 → 报名 → 开奖 → 验证）

## 运维安全

### 日常运营

| 事项 | 频率 | 方式 |
|------|------|------|
| 查看 ECS 安全告警 | 每天 | 云安全中心控制台 |
| 检查合约 owner 地址余额 | 每次发交易前 | 确保有 ETH 付 gas |
| RDS 备份验证 | 每月 | 恢复到一个测试实例 |
| SSL 证书续期 | 到期前 30 天 | 阿里云自动续期 |
| 依赖更新 | 每月 | `go list -u -m all` 检查 |

### 应急预案

| 事件 | 处置 |
|------|------|
| 疑似私钥泄露 | 1. 立即 `changeOwner` 到新地址 2. 排查泄露源 3. 评估是否需要作废当前活动 |
| MySQL 被拖 | 1. 安全组关公网 2. 改密码 3. 排查注入点 4. 通知用户改密码（bcrypt 降低了风险） |
| 合约被异常调用 | 链上不可逆，但可以 changeOwner 阻断后续操作，然后排查 |
| ECS 被入侵 | 1. 创建快照取证 2. 从安全基线镜像重建 3. 检查是否有持久化后门 |

### 日志审计

SLS 需要收集的日志：

```
ECS /var/log/auth.log    →  SSH 登录记录
ECS /var/log/nginx/*.log →  访问日志（留 30 天）
Go 应用日志              →  业务操作（注册、登录、报名、开奖），脱敏后输出
合约事件                 →  createActivity, drawWinner 等
```

**关键审计点：**
- 谁、什么时间、调用了 `drawWinner`
- seed 值是多少
- 对应的上证/深证收盘价是什么

## 总结

核心安全原则——**不信任单一环节**：

```
链上合约（公开透明）
    +
外部随机源（不可预测、不可操控）
    +
合约私钥不落盘（KMS）
    +
数据库加密 + 备份
    +
ECS 最小暴露面（安全组 + 非 root）

= 攻击者需要在多个层面同时突破才能造成实质性损害
```

这个系统的设计天然有一定安全优势：抽奖结果在链上，即使服务器被攻破，攻击者也只能破坏后续活动（换 owner），已开奖的结果不可篡改。安全工作的重心永远是 **保护好合约私钥**。

### 信任边界声明

本系统不是"完全去信任"的，它的保障是**分阶段**的：

- **报名截止后 → 开奖完成**：这一阶段是密码学保障的。参与者名单上链锁定，seed 来自公开数据，开奖算法确定性执行，任何人可独立验证。
- **报名阶段**：由于采用"后端代签、用户无钱包"模式，平台在理论上可以拒绝真实用户报名，也可以用自己的地址刷号增加中奖概率。这一阶段的公平性依赖平台的商业信誉和运营透明度，无法通过纯代码消除。

**这是"无门槛参与"和"完全去信任"之间的必要 trade-off。** 如果未来产品定位转向高价值抽奖，应考虑引入 KYC、质押或灵魂绑定代币等身份层。
