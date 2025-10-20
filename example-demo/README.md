# Go Analysis Client Demo

这是一个完整的 go-analysis-client 使用示例，演示了如何集成和使用分析客户端。

## 功能展示

这个演示程序会发送以下类型的事件：

1. **安装信息** - 完整的设备信息（OS、平台、主机名等）
2. **应用生命周期** - 启动和退出事件
3. **用户行为** - 登录、页面浏览、按钮点击
4. **功能使用** - 搜索、过滤、导出、分享
5. **电商事件** - 商品浏览、加购物车、结账、购买
6. **错误追踪** - 错误和异常信息
7. **性能指标** - 页面加载、API响应、渲染时间
8. **用户交互** - 滚动、悬停、选择等交互流
9. **业务事件** - 订阅、许可证验证等自定义事件

## 前置条件

1. **启动 go-analysis-server**
   ```bash
   cd ../go-analysis-server
   go run main.go
   ```
   服务器应该运行在 `http://localhost:8097`

2. **启动前端（可选）**
   ```bash
   cd ../go-analysis-frontend
   npm install
   npm run dev
   ```
   前端应该运行在 `http://localhost:3000`

## 运行方式

### 方式一：使用脚本（推荐）

```bash
./run.sh
```

这个脚本会：
- 检查服务器和前端是否运行
- 自动安装依赖
- 运行演示程序
- 显示查看结果的链接

### 方式二：手动运行

```bash
# 安装依赖
go mod tidy

# 运行
go run main.go
```

## 查看结果

运行完成后，你可以在以下位置查看结果：

### 1. 前端界面（推荐）

访问 http://localhost:3000

你会看到：
- **事件列表** - 所有发送的事件（按时间排序）
- **设备信息** - 完整的设备详情
- **统计图表** - 事件数量、类型分布等
- **用户行为分析** - 会话信息、活跃度等

### 2. 数据库直接查询

```sql
-- 查看安装信息
SELECT * FROM installs WHERE product = 'DemoApp' ORDER BY created_at DESC LIMIT 1;

-- 查看所有事件
SELECT * FROM events WHERE product = 'DemoApp' ORDER BY timestamp DESC;

-- 按事件类型统计
SELECT 
    name, 
    COUNT(*) as count,
    COUNT(DISTINCT device_id) as unique_devices
FROM events 
WHERE product = 'DemoApp'
GROUP BY name
ORDER BY count DESC;

-- 查看设备信息
SELECT * FROM devices WHERE device_id = (
    SELECT device_id FROM installs WHERE product = 'DemoApp' ORDER BY created_at DESC LIMIT 1
);
```

## 预期输出

运行后你应该看到类似以下的输出：

```
=== Go Analysis Client Demo Started ===
Device ID: 550e8400-e29b-41d4-a716-446655440000
Session ID: 123e4567-e89b-12d3-a456-426614174000

📦 Reporting installation info...
✅ Install info reported successfully

🚀 Tracking app launch...

👤 Simulating user login...

📄 Simulating page views...
  - Viewed: /home
  - Viewed: /products
  ...

🖱️  Simulating button clicks...
  - Clicked: submit
  - Clicked: cancel
  ...

[更多事件...]

=== Demo Completed Successfully ===
📈 Check your analytics dashboard for the results!
🔗 Frontend URL: http://localhost:3000
📊 Events sent for product: DemoApp
🆔 Device ID: 550e8400-e29b-41d4-a716-446655440000
```

## 前端展示内容

在前端界面，你将看到：

### 1. 产品概览
- 产品名称：DemoApp
- 总事件数：~40+ 个事件
- 设备数：1 个设备
- 活跃用户：1 个用户

### 2. 事件列表
按时间排序的所有事件，包括：
- 事件名称（如：user_login, page_view, button_click）
- 时间戳
- 属性详情（JSON 格式）
- 设备 ID
- 会话 ID

### 3. 设备信息
- 设备 ID
- 操作系统信息（macOS/Linux/Windows）
- 平台版本
- 主机名
- 内核版本
- 架构信息
- 运行时间

### 4. 统计图表
- 事件数量趋势（时间序列）
- 事件类型分布（饼图）
- 用户活跃度（柱状图）
- 页面访问排行
- 功能使用频率

### 5. 会话分析
- 会话时长
- 事件流（按时间顺序）
- 用户路径追踪

## 自定义演示

你可以修改 `main.go` 来：

1. **修改产品名称**
   ```go
   client := analytics.NewClient(
       "http://localhost:8097",
       "YourProductName", // 改成你的产品名
       ...
   )
   ```

2. **添加更多事件**
   ```go
   client.Track("your_event", map[string]interface{}{
       "key": "value",
   })
   ```

3. **修改用户 ID**
   ```go
   analytics.WithUserID("your-user-id"),
   ```

4. **启用加密传输**
   ```go
   analytics.WithEncryption("your-32-byte-secret-key-here!!"),
   ```

## 故障排除

### 问题：服务器连接失败

**解决方法：**
1. 确认 go-analysis-server 正在运行
2. 检查端口 8097 是否被占用
3. 查看服务器日志是否有错误

### 问题：前端看不到数据

**解决方法：**
1. 刷新前端页面
2. 检查是否选择了正确的产品（DemoApp）
3. 查看浏览器控制台是否有错误
4. 确认数据库连接正常

### 问题：设备信息不完整

**解决方法：**
- 某些系统可能无法获取全部设备信息
- 这是正常的，SDK 会提供可用的信息

## 性能说明

- 所有事件都是**异步发送**的，不会阻塞主程序
- 默认批量发送（20个事件/批次）
- 自动重试机制
- 本地缓冲防止事件丢失

## 下一步

1. 查看 [../README.md](../README.md) 了解完整的 SDK 文档
2. 查看 [../QUICKSTART.md](../QUICKSTART.md) 了解快速集成指南
3. 查看其他示例：
   - [example-standalone](../example-standalone/) - 简单示例
   - [example-gin](../example-gin/) - Gin 框架集成
   - [example-aes](../example-aes/) - 加密传输示例

## 技术支持

如有问题，请查看：
- GitHub Issues
- 文档：[../README.md](../README.md)
- API 参考：[../QUICKSTART.md](../QUICKSTART.md)
