#!/bin/bash
# =============================================
# ghcr.io Docker 镜像加速拉取脚本（南京大学镜像）
# 使用方法: ./ghcr-pull.sh <镜像> [镜像2 ...]
# =============================================

#MIRROR_REGISTRY="ghcr.nju.edu.cn"   # 你可以改成其他加速源
MIRROR_REGISTRY="ghcr.1ms.run"
ORIGINAL_REGISTRY="ghcr.io"

if [ $# -eq 0 ]; then
  echo "❌ 用法错误"
  echo "   $0 ghcr.io/owner/image:tag"
  echo "   $0 ghcr.io/owner/image:tag ghcr.io/another/image"
  exit 1
fi

for original_img in "$@"; do
  if [[ "$original_img" == ${ORIGINAL_REGISTRY}/* ]]; then
    # 自动替换前缀（支持带 tag 或 digest）
    mirror_img="${original_img/${ORIGINAL_REGISTRY}/${MIRROR_REGISTRY}}"
    
    echo "🚀 正在从南京大学镜像加速拉取: $mirror_img"
    
    if docker pull "$mirror_img"; then
      echo "🏷️  打标签为原始名称: $original_img"
      docker tag "$mirror_img" "$original_img"
      
      echo "🧹 清理临时加速标签..."
      docker rmi "$mirror_img" >/dev/null 2>&1 || true
      
      echo "✅ 成功！镜像已就绪 → $original_img"
    else
      echo "❌ 拉取失败，请检查网络或镜像是否存在: $mirror_img"
    fi
  else
    echo "ℹ️  非 ghcr.io 镜像，直接拉取: $original_img"
    docker pull "$original_img"
  fi
done

echo "🎉 全部处理完成！以后直接用这个脚本拉 ghcr.io 镜像就行了～"
