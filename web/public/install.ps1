# openluckin 安装脚本（Windows）
# 用法: irm https://<域名>/install.ps1 | iex
$ErrorActionPreference = "Stop"

# 安装服务域名
$BaseUrl = "https://openluckin.com"

$InstallDir = Join-Path $env:USERPROFILE ".openluckin\bin"
$BinPath = Join-Path $InstallDir "openluckin.exe"
$Asset = "openluckin-windows-amd64.exe"

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

Write-Host "▸ 下载 $BaseUrl/dl/$Asset"
Invoke-WebRequest -Uri "$BaseUrl/dl/$Asset" -OutFile $BinPath

# sha256 校验
try {
  $checksums = (Invoke-WebRequest -Uri "$BaseUrl/dl/checksums.txt").Content
  $expected = ($checksums -split "`n" | Where-Object { $_ -match [regex]::Escape($Asset) }) -split '\s+' | Select-Object -First 1
  if ($expected) {
    $actual = (Get-FileHash -Algorithm SHA256 -Path $BinPath).Hash.ToLower()
    if ($expected.ToLower() -ne $actual) {
      Remove-Item $BinPath
      throw "sha256 校验失败：文件可能不完整或被篡改"
    }
    Write-Host "✓ sha256 校验通过"
  }
} catch [System.Net.WebException] {
  Write-Host "（跳过校验：checksums.txt 不可用）"
}

# 加入用户 PATH 方便手动使用
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$InstallDir*") {
  [Environment]::SetEnvironmentVariable("Path", "$InstallDir;$userPath", "User")
  Write-Host "✓ 已将 $InstallDir 加入用户 PATH，重开终端生效"
}

Write-Host "✓ 已安装到 $BinPath"
Write-Host ""
Write-Host "下一步：浏览器登录瑞幸账号获取 token"
Write-Host "  & `"$BinPath`" login"
