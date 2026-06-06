.PHONY: setup build test lint fmt vet security run clean check help

APP_NAME := gin-app
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR := bin
MAIN_PATH := ./cmd/server

check-prerequisites:
	@echo "检查前置条件..."
	@command -v go >/dev/null 2>&1 || { echo "❌ Go 未安装，请先安装 Go"; exit 1; }
	@command -v make >/dev/null 2>&1 || { echo "❌ make 未安装，请先安装 make"; exit 1; }
	@command -v git >/dev/null 2>&1 || { echo "❌ Git 未安装，请先安装 Git"; exit 1; }
	@command -v uv >/dev/null 2>&1 || { echo "❌ uv 未安装，请先安装 uv"; exit 1; }
	@echo "✅ 所有前置条件已安装"
	@echo ""

setup: check-prerequisites
	@echo "正在配置开发环境..."
	@echo ""
	@echo "安装 Go 工具..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "✅ Go 工具安装完成"
	@echo ""
	@echo "安装 pre-commit..."
	@uv tool install pre-commit
	@pre-commit install
	@echo "✅ pre-commit 安装并配置完成"
	@echo ""
	@echo "下载 Go 模块..."
	@go mod download
	@go mod verify
	@echo "✅ Go 模块下载完成"
	@echo ""
	@echo "=========================================="
	@echo "✅ 开发环境配置完成！"
	@echo "=========================================="
	@echo ""
	@echo "已安装工具："
	@echo "  - Go: $(shell go version | cut -d' ' -f3)"
	@echo "  - golangci-lint: $(shell golangci-lint version --short 2>/dev/null || echo '未安装')"
	@echo "  - goimports: 已安装"
	@echo "  - gosec: $(shell gosec -version 2>/dev/null || echo '未安装')"
	@echo "  - pre-commit: $(shell pre-commit --version 2>/dev/null || echo '未安装')"
	@echo ""
	@echo "常用命令："
	@echo "  make run      - 运行应用"
	@echo "  make build    - 构建二进制文件"
	@echo "  make test     - 运行测试"
	@echo "  make check    - 运行所有检查"
	@echo "  make help     - 显示所有命令"

build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/server $(MAIN_PATH)
	@echo "✅ 构建完成: $(BUILD_DIR)/server"

build-all:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/server-linux-amd64 $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/server-windows-amd64.exe $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/server-darwin-amd64 $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/server-darwin-arm64 $(MAIN_PATH)
	@echo "✅ 所有平台构建完成"

test:
	go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -1

test-coverage: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ 覆盖率报告: coverage.html"

lint:
	golangci-lint run --timeout=5m

fmt:
	gofmt -w .
	goimports -w .
	@echo "✅ 代码格式化完成"

vet:
	go vet ./...

security:
	gosec ./...

run:
	go run $(MAIN_PATH)

clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "✅ 清理完成"

check: fmt vet lint test
	@echo "✅ 所有检查通过！"

tidy:
	go mod tidy
	go mod verify
	@echo "✅ 依赖整理完成"

pre-commit-run:
	pre-commit run --all-files

help:
	@echo "用法: make <目标>"
	@echo ""
	@echo "环境配置:"
	@echo "  setup              安装所有开发工具（先检查前置条件）"
	@echo "  check-prerequisites 检查前置条件是否已安装"
	@echo "  tidy               整理 Go 模块"
	@echo ""
	@echo "构建:"
	@echo "  build              构建当前平台"
	@echo "  build-all          构建所有平台 (Linux/Windows/macOS)"
	@echo ""
	@echo "测试:"
	@echo "  test               运行测试并生成覆盖率"
	@echo "  test-coverage      生成 HTML 覆盖率报告"
	@echo ""
	@echo "代码质量:"
	@echo "  lint               运行 golangci-lint"
	@echo "  fmt                格式化代码 (gofmt + goimports)"
	@echo "  vet                运行 go vet"
	@echo "  security           运行 gosec 安全扫描"
	@echo "  check              运行 fmt, vet, lint, test"
	@echo "  pre-commit-run     手动运行 pre-commit 检查"
	@echo ""
	@echo "运行:"
	@echo "  run                运行应用"
	@echo "  clean              清理构建产物"
