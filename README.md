# family（家谱）

Go（Gin）+ SQLite 后端，Vue 3 + Vite 前端。管理家族、成员与亲属关系，支持公历/农历生日与企业微信生日提醒。

## 结构

- `server/`：API、定时生日检查、静态托管 `web/dist`
- `web/`：前端源码
- `seed.js`：Node 演示数据脚本（生成项目根目录 `family.db`）

## 环境变量（服务端）


| 变量            | 说明                  | 默认                  |
| ------------- | ------------------- | ------------------- |
| `DB_PATH`     | SQLite 文件路径         | `./family.db`       |
| `SERVER_ADDR` | 监听地址                | `:8080`             |
| `WEBHOOK_URL` | 企业微信 webhook 基础 URL | 官方默认地址              |
| `REMIND_TIME` | 生日检查 cron 表达式       | `0 8 * * *`（每天 8 点） |


家族表中的 `webhook_key` 与 `WEBHOOK_URL` 组合用于推送。

## 后端

```bash
cd server
go run .
```

可选：仅执行一次生日检查后退出

```bash
go run . -check-birthday
```

## 前端开发

```bash
cd web
npm install
npm run dev
```

开发时 Vite 将 `/api` 代理到 `http://localhost:8080`。

## 前端构建与一体部署

```bash
cd web && npm run build
cd ../server && go run .
```

存在 `web/dist` 时，后端会托管前端并做 SPA 回退。

## 演示数据

`database/seed.go`

若数据库在仓库根目录而你在 `server` 下启动服务，请设置 `DB_PATH` 指向该文件。

## API 概要

前缀均为 `/api`。

- 家族：`/families`（CRUD）
- 成员：`/persons`（CRUD）、`/birthdays/today`、`/birthdays/upcoming`
- 关系：`/relations`、`/persons/:id/relations`、`/families/:id/relations`、`/relation-types`
- 健康：`GET /health`
- 调试：`POST /api/debug/check-birthday`（手动触发生日检查）
