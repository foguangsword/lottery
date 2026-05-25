# 透明抽奖 - 前端

原生 HTML + CSS + JavaScript，无框架。Nginx 静态部署。

## 页面

| 页面 | 文件 | 说明 |
|------|------|------|
| 首页 | `index.html` | 活动卡片列表，状态筛选 |
| 活动详情 | `activity.html` | 活动信息、参与者、中奖结果、验证数据 |
| 登录注册 | `login.html` | 登录/注册表单切换 |
| 个人中心 | `user.html` | 我的报名、中奖记录、账号设置 |
| 独立验证 | `verify.html` | 手动验证 seed 和数据完整性 |

## 开发

```shell
# 直接打开（本地开发）
open index.html

# 或使用任意静态服务器
python3 -m http.server 3000
```

## Docker 部署

```shell
# 启动
docker compose up -d

# 访问 http://localhost:8080

# 停止
docker compose down
```

## 对接后端

当前为纯前端原型，mock 数据在 `js/api.js`。对接后端时：

1. 将 `js/api.js` 中的 mock 函数替换为真实 fetch 调用
2. 后端 API 规范见项目根目录 README 的技术方案章节
