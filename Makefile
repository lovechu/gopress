.PHONY: help swagger-init swagger-gen swagger-serve build run test clean

# 默认目标
help:
	@echo "GoPress - Makefile 命令"
	@echo ""
	@echo "可用命令:"
	@echo "  make swagger-init    - 初始化 Swagger 文档（首次使用）"
	@echo "  make swagger-gen     - 根据代码注释生成/更新 Swagger 文档"
	@echo "  make swagger-serve   - 启动 Swagger UI 文档服务器"
	@echo "  make build           - 编译项目"
	@echo "  make run             - 运行项目"
	@echo "  make test            - 运行测试"
	@echo "  make clean           - 清理编译文件"

# Swagger 相关命令
swagger-init:
	@echo "初始化 Swagger 文档..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@swag init -g cmd/server/main.go -o docs
	@echo "Swagger 文档初始化完成！访问 http://localhost:8080/swagger/index.html 查看"

swagger-gen:
	@echo "生成/更新 Swagger 文档..."
	@swag init -g cmd/server/main.go -o docs
	@echo "Swagger 文档生成完成！"

swagger-serve:
	@echo "启动 Swagger UI 服务器..."
	@swag fmt
	@go run cmd/server/main.go

# 项目构建和运行
build:
	@echo "编译项目..."
	@go build -o bin/gopress cmd/server/main.go
	@echo "编译完成：bin/gopress"

run:
	@echo "运行项目..."
	@go run cmd/server/main.go

test:
	@echo "运行测试..."
	@go test -v ./...

# 代码质量
lint:
	@echo "运行代码检查..."
	@golangci-lint run ./...

fmt:
	@echo "格式化代码..."
	@go fmt ./...
	@goimports -w .

# 清理
clean:
	@echo "清理编译文件..."
	@rm -rf bin/
	@go clean
	@echo "清理完成！"

# 依赖管理
deps:
	@echo "下载依赖..."
	@go mod download
	@echo "整理依赖..."
	@go mod tidy

# 安装 Swagger 工具
install-swagger:
	@echo "安装 Swagger 工具..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Swagger 工具安装完成！"
	@echo "请确保 $(go env GOPATH)/bin 在您的 PATH 中"

# 数据库相关（如果需要）
# migrate-up:
# 	@migrate -path migrations -database "$(DATABASE_URL)" up

# migrate-down:
# 	@migrate -path migrations -database "$(DATABASE_URL)" down

# migrate-create:
# 	@migrate create -ext sql -dir migrations -seq $(name)
