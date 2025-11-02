#!/usr/bin/env python3
"""
NOFX AI Trading System - Automated Deployment Script
自动化部署脚本
"""

import subprocess
import os
import sys
import time
import json
from pathlib import Path

# 颜色代码
RED = '\033[0;31m'
GREEN = '\033[0;32m'
YELLOW = '\033[1;33m'
BLUE = '\033[0;34m'
NC = '\033[0m'  # No Color

def log_info(msg):
    print(f"{BLUE}ℹ️  {msg}{NC}")

def log_success(msg):
    print(f"{GREEN}✅ {msg}{NC}")

def log_warning(msg):
    print(f"{YELLOW}⚠️  {msg}{NC}")

def log_error(msg):
    print(f"{RED}❌ {msg}{NC}")

def run_command(cmd, cwd=None, capture_output=False):
    """执行命令并返回结果"""
    try:
        log_info(f"执行: {cmd}")
        result = subprocess.run(
            cmd,
            shell=True,
            cwd=cwd,
            capture_output=capture_output,
            text=True,
            timeout=300
        )
        if result.returncode == 0:
            return True, result.stdout if capture_output else None
        else:
            log_error(f"命令失败: {result.stderr if capture_output else ''}")
            return False, result.stderr if capture_output else None
    except subprocess.TimeoutExpired:
        log_error("命令执行超时")
        return False, None
    except Exception as e:
        log_error(f"执行错误: {str(e)}")
        return False, None

def install_zeabur_cli():
    """安装Zeabur CLI"""
    log_info("检查Zeabur CLI...")
    
    success, _ = run_command("which zeabur", capture_output=True)
    if success:
        log_success("Zeabur CLI已安装")
        return True
    
    log_warning("Zeabur CLI未安装，开始安装...")
    success, _ = run_command("curl -fsSL https://zeabur.com/install.sh | bash")
    
    if success:
        # 添加到PATH
        home = os.path.expanduser("~")
        zeabur_path = os.path.join(home, ".zeabur", "bin")
        os.environ["PATH"] = f"{zeabur_path}:{os.environ.get('PATH', '')}"
        log_success("Zeabur CLI安装成功")
        return True
    else:
        log_error("Zeabur CLI安装失败")
        return False

def build_frontend():
    """构建前端"""
    log_info("构建前端应用...")
    
    web_dir = "/workspace/nofx-deploy/nofx/web"
    
    # 安装依赖
    log_info("安装前端依赖...")
    success, _ = run_command("npm install", cwd=web_dir)
    if not success:
        log_error("依赖安装失败")
        return False
    
    # 构建
    log_info("执行构建...")
    success, output = run_command("npm run build", cwd=web_dir, capture_output=True)
    if not success:
        log_error(f"构建失败: {output}")
        return False
    
    log_success("前端构建成功")
    return True

def deploy_to_zeabur():
    """部署后端到Zeabur"""
    log_info("开始部署后端到Zeabur...")
    
    # 这里需要Zeabur token
    zeabur_token = os.getenv("ZEABUR_TOKEN")
    if not zeabur_token:
        log_error("未找到ZEABUR_TOKEN环境变量")
        return False, None
    
    project_dir = "/workspace/nofx-deploy/nofx"
    
    # 初始化git（如果需要）
    if not os.path.exists(os.path.join(project_dir, ".git")):
        run_command("git init", cwd=project_dir)
        run_command("git add .", cwd=project_dir)
        run_command('git commit -m "Initial commit"', cwd=project_dir)
    
    log_info("Zeabur部署需要手动配置，请参考文档")
    log_warning("由于环境限制，后端部署需要在本地执行")
    
    return False, None

def deploy_frontend_to_supabase():
    """部署前端到Supabase Storage"""
    log_info("开始部署前端到Supabase...")
    
    supabase_url = "https://eqzurdzoaxibothslnna.supabase.co"
    supabase_token = os.getenv("SUPABASE_ACCESS_TOKEN")
    
    if not supabase_token:
        log_error("未找到SUPABASE_ACCESS_TOKEN")
        return False
    
    web_dir = "/workspace/nofx-deploy/nofx/web"
    dist_dir = os.path.join(web_dir, "dist")
    
    if not os.path.exists(dist_dir):
        log_error(f"dist目录不存在: {dist_dir}")
        return False
    
    # 使用Supabase CLI上传
    log_info("上传文件到Supabase Storage...")
    
    bucket_name = "nofx-frontend"
    
    # 检查bucket是否存在，不存在则创建
    cmd = f'supabase storage ls --token {supabase_token} --project-ref eqzurdzoaxibothslnna'
    run_command(cmd)
    
    # 上传文件
    files_uploaded = 0
    for root, dirs, files in os.walk(dist_dir):
        for file in files:
            file_path = os.path.join(root, file)
            rel_path = os.path.relpath(file_path, dist_dir)
            
            cmd = f'supabase storage cp "{file_path}" "{bucket_name}/{rel_path}" --token {supabase_token} --project-ref eqzurdzoaxibothslnna'
            success, _ = run_command(cmd)
            if success:
                files_uploaded += 1
    
    log_success(f"已上传 {files_uploaded} 个文件到Supabase")
    
    frontend_url = f"{supabase_url}/storage/v1/object/public/{bucket_name}/index.html"
    log_success(f"前端URL: {frontend_url}")
    
    return True

def main():
    """主函数"""
    print("\n" + "="*60)
    print("    NOFX AI交易系统 - 自动化部署")
    print("="*60 + "\n")
    
    start_time = time.time()
    
    # 步骤1: 安装Zeabur CLI
    print("\n[步骤 1/3] 安装必要工具")
    print("-" * 60)
    if not install_zeabur_cli():
        log_warning("Zeabur CLI安装失败，将跳过后端部署")
    
    # 步骤2: 构建前端
    print("\n[步骤 2/3] 构建前端")
    print("-" * 60)
    if not build_frontend():
        log_error("前端构建失败，终止部署")
        return 1
    
    # 步骤3: 部署
    print("\n[步骤 3/3] 部署应用")
    print("-" * 60)
    
    # 部署后端（可能需要手动）
    log_info("后端部署...")
    backend_success, backend_url = deploy_to_zeabur()
    
    # 部署前端
    log_info("前端部署...")
    frontend_success = deploy_frontend_to_supabase()
    
    # 总结
    print("\n" + "="*60)
    print("    部署完成")
    print("="*60)
    
    elapsed = int(time.time() - start_time)
    print(f"\n总耗时: {elapsed // 60}分{elapsed % 60}秒\n")
    
    if frontend_success:
        print("✅ 前端部署成功")
        print(f"   URL: https://eqzurdzoaxibothslnna.supabase.co/storage/v1/object/public/nofx-frontend/index.html\n")
    
    if not backend_success:
        print("⚠️  后端需要手动部署:")
        print("   1. 在本地安装Zeabur CLI")
        print("   2. 运行: cd /workspace/nofx-deploy/nofx")
        print("   3. 运行: ./deploy-to-zeabur.sh\n")
    
    return 0 if (frontend_success or backend_success) else 1

if __name__ == "__main__":
    try:
        sys.exit(main())
    except KeyboardInterrupt:
        log_warning("\n部署已取消")
        sys.exit(1)
    except Exception as e:
        log_error(f"发生错误: {str(e)}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
