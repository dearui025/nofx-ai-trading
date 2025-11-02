# NOFX AI交易系统 - Railway部署脚本 (PowerShell版本)
# 项目ID: d9845ff4-c4a3-4c5d-8e9f-db95151d21bc

Write-Host "🚀 开始部署NOFX AI交易系统到Railway..." -ForegroundColor Green

# 检查Railway CLI是否安装
try {
    railway --version | Out-Null
    Write-Host "✅ Railway CLI已安装" -ForegroundColor Green
} catch {
    Write-Host "❌ Railway CLI未安装，请先安装：" -ForegroundColor Red
    Write-Host "npm install -g @railway/cli" -ForegroundColor Yellow
    exit 1
}

# 检查登录状态
Write-Host "🔐 检查Railway登录状态..." -ForegroundColor Cyan
try {
    railway whoami | Out-Null
    Write-Host "✅ 已登录Railway" -ForegroundColor Green
} catch {
    Write-Host "请先登录Railway：" -ForegroundColor Yellow
    railway login
}

# 连接到指定项目
Write-Host "🔗 连接到Railway项目..." -ForegroundColor Cyan
railway link d9845ff4-c4a3-4c5d-8e9f-db95151d21bc

# 检查环境变量文件
if (Test-Path ".env.railway") {
    Write-Host "⚙️ 找到.env.railway文件" -ForegroundColor Green
    Write-Host "请手动在Railway控制台配置环境变量，或使用Railway CLI上传" -ForegroundColor Yellow
} else {
    Write-Host "⚠️ 未找到.env.railway文件，请手动在Railway控制台配置环境变量" -ForegroundColor Yellow
}

# 部署应用
Write-Host "📦 开始部署..." -ForegroundColor Cyan
railway up --detach

Write-Host "✅ 部署命令已执行！" -ForegroundColor Green
Write-Host "🌐 您可以在Railway控制台查看部署状态：" -ForegroundColor Cyan
Write-Host "   https://railway.app/project/d9845ff4-c4a3-4c5d-8e9f-db95151d21bc" -ForegroundColor Blue

# 尝试获取服务URL
Write-Host "🔍 尝试获取服务URL..." -ForegroundColor Cyan
try {
    $serviceUrl = railway domain 2>$null
    if ($serviceUrl) {
        Write-Host "🎉 您的应用已部署到: $serviceUrl" -ForegroundColor Green
        Write-Host "🏥 健康检查: $serviceUrl/health" -ForegroundColor Green
    } else {
        Write-Host "ℹ️ 服务URL将在部署完成后可用" -ForegroundColor Yellow
    }
} catch {
    Write-Host "ℹ️ 服务URL将在部署完成后可用" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "📋 下一步：" -ForegroundColor Cyan
Write-Host "1. 在Railway控制台配置必需的环境变量（API密钥等）" -ForegroundColor White
Write-Host "2. 等待构建完成（通常需要2-5分钟）" -ForegroundColor White
Write-Host "3. 访问您的应用URL进行测试" -ForegroundColor White
Write-Host ""
Write-Host "🆘 如需帮助，请查看 railway-deploy.md 文档" -ForegroundColor Yellow

# 暂停以便用户查看输出
Write-Host ""
Write-Host "按任意键继续..." -ForegroundColor Gray
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")