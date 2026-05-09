# GoPress

一个基于 Go 语言开发的现代化 CMS（内容管理系统），采用简洁的分层架构设计，支持主题系统、短代码扩展、媒体管理等功能。

## 特性列表

- **用户与权限管理**：支持多角色（Admin、Editor、Author、Subscriber），基于 JWT 的认证系统
- **内容管理**：文章和页面管理，支持草稿/发布状态，版本历史追踪
- **分类法系统**：灵活的分类（Category）和标签（Tag）管理，支持层级结构
- **媒体管理**：支持本地上传和 MinIO 对象存储，自动生成缩略图
- **主题系统**：基于 Go `html/template` 的模板引擎，支持模板继承和部分模板
- **短代码**：可扩展的短代码系统，内置图库、视频、按钮、提示框等组件
- **中间件**：CORS、请求限流、日志记录、请求追踪、Recovery 等开箱即用的中间件
- **配置管理**：基于 Viper 的配置系统，支持 YAML 配置和环境变量覆盖
- **数据库**：使用 GORM 作为 ORM，支持 MySQL，自动迁移

---

## 快速开始

### 环境要求

- Go 1.22+
- MySQL 5.7+ / 8.0+
- Redis 6.0+（可选，用于缓存）
- Git

### 安装步骤

#### 1. 克隆项目

```bash
git clone https://github.com/yourorg/gopress.git
cd gopress
```

#### 2. 安装依赖

```bash
go mod download
```

#### 3. 配置数据库

创建 MySQL 数据库：

```sql
CREATE DATABASE gopress CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

复制并编辑配置文件：

```bash
cp config.yaml config.yaml.local
```

修改 `config.yaml.local` 中的数据库连接信息：

```yaml
database:
  host: "127.0.0.1"
  port: 3306
  name: "gopress"
  user: "your_user"
  password: "your_password"
```

#### 4. 启动服务

```bash
go run ./cmd/server/main.go
```

服务启动后访问 `http://localhost:8080`

### Docker 部署

#### 使用 Docker Compose（推荐）

创建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./config.yaml.local:/app/config.yaml
      - ./storage:/app/storage
    depends_on:
      - mysql
      - redis
    environment:
      - GOPRESS_DATABASE_HOST=mysql
      - GOPRESS_REDIS_ADDR=redis:6379

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: gopress
      MYSQL_USER: gopress
      MYSQL_PASSWORD: gopress123
    volumes:
      - mysql_data:/var/lib/mysql
    ports:
      - "3306:3306"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  mysql_data:
  redis_data:
```

启动服务：

```bash
docker-compose up -d
```

#### Docker 构建

```bash
# 构建镜像
docker build -t gopress:latest .

# 运行容器
docker run -d \
  --name gopress \
  -p 8080:8080 \
  -v $(pwd)/config.yaml.local:/app/config.yaml \
  -v $(pwd)/storage:/app/storage \
  gopress:latest
```

### 手动编译运行

```bash
# 编译
go build -o gopress ./cmd/server

# 运行
./gopress
```

---

## 架构说明

### 目录结构

```
gopress/
├── cmd/                        # 应用程序入口
│   └── server/
│       └── main.go             # 服务启动入口
├── internal/                   # 内部包（不对外暴露）
│   ├── bootstrap/              # 应用初始化和路由注册
│   │   ├── app.go             # 应用初始化
│   │   └── router.go          # 路由配置
│   ├── config/                # 配置管理
│   │   └── config.go          # 配置结构定义
│   ├── user/                  # 用户模块
│   │   ├── model.go           # 数据模型
│   │   ├── repository.go      # 数据访问层
│   │   ├── service.go         # 业务逻辑层
│   │   └── handler.go         # HTTP 处理层
│   ├── content/               # 内容模块
│   │   ├── model.go           # 数据模型
│   │   ├── repository.go      # 数据访问层
│   │   ├── service.go         # 业务逻辑层
│   │   └── handler.go         # HTTP 处理层
│   ├── taxonomy/              # 分类法模块
│   │   ├── model.go           # 数据模型
│   │   ├── repository.go      # 数据访问层
│   │   ├── service.go         # 业务逻辑层
│   │   └── handler.go         # HTTP 处理层
│   ├── media/                 # 媒体模块
│   │   ├── model.go           # 数据模型
│   │   ├── repository.go      # 数据访问层
│   │   ├── service.go         # 业务逻辑层
│   │   ├── handler.go         # HTTP 处理层
│   │   ├── storage.go         # 存储接口
│   │   ├── local.go           # 本地存储实现
│   │   └── minio.go           # MinIO 存储实现
│   ├── theme/                 # 主题模块
│   │   ├── model.go           # 数据模型
│   │   ├── engine.go          # 主题引擎
│   │   ├── service.go         # 业务逻辑层
│   │   ├── handler.go         # HTTP 处理层
│   │   └── shortcode.go       # 短代码解析
│   └── middleware/            # 中间件
│       ├── auth.go            # JWT 认证
│       ├── rbac.go            # 角色权限控制
│       ├── cors.go            # 跨域资源共享
│       ├── logger.go          # 请求日志
│       ├── ratelimit.go       # 限流
│       ├── recover.go         # 异常恢复
│       └── requestid.go       # 请求追踪
├── pkg/                       # 公共工具包
│   ├── database/              # 数据库工具
│   │   ├── mysql.go           # MySQL 连接
│   │   └── redis.go          # Redis 连接
│   ├── jwt/                   # JWT 工具
│   │   └── jwt.go
│   └── response/              # HTTP 响应封装
│       └── response.go
├── themes/                    # 用户主题目录
│   └── default/               # 默认主题
│       ├── theme.yaml         # 主题配置
│       ├── templates/         # 模板文件
│       └── assets/           # 静态资源
├── migrations/               # 数据库迁移脚本
├── docs/                      # 开发文档
├── Dockerfile                # Docker 构建文件
├── go.mod                    # Go 模块定义
├── go.sum                    # 依赖校验
└── config.yaml               # 配置文件示例
```

### 模块说明

#### User 模块 (`internal/user/`)

用户管理模块，处理用户注册、登录、角色权限等功能。

**数据模型**：

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | uint | 主键 |
| Username | string | 用户名（唯一） |
| Email | string | 邮箱（唯一） |
| PasswordHash | string | 密码哈希 |
| DisplayName | string | 显示名称 |
| Avatar | string | 头像 URL |
| Role | string | 角色 |
| Bio | string | 个人简介 |
| IsActive | bool | 是否启用 |
| LastLoginAt | time | 最后登录时间 |

**角色层级**：

```
Admin (4) > Editor (3) > Author (2) > Subscriber (1)
```

#### Content 模块 (`internal/content/`)

内容管理模块，处理文章（Post）和页面（Page）的 CRUD 操作。

**内容类型**：

- `post`：博客文章，支持分类和标签
- `page`：静态页面

**状态**：

- `draft`：草稿
- `published`：已发布
- `trash`：回收站

#### Taxonomy 模块 (`internal/taxonomy/`)

分类法模块，管理分类（Category）和标签（Tag）。

**分类法类型**：

- `category`：分类，支持层级结构
- `tag`：标签，扁平结构

#### Media 模块 (`internal/media/`)

媒体管理模块，处理文件上传、存储和管理。

**存储驱动**：

- `local`：本地文件系统
- `minio`：MinIO 对象存储

**功能**：

- 文件上传
- 缩略图生成
- 元数据管理

#### Theme 模块 (`internal/theme/`)

主题系统，处理模板渲染和短代码解析。

**模板文件**：

- `index.html`：首页模板
- `single.html`：文章详情模板
- `page.html`：页面模板
- `layout/base.html`：基础布局

**内置模板函数**：

- `dateformat`：日期格式化
- `truncate`：字符串截断
- `safeHTML`：标记安全 HTML
- `add`/`sub`：数学运算

#### Middleware 模块 (`internal/middleware/`)

中间件模块，提供各种 HTTP 中间件功能。

| 中间件 | 说明 |
|--------|------|
| JWTAuth | JWT Token 认证 |
| RequireRole | 角色权限校验 |
| CORS | 跨域资源共享 |
| RateLimit | 请求限流 |
| Logger | 请求日志 |
| Recovery | Panic 恢复 |
| RequestID | 请求追踪 |

### 分层架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        HTTP 请求                                 │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Middleware Layer                            │
│  Recovery │ Logger │ CORS │ RateLimit │ RequestID │ Auth │ RBAC  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Handler Layer                             │
│  User │ Content │ Taxonomy │ Media │ Theme                      │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Service Layer                             │
│  UserService │ ContentService │ TaxonomyService │ MediaService  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Repository Layer                            │
│  UserRepo │ ContentRepo │ TaxonomyRepo │ MediaRepo │ ThemeRepo   │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Database Layer                             │
│              MySQL (GORM) │ Redis (Cache)                       │
└─────────────────────────────────────────────────────────────────┘
```

---

## 配置说明

### config.yaml 完整配置示例

```yaml
# GoPress CMS 配置文件
# 复制此文件为 config.yaml 并填入实际值

# 应用配置
app:
  name: "GoPress"
  env: "development"           # development | staging | production
  port: "8080"
  base_url: "http://localhost:8080"
  base_path: "."               # 项目根目录
  debug: true
  timezone: "Asia/Shanghai"
  secret_key: "change-me-to-a-random-secret-key"

# 数据库配置
database:
  driver: "mysql"              # mysql | postgres | sqlite
  host: "127.0.0.1"
  port: 3306
  name: "gopress"
  user: "root"
  password: "password"
  charset: "utf8mb4"
  max_open_conns: 50
  max_idle_conns: 10
  conn_max_lifetime: 3600     # 秒
  log_level: "warn"           # debug | info | warn | error

# Redis 配置
redis:
  addr: "127.0.0.1:6379"
  password: ""
  db: 0
  pool_size: 20
  prefix: "gopress:"

# JWT 配置
jwt:
  access_secret: "change-access-secret-key"
  refresh_secret: "change-refresh-secret-key"
  access_expire: 3600         # Access Token 有效期（秒）
  refresh_expire: 604800       # Refresh Token 有效期（秒）

# 媒体存储配置
media:
  storage: local               # local | minio
  max_file_size: 52428800     # 50MB
  allowed_types:
    - "image/jpeg"
    - "image/png"
    - "image/gif"
    - "image/webp"
    - "application/pdf"
    - "application/msword"
    - "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

  # 本地存储配置
  local:
    base_path: "./uploads"
    base_url: "/uploads"

  # MinIO 存储配置
  minio:
    endpoint: "localhost:9000"
    access_key: "minioadmin"
    secret_key: "minioadmin"
    bucket: "gopress"
    use_ssl: false
    region: "us-east-1"
    base_url: "http://localhost:9000/gopress"

# 缓存配置
cache:
  driver: "redis"
  default_ttl: 300             # 默认缓存时间（秒）

# 日志配置
log:
  level: "info"               # debug | info | warn | error
  format: "json"              # json | text
  output: "stdout"

# CORS 配置
cors:
  allowed_origins:
    - "http://localhost:3000"
    - "http://localhost:8080"
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "PATCH"
    - "DELETE"
    - "OPTIONS"
  allowed_headers:
    - "Authorization"
    - "Content-Type"
    - "X-Request-ID"
  max_age: 86400              # 秒

# 限流配置
rate_limit:
  enabled: true
  requests: 100               # 窗口内最大请求数
  window: 60                  # 窗口大小（秒）
```

### 环境变量覆盖

配置支持通过环境变量覆盖，使用 `GOPRESS_` 前缀，层级之间用下划线连接：

```bash
# 数据库配置
export GOPRESS_DATABASE_HOST=production-db.example.com
export GOPRESS_DATABASE_PORT=3306
export GOPRESS_DATABASE_NAME=gopress_prod
export GOPRESS_DATABASE_USER=prod_user
export GOPRESS_DATABASE_PASSWORD=secure_password

# Redis 配置
export GOPRESS_REDIS_ADDR=redis.example.com:6379
export GOPRESS_REDIS_PASSWORD=redis_password

# JWT 配置
export GOPRESS_JWT_ACCESS_SECRET=production_access_secret
export GOPRESS_JWT_REFRESH_SECRET=production_refresh_secret

# 应用配置
export GOPRESS_APP_ENV=production
export GOPRESS_APP_DEBUG=false
```

---

## API 文档

### REST API 端点列表

#### 认证相关

| 方法 | 端点 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 | 否 |
| POST | `/api/v1/auth/login` | 用户登录 | 否 |
| POST | `/api/v1/auth/refresh` | 刷新 Token | 否 |

#### 用户相关

| 方法 | 端点 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/users/me` | 获取当前用户信息 | JWT |
| PUT | `/api/v1/users/me` | 更新用户资料 | JWT |
| POST | `/api/v1/users/me/password` | 修改密码 | JWT |
| GET | `/api/v1/admin/users` | 获取用户列表 | Admin |

#### 内容相关

| 方法 | 端点 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/posts` | 获取文章列表（公开） | 否 |
| GET | `/api/v1/posts/:id` | 获取文章详情（公开） | 否 |
| GET | `/api/v1/pages` | 获取页面列表（公开） | 否 |
| GET | `/api/v1/pages/:id` | 获取页面详情（公开） | 否 |
| POST | `/api/v1/posts` | 创建文章 | JWT |
| PUT | `/api/v1/posts/:id` | 更新文章 | JWT |
| DELETE | `/api/v1/posts/:id` | 删除文章 | JWT |
| GET | `/api/v1/posts/:id/revisions` | 获取文章版本历史 | JWT |

#### 分类法相关

| 方法 | 端点 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/v1/terms` | 获取分类/标签列表 | 否 |
| GET | `/api/v1/terms/:id` | 获取分类/标签详情 | 否 |
| POST | `/api/v1/admin/terms` | 创建分类/标签 | Admin |
| PUT | `/api/v1/admin/terms/:id` | 更新分类/标签 | Admin |
| DELETE | `/api/v1/admin/terms/:id` | 删除分类/标签 | Admin |

#### 媒体相关

| 方法 | 端点 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/media/upload` | 上传文件 | JWT |
| GET | `/api/v1/media` | 获取媒体列表 | JWT |
| GET | `/api/v1/media/:uuid` | 获取媒体详情 | JWT |
| PUT | `/api/v1/media/:uuid` | 更新媒体元信息 | JWT |
| DELETE | `/api/v1/media/:uuid` | 删除媒体 | JWT |
| POST | `/api/v1/media/:uuid/thumbnails` | 生成缩略图 | JWT |

#### 主题相关（公开）

| 方法 | 端点 | 说明 | 认证 |
|------|------|------|------|
| GET | `/` | 首页 | 否 |
| GET | `/post/:slug` | 文章详情页 | 否 |
| GET | `/page/:slug` | 页面详情页 | 否 |
| GET | `/theme/assets/*filepath` | 主题静态资源 | 否 |

#### 健康检查

| 方法 | 端点 | 说明 | 认证 |
|------|------|------|------|
| GET | `/health` | 健康检查 | 否 |

### API 请求示例

#### 用户注册

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "email": "newuser@example.com",
    "password": "SecurePassword123"
  }'
```

#### 用户登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "identity": "newuser",
    "password": "SecurePassword123"
  }'
```

#### 创建文章

```bash
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "title": "我的第一篇文章",
    "slug": "my-first-post",
    "content": "这是文章内容...",
    "excerpt": "文章摘要",
    "status": "published",
    "type": "post",
    "term_ids": [1, 2]
  }'
```

### Swagger 访问说明

API 文档可通过 Swagger UI 访问（如果已配置 Swagger 注解）：

```
http://localhost:8080/swagger/index.html
```

---

## 如何开发主题

### 主题目录结构

```
themes/
└── your-theme/               # 主题目录（以 slug 命名）
    ├── theme.yaml            # 主题配置文件（必需）
    ├── screenshot.png        # 主题截图（可选）
    ├── templates/            # 模板文件目录
    │   ├── index.html        # 首页模板
    │   ├── single.html       # 文章详情模板
    │   ├── page.html         # 页面模板
    │   └── layout/           # 布局模板目录
    │       └── base.html     # 基础布局
    │   └── partials/         # 部分模板目录
    │       ├── header.html   # 页头部分
    │       └── footer.html   # 页脚部分
    └── assets/               # 静态资源目录
        ├── css/
        │   └── style.css    # 样式文件
        ├── js/
        │   └── main.js      # 脚本文件
        └── images/          # 图片资源
```

### 模板文件说明

#### theme.yaml 主题配置

```yaml
name: 主题名称
slug: theme-slug          # 主题标识（与目录名一致）
version: 1.0.0
author: 作者名称
description: 主题描述
screenshot: screenshot.png
```

#### 模板文件结构

**index.html（首页模板）**：

```html
{{ define "index" }}
{{ template "layout/base" . }}
{{ end }}

{{ define "content" }}
<div class="posts-list">
    {{ range .Posts }}
    <article class="post-card">
        <h2><a href="/post/{{ .Slug }}">{{ .Title }}</a></h2>
        <div class="meta">
            <span>{{ .AuthorName }}</span>
            <span>{{ dateformat .PublishedAt "2006-01-02" }}</span>
            <span>{{ .ViewCount }} 阅读</span>
        </div>
        <div class="excerpt">{{ .Excerpt }}</div>
        <a href="/post/{{ .Slug }}">阅读更多</a>
    </article>
    {{ end }}

    {{ if .Pagination.HasPrev }}
    <a href="/?page={{ sub .Pagination.Page 1 }}">上一页</a>
    {{ end }}
    {{ if .Pagination.HasNext }}
    <a href="/?page={{ add .Pagination.Page 1 }}">下一页</a>
    {{ end }}
</div>
{{ end }}
```

**single.html（文章详情模板）**：

```html
{{ define "single" }}
{{ template "layout/base" . }}
{{ end }}

{{ define "content" }}
<article class="post">
    <h1>{{ .Post.Title }}</h1>
    <div class="meta">
        <span>作者：{{ .Post.AuthorName }}</span>
        <span>发布于：{{ dateformat .Post.PublishedAt "2006-01-02" }}</span>
        <span>{{ .Post.ViewCount }} 阅读</span>
    </div>

    <div class="categories">
        {{ range .Categories }}
        <a href="/category/{{ .Slug }}">{{ .Name }}</a>
        {{ end }}
    </div>

    <div class="tags">
        {{ range .Tags }}
        <a href="/tag/{{ .Slug }}">{{ .Name }}</a>
        {{ end }}
    </div>

    <div class="content">
        {{ .Post.Content }}
    </div>
</article>
{{ end }}
```

**page.html（页面模板）**：

```html
{{ define "page" }}
{{ template "layout/base" . }}
{{ end }}

{{ define "content" }}
<article class="page">
    <h1>{{ .Page.Title }}</h1>
    <div class="content">
        {{ .Page.Content }}
    </div>
</article>
{{ end }}
```

**layout/base.html（基础布局）**：

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ template "title" . }}</title>
    <link rel="stylesheet" href="/theme/assets/css/style.css">
</head>
<body>
    {{ template "header" . }}
    <main>
        {{ template "content" . }}
    </main>
    {{ template "footer" . }}
    <script src="/theme/assets/js/main.js"></script>
</body>
</html>
```

### 模板变量

#### 首页上下文 (HomeContext)

| 变量 | 类型 | 说明 |
|------|------|------|
| `.Posts` | `[]PostSummary` | 文章列表 |
| `.Pagination.Page` | int | 当前页码 |
| `.Pagination.TotalPages` | int | 总页数 |
| `.Pagination.HasPrev` | bool | 是否有上一页 |
| `.Pagination.HasNext` | bool | 是否有下一页 |
| `.SiteName` | string | 网站名称 |
| `.SiteURL` | string | 网站 URL |
| `.ThemePath` | string | 主题路径 |
| `.AssetsPath` | string | 资源路径 |

#### 文章详情上下文 (PostContext)

| 变量 | 类型 | 说明 |
|------|------|------|
| `.Post` | `PostDetail` | 文章详情 |
| `.Categories` | `[]Taxonomy` | 分类列表 |
| `.Tags` | `[]Taxonomy` | 标签列表 |

#### PostDetail 结构

| 字段 | 类型 | 说明 |
|------|------|------|
| `.ID` | uint | 文章 ID |
| `.Title` | string | 标题 |
| `.Slug` | string | 友好链接 |
| `.Content` | string | 内容（已解析短代码） |
| `.RawContent` | string | 原始内容 |
| `.Excerpt` | string | 摘要 |
| `.AuthorID` | uint | 作者 ID |
| `.AuthorName` | string | 作者名称 |
| `.ViewCount` | int | 阅读数 |
| `.PublishedAt` | time | 发布时间 |
| `.CreatedAt` | time | 创建时间 |

### 内置模板函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `dateformat` | 日期格式化 | `{{ dateformat .Post.PublishedAt "2006-01-02" }}` |
| `truncate` | 字符串截断 | `{{ truncate .Post.Excerpt 100 }}` |
| `safeHTML` | 标记安全 HTML | `{{ safeHTML .Post.Content }}` |
| `add` | 加法 | `{{ add .Page 1 }}` |
| `sub` | 减法 | `{{ sub .Page 1 }}` |
| `len` | 获取长度 | `{{ len .Posts }}` |

### 短代码使用

在文章或页面内容中使用短代码：

#### 图库短代码

```html
[gallery id=1]
```

#### 视频短代码

```html
[video url="https://example.com/video.mp4"]
```

#### 按钮短代码

```html
[button text="点击这里" url="/contact" style="primary"]
```

#### 提示框短代码

```html
[alert type="success" text="操作成功！"]
[alert type="danger" text="发生错误"]
```

### 主题配置

在 `theme.yaml` 中配置主题元信息：

```yaml
name: 我的主题
slug: my-theme
version: 1.0.0
author: 开发者名称
description: 这是一个自定义主题
screenshot: screenshot.png
```

---

## 如何开发插件/扩展

### Hook 系统说明

GoPress 采用基于中间件的扩展模式，可通过以下方式扩展功能：

#### 1. 自定义中间件

创建新的中间件文件 `internal/middleware/custom.go`：

```go
package middleware

import (
    "github.com/gin-gonic/gin"
)

func CustomMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 前置处理
        c.Set("custom_key", "custom_value")

        // 继续处理请求
        c.Next()

        // 后置处理
        // ...
    }
}
```

注册到路由 `internal/bootstrap/router.go`：

```go
r.Use(middleware.CustomMiddleware())
```

#### 2. 自定义短代码

创建新的短代码处理器 `internal/theme/shortcodes/custom.go`：

```go
package theme

import "fmt"

type CustomShortcode struct{}

func (c *CustomShortcode) Name() string {
    return "custom"
}

func (c *CustomShortcode) Render(params map[string]string, content string) string {
    title := params["title"]
    if title == "" {
        title = "Default Title"
    }
    return fmt.Sprintf(`<div class="custom-component" data-title="%s"></div>`, title)
}
```

注册短代码 `internal/bootstrap/app.go`：

```go
shortcodeRegistry.Register(&CustomShortcode{})
```

### 扩展点列表

| 扩展点 | 说明 | 方式 |
|--------|------|------|
| 中间件 | 请求/响应拦截处理 | 新增 middleware 文件 |
| 短代码 | 内容中嵌入组件 | 实现 ShortcodeHandler |
| 存储驱动 | 文件存储方式 | 实现 Storage 接口 |
| 缓存驱动 | 数据缓存方式 | 实现 Cache 接口 |
| 模板函数 | 模板中可调用的函数 | 注册到 FuncMap |
| API Handler | REST API 端点 | 新增 handler 文件 |

---

## 开发指南

### 本地开发

#### 1. 安装开发工具

```bash
# 安装 Air（热重载）
go install github.com/air-verse/air@latest

# 安装代码格式化工具
go install mvdan.cc/gofumpt@latest
```

#### 2. 使用热重载运行

```bash
air
```

#### 3. 配置 IDE

推荐使用 VS Code 或 GoLand，配置 `go.mod` 路径和格式化工具。

### 测试

#### 运行单元测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### 运行集成测试

```bash
# 需要启动 MySQL 和 Redis
go test -tags=integration ./...
```

#### 测试示例

```go
// internal/user/service_test.go
package user

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestUserService_Register(t *testing.T) {
    // 测试代码
    assert.NotNil(t, nil)
}
```

### 代码规范

#### 格式化

```bash
# 格式化代码
gofumpt -w .

# 导入排序
goimports -w .
```

#### 静态分析

```bash
# 运行静态分析
go vet ./...

# 运行所有检查
golangci-lint run
```

#### 提交规范

使用语义化提交信息：

```
feat: 添加新功能
fix: 修复 bug
docs: 文档更新
style: 代码格式（不影响功能）
refactor: 重构
test: 测试相关
chore: 构建/工具相关
```

---

## 部署指南

### 生产环境

#### 1. 服务器要求

- CPU: 2 核+
- 内存: 4GB+
- 磁盘: 50GB+
- 操作系统: Linux (Ubuntu 20.04+)

#### 2. 安装依赖

```bash
# 安装 MySQL
sudo apt update
sudo apt install mysql-server

# 安装 Redis
sudo apt install redis-server

# 安装 Nginx（反向代理）
sudo apt install nginx
```

#### 3. 配置 MySQL

```sql
CREATE DATABASE gopress CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'gopress'@'localhost' IDENTIFIED BY 'strong_password';
GRANT ALL PRIVILEGES ON gopress.* TO 'gopress'@'localhost';
FLUSH PRIVILEGES;
```

#### 4. 构建二进制

```bash
# 在项目目录执行
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gopress ./cmd/server
```

#### 5. 配置 Systemd 服务

创建服务文件 `/etc/systemd/system/gopress.service`：

```ini
[Unit]
Description=GoPress CMS
After=network.target mysql.service redis.service

[Service]
Type=simple
User=gopress
Group=gopress
WorkingDirectory=/opt/gopress
ExecStart=/opt/gopress/gopress
Restart=always
RestartSec=5
Environment="GOPRESS_APP_ENV=production"

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable gopress
sudo systemctl start gopress
```

#### 6. 配置 Nginx 反向代理

创建配置文件 `/etc/nginx/sites-available/gopress`：

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /uploads {
        alias /opt/gopress/storage/uploads;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }
}
```

启用站点：

```bash
sudo ln -s /etc/nginx/sites-available/gopress /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### Docker 部署

#### 使用 Docker Compose（生产环境）

```yaml
version: '3.8'

services:
  app:
    image: gopress:latest
    restart: always
    ports:
      - "127.0.0.1:8080:8080"
    volumes:
      - ./config/production.yaml:/app/config.yaml:ro
      - ./storage:/app/storage
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_started
    environment:
      - GOPRESS_APP_ENV=production
      - GOPRESS_DATABASE_HOST=mysql
      - GOPRESS_DATABASE_PASSWORD=${DB_PASSWORD}
      - GOPRESS_REDIS_ADDR=redis:6379
    networks:
      - gopress-network

  mysql:
    image: mysql:8.0
    restart: always
    environment:
      MYSQL_ROOT_PASSWORD: ${ROOT_PASSWORD}
      MYSQL_DATABASE: gopress
      MYSQL_USER: gopress
      MYSQL_PASSWORD: ${DB_PASSWORD}
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - gopress-network

  redis:
    image: redis:7-alpine
    restart: always
    volumes:
      - redis_data:/data
    networks:
      - gopress-network

volumes:
  mysql_data:
  redis_data:

networks:
  gopress-network:
    driver: bridge
```

### 环境变量

| 变量 | 说明 | 示例 |
|------|------|------|
| `GOPRESS_APP_ENV` | 运行环境 | `production` |
| `GOPRESS_DATABASE_HOST` | 数据库主机 | `mysql` |
| `GOPRESS_DATABASE_PORT` | 数据库端口 | `3306` |
| `GOPRESS_DATABASE_NAME` | 数据库名称 | `gopress` |
| `GOPRESS_DATABASE_USER` | 数据库用户 | `gopress` |
| `GOPRESS_DATABASE_PASSWORD` | 数据库密码 | `password` |
| `GOPRESS_REDIS_ADDR` | Redis 地址 | `redis:6379` |
| `GOPRESS_REDIS_PASSWORD` | Redis 密码 | `password` |
| `GOPRESS_JWT_ACCESS_SECRET` | JWT Access Secret | `secret` |
| `GOPRESS_JWT_REFRESH_SECRET` | JWT Refresh Secret | `secret` |

---

## 贡献指南

### 开发流程

1. **Fork 项目**：点击 GitHub 页面右上角的 Fork 按钮

2. **克隆仓库**：
   ```bash
   git clone https://github.com/YOUR_USERNAME/gopress.git
   cd gopress
   ```

3. **创建分支**：
   ```bash
   git checkout -b feature/your-feature-name
   ```

4. **开发**：
   - 编写代码
   - 添加测试
   - 确保通过所有测试

5. **提交**：
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

6. **推送**：
   ```bash
   git push origin feature/your-feature-name
   ```

7. **创建 Pull Request**：在 GitHub 上创建 PR，等待代码审查

### 代码审查标准

- [ ] 代码符合 Go 编码规范
- [ ] 添加了必要的单元测试
- [ ] 文档已更新（如需要）
- [ ] 没有引入新的编译警告
- [ ] 测试全部通过

### 报告问题

通过 GitHub Issues 报告问题，请包含：

- 问题描述
- 复现步骤
- 预期行为
- 实际行为
- 环境信息（Go 版本、操作系统等）

---

## 许可证

本项目采用 MIT 许可证，详见 [LICENSE](LICENSE) 文件。

## 联系方式

- 项目主页：https://github.com/yourorg/gopress
- 问题反馈：https://github.com/yourorg/gopress/issues
- 讨论交流：https://github.com/yourorg/gopress/discussions
