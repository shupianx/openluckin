#!/usr/bin/env bash
# openluckin 安装脚本（macOS / Linux）
# 用法: curl -fsSL https://<域名>/install.sh | bash
set -euo pipefail

# 安装服务域名
BASE_URL="${OPENLUCKIN_BASE_URL:-https://openluckin.com}"

INSTALL_DIR="$HOME/.openluckin/bin"
BIN_PATH="$INSTALL_DIR/openluckin"

os="$(uname -s)"
arch="$(uname -m)"

case "$os" in
  Darwin) os=darwin ;;
  Linux)  os=linux ;;
  *)
    echo "✗ 不支持的系统: ${os}（Windows 请用: irm ${BASE_URL}/install.ps1 | iex）" >&2
    exit 1
    ;;
esac

case "$arch" in
  arm64|aarch64) arch=arm64 ;;
  x86_64|amd64)  arch=amd64 ;;
  *) echo "✗ 不支持的架构: $arch" >&2; exit 1 ;;
esac

asset="openluckin-$os-$arch"
url="$BASE_URL/dl/$asset"

echo "▸ 下载 $url"
mkdir -p "$INSTALL_DIR"
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT
curl -fSL --progress-bar "$url" -o "$tmp"

# sha256 校验（机器上没有校验工具时跳过）
expected="$(curl -fsSL "$BASE_URL/dl/checksums.txt" | awk -v a="$asset" '$2==a {print $1}' || true)"
if [ -n "$expected" ]; then
  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "$tmp" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    actual="$(shasum -a 256 "$tmp" | awk '{print $1}')"
  else
    actual=""
  fi
  if [ -n "$actual" ] && [ "$expected" != "$actual" ]; then
    echo "✗ sha256 校验失败：文件可能不完整或被篡改" >&2
    exit 1
  fi
  [ -n "$actual" ] && echo "✓ sha256 校验通过"
fi

mv "$tmp" "$BIN_PATH"
trap - EXIT
chmod +x "$BIN_PATH"
echo "✓ 已安装到 $BIN_PATH"

# 加入 PATH 方便手动使用（AI agent 调用走绝对路径，不依赖这一步）
shell_rc=""
case "${SHELL:-}" in
  */zsh)  shell_rc="$HOME/.zshrc" ;;
  */bash) shell_rc="$HOME/.bashrc" ;;
esac
if [ -n "$shell_rc" ] && ! grep -qs '\.openluckin/bin' "$shell_rc"; then
  printf '\nexport PATH="$HOME/.openluckin/bin:$PATH"\n' >> "$shell_rc"
  echo "✓ 已将 ~/.openluckin/bin 加入 PATH（${shell_rc}），重开终端生效"
fi

echo
echo "下一步：浏览器登录瑞幸账号获取 token"
echo "  \"$BIN_PATH\" login"
