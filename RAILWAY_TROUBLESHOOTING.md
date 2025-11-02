# 🚨 Railway部署故障排除指南

## 🎯 故障排除概览

本指南帮助您诊断和解决NOFX AI交易系统在Railway部署过程中可能遇到的问题。

## 🔍 常见问题分类

### 📦 构建阶段问题

#### 问题1：Docker构建失败
**症状：**
```
Error: failed to solve: failed to read dockerfile
```

**可能原因：**
- Dockerfile语法错误
- 缺少必要文件
- 权限问题

**解决方案：**
1. 检查Dockerfile语法：
   ```bash
   # 本地验证Dockerfile
   docker build -t test-build .
   ```

2. 确保必要文件存在：
   - [ ] `go.mod`
   - [ ] `go.sum`
   - [ ] `main.go`
   - [ ] `config/` 目录

3. 检查.dockerignore文件，确保没有排除必要文件

#### 问题2：Go依赖下载失败
**症状：**
```
Error: go mod download failed
```

**解决方案：**
1. 验证go.mod文件格式：
   ```go
   module nofx
   
   go 1.21
   
   require (
       // 依赖列表
   )
   ```

2. 更新依赖：
   ```bash
   go mod tidy
   go mod download
   ```

3. 检查网络连接和代理设置

#### 问题3：编译失败
**症状：**
```
Error: build failed with exit code 1
```

**解决方案：**
1. 本地编译测试：
   ```bash
   CGO_ENABLED=0 GOOS=linux go build -o nofx main.go
   ```

2. 检查Go版本兼容性
3. 修复代码语法错误

### 🚀 启动阶段问题

#### 问题4：应用启动失败
**症状：**
- 应用立即崩溃
- 健康检查失败
- 无法访问端点

**诊断步骤：**
1. 查看Railway日志：
   ```
   Railway控制台 → Deployments → 点击部署 → 查看日志
   ```

2. 检查常见错误：
   ```
   ❌ panic: runtime error
   ❌ connection refused
   ❌ environment variable not set
   ❌ permission denied
   ```

**解决方案：**

**4.1 环境变量缺失**
```
Error: required environment variable not set
```
- 检查Railway控制台Variables标签页
- 确保所有必需变量已配置：
  - `BINANCE_API_KEY`
  - `BINANCE_SECRET_KEY`
  - `JWT_SECRET`
  - `QWEN_API_KEY` 或 `DEEPSEEK_API_KEY`

**4.2 端口配置错误**
```
Error: bind: address already in use
```
- 确保使用Railway提供的PORT环境变量
- 检查代码中端口配置逻辑

**4.3 权限问题**
```
Error: permission denied
```
- 检查Dockerfile中的用户权限设置
- 确保应用有读写必要目录的权限

### 🔌 API连接问题

#### 问题5：币安API连接失败
**症状：**
```
Error: Binance API authentication failed
```

**解决方案：**
1. 验证API密钥：
   - 检查密钥格式（无多余空格）
   - 确认密钥权限设置
   - 验证IP白名单设置

2. 测试API连接：
   ```bash
   curl -H "X-MBX-APIKEY: your_api_key" \
        "https://api.binance.com/api/v3/account"
   ```

3. 检查网络连接：
   - Railway服务器IP可能需要加入白名单
   - 检查防火墙设置

#### 问题6：AI模型API连接失败
**症状：**
```
Error: AI API request failed
```

**解决方案：**
1. 验证API密钥有效性
2. 检查API配额和限制
3. 验证API端点URL正确性
4. 检查请求格式和参数

### 🌐 网络访问问题

#### 问题7：健康检查失败
**症状：**
```
Health check failed: connection timeout
```

**解决方案：**
1. 检查健康检查端点：
   ```go
   // 确保/health端点存在且正常响应
   router.GET("/health", func(c *gin.Context) {
       c.JSON(200, gin.H{"status": "ok"})
   })
   ```

2. 验证端口配置：
   ```go
   port := os.Getenv("PORT")
   if port == "" {
       port = "8080"
   }
   ```

3. 检查防火墙和安全组设置

#### 问题8：API端点无法访问
**症状：**
- 404 Not Found
- 500 Internal Server Error
- Connection refused

**解决方案：**
1. 检查路由配置
2. 验证中间件设置
3. 检查CORS配置
4. 查看详细错误日志

### 💾 数据库连接问题

#### 问题9：数据库连接失败
**症状：**
```
Error: failed to connect to database
```

**解决方案：**
1. 检查DATABASE_URL环境变量
2. 验证数据库服务状态
3. 检查连接池配置
4. 验证数据库权限

### 🔧 性能问题

#### 问题10：内存使用过高
**症状：**
- 应用频繁重启
- 响应缓慢
- 内存使用率 > 90%

**解决方案：**
1. 检查内存泄漏：
   ```go
   // 使用pprof进行内存分析
   import _ "net/http/pprof"
   ```

2. 优化数据结构和算法
3. 增加Railway服务规格
4. 实现数据缓存策略

#### 问题11：CPU使用过高
**症状：**
- 响应时间长
- CPU使用率持续 > 80%

**解决方案：**
1. 分析CPU热点
2. 优化算法复杂度
3. 实现异步处理
4. 增加服务实例

## 🛠️ 诊断工具

### Railway控制台工具
1. **日志查看：** Deployments → 选择部署 → 查看日志
2. **指标监控：** Metrics标签页
3. **环境变量：** Variables标签页
4. **域名设置：** Settings → Domains

### 本地诊断工具
```bash
# 1. 本地构建测试
docker build -t nofx-test .
docker run -p 8080:8080 nofx-test

# 2. 健康检查测试
curl http://localhost:8080/health

# 3. API端点测试
curl http://localhost:8080/api/traders

# 4. 压力测试
ab -n 100 -c 10 http://localhost:8080/health
```

### 日志分析
```bash
# 在Railway控制台查看实时日志
# 关注以下关键词：
- ERROR
- FATAL
- panic
- connection failed
- authentication failed
```

## 📞 获取帮助

### 1. 检查文档
- **Railway文档：** https://docs.railway.app/
- **项目README：** 查看项目根目录
- **API文档：** 查看相关API官方文档

### 2. 社区支持
- **Railway Discord：** https://discord.gg/railway
- **GitHub Issues：** 项目仓库issues页面
- **Stack Overflow：** 搜索相关问题

### 3. 联系支持
- **Railway支持：** help@railway.app
- **项目维护者：** 查看项目联系方式

## 🔄 故障排除流程

### 标准排查流程
1. **确定问题阶段：**
   - [ ] 构建阶段
   - [ ] 启动阶段
   - [ ] 运行阶段

2. **收集信息：**
   - [ ] 查看Railway日志
   - [ ] 检查环境变量配置
   - [ ] 验证API密钥
   - [ ] 测试网络连接

3. **逐步排查：**
   - [ ] 本地复现问题
   - [ ] 检查配置文件
   - [ ] 验证依赖版本
   - [ ] 测试API连接

4. **应用解决方案：**
   - [ ] 修复配置问题
   - [ ] 更新代码
   - [ ] 重新部署
   - [ ] 验证修复效果

### 紧急恢复流程
如果生产环境出现严重问题：

1. **立即回滚：**
   ```
   Railway控制台 → Deployments → 选择上一个稳定版本 → Redeploy
   ```

2. **临时修复：**
   - 修改环境变量
   - 重启服务
   - 切换到备用配置

3. **根本修复：**
   - 修复代码问题
   - 测试验证
   - 重新部署

## ✅ 预防措施

### 部署前检查
- [ ] 本地完整测试
- [ ] 环境变量验证
- [ ] API连接测试
- [ ] 性能基准测试

### 监控设置
- [ ] 设置告警阈值
- [ ] 配置日志监控
- [ ] 实现健康检查
- [ ] 定期备份数据

### 最佳实践
- [ ] 使用测试网进行初始部署
- [ ] 实施渐进式部署
- [ ] 保持文档更新
- [ ] 定期安全审计

---

**记住：大多数问题都可以通过仔细检查配置和日志来解决！** 🔧

**项目信息：**
- 项目ID: `d9845ff4-c4a3-4c5d-8e9f-db95151d21bc`
- 支持文档: 查看项目根目录其他文档