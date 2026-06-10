VERSION ?= 0.1.0
RELEASE_LDFLAGS := -s -w -X github.com/yu/openluckin/internal/cli.version=$(VERSION)
DL_DIR := web/public/dl
SKILL_ZIP := web/public/openluckin-order.zip

.PHONY: build snapshot generate check release skill-assets

build:
	go build -o bin/openluckin ./cmd/openluckin

# 第1步：从官方 MCP 拉取工具清单快照（需要 token）
snapshot:
	go run ./cmd/openluckin tools > schema/tools.json

# 第2步：从快照生成子命令表 + SKILL.md
generate:
	go generate ./...

check:
	go build ./... && go vet ./... && go test ./...

# 把 skill 打包成 zip 放进网站静态目录，落地页提供下载
skill-assets:
	rm -f $(SKILL_ZIP)
	cd skill && zip -qr ../$(SKILL_ZIP) openluckin-order

# 发布构建：五平台交叉编译（strip 调试信息）输出到网站静态目录 + sha256 校验和
release: skill-assets
	mkdir -p $(DL_DIR)
	GOOS=darwin  GOARCH=arm64 go build -trimpath -ldflags="$(RELEASE_LDFLAGS)" -o $(DL_DIR)/openluckin-darwin-arm64      ./cmd/openluckin
	GOOS=darwin  GOARCH=amd64 go build -trimpath -ldflags="$(RELEASE_LDFLAGS)" -o $(DL_DIR)/openluckin-darwin-amd64      ./cmd/openluckin
	GOOS=linux   GOARCH=arm64 go build -trimpath -ldflags="$(RELEASE_LDFLAGS)" -o $(DL_DIR)/openluckin-linux-arm64       ./cmd/openluckin
	GOOS=linux   GOARCH=amd64 go build -trimpath -ldflags="$(RELEASE_LDFLAGS)" -o $(DL_DIR)/openluckin-linux-amd64       ./cmd/openluckin
	GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="$(RELEASE_LDFLAGS)" -o $(DL_DIR)/openluckin-windows-amd64.exe ./cmd/openluckin
	cd $(DL_DIR) && shasum -a 256 openluckin-* > checksums.txt
	@echo "✓ release $(VERSION) -> $(DL_DIR)"
