#!/usr/bin/env python3
"""
上传前端文件到Supabase Storage
"""

import os
import requests
import mimetypes
from pathlib import Path

# Supabase配置
SUPABASE_URL = "https://eqzurdzoaxibothslnna.supabase.co"
SUPABASE_ACCESS_TOKEN = "sbp_cb3f3a6f373315e288f532e1ede5442ef4fbf311"
SUPABASE_ANON_KEY = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImVxenVyZHpvYXhpYm90aHNsbm5hIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjE4NzY2NjUsImV4cCI6MjA3NzQ1MjY2NX0.h2EQOkofLavh-DL68AGfFX7ZvJ4SipNsiO7K5uTh20Y"
BUCKET_NAME = "nofx-frontend"

def create_bucket():
    """创建Storage桶（如果不存在）"""
    print(f"🔧 检查Storage桶: {BUCKET_NAME}...")
    
    url = f"{SUPABASE_URL}/storage/v1/bucket"
    headers = {
        "Authorization": f"Bearer {SUPABASE_ACCESS_TOKEN}",
        "Content-Type": "application/json",
        "apikey": SUPABASE_ANON_KEY
    }
    data = {
        "id": BUCKET_NAME,
        "name": BUCKET_NAME,
        "public": True
    }
    
    try:
        response = requests.post(url, headers=headers, json=data)
        if response.status_code in [200, 201]:
            print(f"✅ Storage桶创建成功")
        elif response.status_code == 409:
            print(f"ℹ️  Storage桶已存在")
        else:
            print(f"⚠️  桶状态: {response.status_code}")
    except Exception as e:
        print(f"⚠️  创建桶时出错（可能已存在）: {str(e)}")

def delete_file(file_path):
    """删除已存在的文件"""
    url = f"{SUPABASE_URL}/storage/v1/object/{BUCKET_NAME}/{file_path}"
    headers = {
        "Authorization": f"Bearer {SUPABASE_ACCESS_TOKEN}",
        "apikey": SUPABASE_ANON_KEY
    }
    
    try:
        requests.delete(url, headers=headers)
    except:
        pass

def upload_file(local_path, remote_path):
    """上传单个文件到Supabase Storage"""
    # 先尝试删除旧文件
    delete_file(remote_path)
    
    url = f"{SUPABASE_URL}/storage/v1/object/{BUCKET_NAME}/{remote_path}"
    
    # 获取MIME类型
    mime_type, _ = mimetypes.guess_type(local_path)
    if mime_type is None:
        mime_type = "application/octet-stream"
    
    headers = {
        "Authorization": f"Bearer {SUPABASE_ACCESS_TOKEN}",
        "Content-Type": mime_type,
        "apikey": SUPABASE_ANON_KEY
    }
    
    try:
        with open(local_path, 'rb') as f:
            response = requests.post(url, headers=headers, data=f)
            
        if response.status_code in [200, 201]:
            return True, None
        else:
            return False, f"HTTP {response.status_code}: {response.text[:100]}"
    except Exception as e:
        return False, str(e)

def upload_directory(dist_dir):
    """上传整个dist目录"""
    print(f"\n📤 开始上传文件到Supabase Storage...")
    print(f"源目录: {dist_dir}\n")
    
    uploaded = 0
    failed = 0
    
    dist_path = Path(dist_dir)
    
    for file_path in dist_path.rglob('*'):
        if file_path.is_file():
            # 计算相对路径
            relative_path = file_path.relative_to(dist_path)
            remote_path = str(relative_path).replace('\\', '/')
            
            print(f"上传: {remote_path}...", end=" ")
            
            success, error = upload_file(str(file_path), remote_path)
            
            if success:
                print("✅")
                uploaded += 1
            else:
                print(f"❌ {error}")
                failed += 1
    
    return uploaded, failed

def main():
    print("\n" + "="*60)
    print("    NOFX前端 - Supabase Storage部署")
    print("="*60 + "\n")
    
    # 创建桶
    create_bucket()
    
    # 上传文件
    dist_dir = "/workspace/nofx-deploy/nofx/web/dist"
    
    if not os.path.exists(dist_dir):
        print(f"❌ dist目录不存在: {dist_dir}")
        return 1
    
    uploaded, failed = upload_directory(dist_dir)
    
    # 结果
    print("\n" + "="*60)
    print(f"✅ 上传完成: {uploaded} 个文件成功")
    if failed > 0:
        print(f"❌ 失败: {failed} 个文件")
    print("="*60 + "\n")
    
    # 显示访问URL
    frontend_url = f"{SUPABASE_URL}/storage/v1/object/public/{BUCKET_NAME}/index.html"
    print("🌐 前端访问URL:")
    print(f"   {frontend_url}\n")
    
    print("📝 提示:")
    print("   - 首次访问可能需要几秒钟加载")
    print("   - 可以在Supabase控制台查看所有文件")
    print("   - 建议配置自定义域名以获得更好的访问体验\n")
    
    return 0 if failed == 0 else 1

if __name__ == "__main__":
    import sys
    sys.exit(main())
