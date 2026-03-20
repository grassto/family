.PHONY: build run clean tidy web-dev web-build test-birthday all

# Go
build:
	cd server && go build -o ../family-tree .

run: build
	./family-tree

tidy:
	cd server && go mod tidy

clean:
	rm -f family-tree family.db

# 前端
web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

# 全量构建（前端 + 后端）
all: web-build build

# 测试
test-birthday:
	cd server && go run . --check-birthday
