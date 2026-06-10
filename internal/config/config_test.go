package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveTokenCreatesFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path, err := SaveToken("tok1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取失败: %v", err)
	}
	if string(data) != EnvToken+"=tok1\n" {
		t.Fatalf("文件内容不符: %q", data)
	}
}

func TestSaveTokenReplacesOldAndKeepsOthers(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ConfigDirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	old := "# 注释保留\n" + EnvTokenAlias + "=oldtok\nOTHER_KEY=keep\n"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(old), 0o600); err != nil {
		t.Fatal(err)
	}

	path, err := SaveToken("newtok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(path)
	s := string(data)
	if strings.Contains(s, "oldtok") {
		t.Fatalf("旧 token 应被替换: %q", s)
	}
	if !strings.Contains(s, EnvToken+"=newtok") {
		t.Fatalf("缺少新 token: %q", s)
	}
	if !strings.Contains(s, "OTHER_KEY=keep") || !strings.Contains(s, "# 注释保留") {
		t.Fatalf("其他配置应保留: %q", s)
	}
}
