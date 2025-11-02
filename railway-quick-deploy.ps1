# NOFX AI交易系统 - Railway快速部署脚本
# 项目ID: d9845ff4-c4a3-4c5d-8e9f-db95151d21bc

param(
    [switch]$CheckOnly,
    [switch]$SkipValidation
)

Write-Host "🚀 NOFX AI交易系统 - Railway快速部署" -ForegroundColor Green
Write-Host "项目ID: d9845ff4-c4a3-4c5d-8e9f-db95151d21bc" -ForegroundColor Cyan
Write-Host ""

# 检查Railway CLI
Write-Host "🔍 检查Railway CLI..." -ForegroundColor Yellow
try {
    $railwayVersion = railway --version 2>$null
    Write-Host "✅ Railway CLI已安装: $railwayVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Railway CLI未安装" -ForegroundColor Red
    Write-Host "请运行: npm install -g @railway/cli" -ForegroundColor Yellow
    exit 1
}

# 检查登录状态
Write-Host "🔐 检查Railway登录状态..." -ForegroundColor Yellow
try {
    $user = railway whoami 2>$null
    Write-Host "✅ 已登录Railway: $user" -ForegroundColor Green
} catch {
    Write-Host "❌ 未登录Railway" -ForegroundColor Red
    Write-Host "请运行: railway login" -ForegroundColor Yellow
    exit 1
}

# 检查必需文件
Write-Host "📁 检查部署文件..." -ForegroundColor Yellow
$requiredFiles = @(
    "railway.json",
    ".env.railway", 
    "Dockerfile",
    "go.mod",
    "go.sum",
    "main.go"
)

$missingFiles = @()
foreach ($file in $requiredFiles) {
    if (Test-Path $file) {
        Write-Host "✅ $file" -ForegroundColor Green
    } else {
        Write-Host "❌ $file" -ForegroundColor Red
        $missingFiles += $file
    }
}

if ($missingFiles.Count -gt 0) {
    Write-Host "❌ 缺少必需文件，无法继续部署" -ForegroundColor Red
    exit 1
}

# 如果只是检查，到此结束
if ($CheckOnly) {
    Write-Host "✅ 所有检查通过，可以开始部署" -ForegroundColor Green
    exit 0
}

# 连接到Railway项目
Write-Host "🔗 连接到Railway项目..." -ForegroundColor Yellow
try {
    railway link d9845ff4-c4a3-4c5d-8e9f-db95151d21bc
    Write-Host "✅ 项目连接成功" -ForegroundColor Green
} catch {
    Write-Host "❌ 项目连接失败" -ForegroundColor Red
    Write-Host "请检查项目ID是否正确" -ForegroundColor Yellow
    exit 1
}

# 环境变量提醒
Write-Host ""
Write-Host "⚠️  重要提醒：请确保在Railway控制台中配置了以下必需的环境变量：" -ForegroundColor Yellow
Write-Host "   🔑 BINANCE_API_KEY" -ForegroundColor White
Write-Host "   🔑 BINANCE_SECRET_KEY" -ForegroundColor White
Write-Host "   🔑 JWT_SECRET" -ForegroundColor White
Write-Host "   🤖 QWEN_API_KEY 或 DEEPSEEK_API_KEY" -ForegroundColor White
Write-Host ""

if (-not $SkipValidation) {
    $continue = Read-Host "是否已配置所有必需的环境变量？(y/N)"
    if ($continue -ne "y" -and $continue -ne "Y") {
        Write-Host "请先在Railway控制台配置环境变量，然后重新运行此脚本" -ForegroundColor Yellow
        Write-Host "Railway控制台: https://railway.app/project/d9845ff4-c4a3-4c5d-8e9f-db95151d21bc" -ForegroundColor Cyan
        exit 0
    }
}

# 开始部署
Write-Host "📦 开始部署到Railway..." -ForegroundColor Yellow
try {
    railway up --detach
    Write-Host "✅ 部署命令已执行" -ForegroundColor Green
} catch {
    Write-Host "❌ 部署失败" -ForegroundColor Red
    Write-Host "请检查Railway控制台日志获取详细信息" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "🎉 部署已启动！" -ForegroundColor Green
Write-Host ""
Write-Host "📋 下一步操作：" -ForegroundColor Cyan
Write-Host "1. 在Railway控制台查看部署进度" -ForegroundColor White
Write-Host "2. 等待构建完成（通常需要3-5分钟）" -ForegroundColor White
Write-Host "3. 获取应用URL并测试健康检查" -ForegroundColor White
Write-Host "4. 验证API端点功能" -ForegroundColor White
Write-Host ""
Write-Host "🌐 Railway控制台: https://railway.app/project/d9845ff4-c4a3-4c5d-8e9f-db95151d21bc" -ForegroundColor Cyan
Write-Host ""

# 尝试获取域名
Write-Host "🔍 尝试获取应用域名..." -ForegroundColor Yellow
try {
    $domain = railway domain 2>$null
    if ($domain) {
        Write-Host "🎯 应用URL: https://$domain" -ForegroundColor Green
        Write-Host "🏥 健康检查: https://$domain/health" -ForegroundColor Green
    } else {
        Write-Host "ℹ️  域名将在部署完成后可用" -ForegroundColor Yellow
    }
} catch {
    Write-Host "ℹ️  域名将在部署完成后可用" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "📚 相关文档：" -ForegroundColor Cyan
Write-Host "   📖 部署指南: RAILWAY_DEPLOY_STEPS.md" -ForegroundColor White
Write-Host "   ✅ 验证指南: RAILWAY_VERIFICATION_GUIDE.md" -ForegroundColor White
Write-Host "   🚨 故障排除: RAILWAY_TROUBLESHOOTING.md" -ForegroundColor White
Write-Host "   🔑 环境变量: RAILWAY_ENV_CHECKLIST.md" -ForegroundColor White

Write-Host ""
Write-Host "🆘 如需帮助，请查看相关文档或联系支持" -ForegroundColor Gray
Write-Host ""
Write-Host "按任意键退出..." -ForegroundColor Gray
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")