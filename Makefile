.PHONY: help dist dist-linux-arm64 dist-linux-amd64 dist-linux clean test-docker

help:
	@echo "make dist              # 构建 Linux arm64 + amd64 到 dist/"
	@echo "make test-docker         # 构建后启动 debian 容器，把对应架构二进制挂到 /config-cli 并执行"
	@echo "  TARGET_ARCH=arm64|amd64 make test-docker  # 强制选择架构（可选）"
	@echo "make dist-linux-arm64"
	@echo "make dist-linux-amd64"
	@echo "make clean             # 删除 dist/"

DIST_DIR := dist
APP := config-cli
VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)

dist-linux-arm64:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP)-linux-arm64 .

dist-linux-amd64:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP)-linux-amd64 .

dist: dist-linux-arm64 dist-linux-amd64
	@ls -lh $(DIST_DIR) | sed -n '1,5p' || true

dist-linux: dist

clean:
	rm -rf $(DIST_DIR)

test-docker: dist
	bash ./scripts/docker-test.sh
