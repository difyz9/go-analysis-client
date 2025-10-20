# 🎉 Go Analysis Client 演示成功运行报告

## ✅ 运行结果

演示程序已成功运行并发送了所有事件到 go-analysis-server！

### 发送的数据统计

- **产品名称**: DemoApp
- **设备ID**: `67e72c13-e684-5662-ad51-975df7b07f8d`
- **会话ID**: `7ad3e173-ab18-4cbd-8a20-b23954eae105`
- **总事件数**: ~33 个事件
- **事件批次**: 4 批次发送

### 发送的事件类型

1. ✅ **安装信息** - 包含完整设备信息
2. ✅ **app_launch** - 应用启动事件
3. ✅ **user_login** - 用户登录事件
4. ✅ **page_view** (5次) - 页面浏览事件
5. ✅ **button_click** (5次) - 按钮点击事件
6. ✅ **feature_use** (4次) - 功能使用事件
7. ✅ **product_view** - 商品浏览
8. ✅ **add_to_cart** - 添加购物车
9. ✅ **checkout_start** - 开始结账
10. ✅ **purchase** - 完成购买
11. ✅ **error** - 错误追踪
12. ✅ **performance** (3次) - 性能指标
13. ✅ **interaction** (5次) - 用户交互
14. ✅ **subscription_activated** - 订阅激活
15. ✅ **license_verified** - 许可证验证
16. ✅ **app_exit** - 应用退出

## 📊 如何查看数据

### 方式 1: 前端界面（最佳体验）

1. **启动前端**（如果还没启动）:
   ```bash
   cd /Users/apple/opt/difyz10/1018/go-analysis-frontend
   npm install
   npm run dev
   ```

2. **访问前端**:
   打开浏览器访问: http://localhost:3000

3. **查看内容**:
   - 📈 **仪表板**: 总览统计数据
   - 📋 **事件列表**: 查看所有 DemoApp 的事件
   - 💻 **设备信息**: 查看设备详细信息
   - 📊 **图表分析**: 事件趋势、类型分布等
   - 🔍 **筛选功能**: 按产品、事件类型、时间范围筛选

### 方式 2: 数据库直接查询

使用 PostgreSQL 客户端连接数据库:

```bash
# 连接信息（来自 config.toml）
Host: 124.222.202.16
Port: 5432
Database: crypto_wallet
Username: postgres
Password: bJYnUfNUtSAs
```

**查询示例** (见 queries.sql 文件):

```sql
-- 1. 查看安装信息
SELECT * FROM installs WHERE product = 'DemoApp';

-- 2. 查看所有事件
SELECT * FROM events WHERE product = 'DemoApp' ORDER BY timestamp DESC;

-- 3. 事件统计
SELECT name, COUNT(*) as count 
FROM events 
WHERE product = 'DemoApp' 
GROUP BY name 
ORDER BY count DESC;
```

### 方式 3: API 查询

使用 HTTP API 直接查询:

```bash
# 1. 查询事件
curl "http://localhost:8097/api/events/query?product=DemoApp&limit=50"

# 2. 查询统计信息
curl "http://localhost:8097/api/stats?product=DemoApp"

# 3. 查询安装信息
curl "http://localhost:8097/api/installs/query?product=DemoApp"
```

## 📋 在前端你将看到的内容

### 1. 产品概览卡片
```
┌─────────────────────────────────┐
│ 🎯 DemoApp                      │
├─────────────────────────────────┤
│ 📊 总事件: 33                   │
│ 💻 设备数: 1                    │
│ 👥 活跃用户: 1                  │
│ 📅 最后活动: 刚刚               │
└─────────────────────────────────┘
```

### 2. 事件时间线
```
🚀 15:49:02 - app_launch
   └─ version: 1.0.0, build_number: 100

👤 15:49:02 - user_login
   └─ method: email, success: true

📄 15:49:02 - page_view
   └─ page: /home, duration: 28s

🖱️ 15:49:03 - button_click
   └─ button_name: submit, screen: main

... (更多事件)

🛒 15:49:08 - purchase
   └─ amount: 99.99, currency: USD
```

### 3. 设备信息卡片
```
┌─────────────────────────────────┐
│ 💻 设备信息                     │
├─────────────────────────────────┤
│ 🆔 Device ID:                   │
│    67e72c13-e684-5662-ad51...   │
│                                 │
│ 🖥️  操作系统: darwin            │
│ 📦 平台: darwin                 │
│ 🏷️  主机名: apple15             │
│ 🔧 架构: arm64                  │
│ ⏱️  运行时间: [系统运行时间]    │
└─────────────────────────────────┘
```

### 4. 统计图表

**事件类型分布（饼图）**:
- page_view: 5 次
- button_click: 5 次
- interaction: 5 次
- feature_use: 4 次
- performance: 3 次
- ... (其他事件)

**事件趋势（折线图）**:
显示从 15:49:02 到 15:49:15 的事件分布

**热门页面（柱状图）**:
1. /home
2. /products
3. /about
4. /contact
5. /pricing

## 🔄 再次运行演示

你可以多次运行演示程序来生成更多数据:

```bash
cd /Users/apple/opt/difyz10/1018/go-analysis-client/example-demo
./run.sh
```

每次运行都会:
- 使用相同的设备ID（稳定的设备标识）
- 生成新的会话ID（新的使用会话）
- 发送所有类型的事件
- 累积到数据库中

## 🎨 自定义和扩展

### 修改产品名称

编辑 `main.go` 的第 15 行:
```go
client := analytics.NewClient(
    "http://localhost:8097",
    "MyCustomApp", // 改成你的产品名
    ...
)
```

### 添加自定义事件

在 `main.go` 中添加:
```go
client.Track("my_custom_event", map[string]interface{}{
    "custom_field": "custom_value",
    "number": 123,
    "boolean": true,
})
```

### 启用加密传输

添加加密选项:
```go
client := analytics.NewClient(
    "http://localhost:8097",
    "DemoApp",
    analytics.WithDebug(true),
    analytics.WithEncryption("your-32-byte-secret-key-here!!"), // AES-256
)
```

然后在服务器 `config.toml` 中启用解密:
```toml
[aes]
  enabled = true
  secret_key = "your-32-byte-secret-key-here!!"
  enable_decryption = true
```

## 📈 数据分析建议

基于收集的数据，你可以:

1. **用户行为分析**
   - 页面访问路径
   - 功能使用频率
   - 按钮点击热力图

2. **电商分析**
   - 转化漏斗: 浏览 → 加购 → 结账 → 购买
   - 平均订单金额
   - 购买成功率

3. **性能监控**
   - 页面加载时间
   - API 响应时间
   - 渲染性能

4. **错误追踪**
   - 错误类型分布
   - 错误发生频率
   - 受影响的用户数

5. **设备分析**
   - 操作系统分布
   - 设备型号统计
   - 活跃设备趋势

## 🚀 下一步

1. ✅ **集成到你的应用**: 参考 [../README.md](../README.md)
2. ✅ **配置前端**: 查看 [../../go-analysis-frontend/README.md](../../go-analysis-frontend/README.md)
3. ✅ **生产环境部署**: 查看服务器部署文档
4. ✅ **自定义仪表板**: 根据业务需求定制前端展示

## 📞 技术支持

如有问题:
- 📖 查看文档: [../README.md](../README.md)
- 🐛 报告 Bug: GitHub Issues
- 💬 讨论: GitHub Discussions

---

**祝贺！🎉 你已经成功运行了完整的分析系统！**
