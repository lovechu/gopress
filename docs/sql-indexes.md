# GoPress SQL 索引建议

## 已创建的索引（在 migration 文件中）

### users 表
- `idx_users_username` — `username` 唯一索引（登录查找）
- `idx_users_email` — `email` 唯一索引（登录查找）
- `idx_users_last_login_at` — `last_login_at` 普通索引（活跃用户查询）
- `idx_users_deleted_at` — `deleted_at` 普通索引（软删除查询）

### terms 表
- `idx_terms_taxonomy_slug` — `(taxonomy, slug)` 联合唯一索引（按分类/标签查找）
- `idx_terms_parent_id` — `parent_id` 普通索引（层级查询）
- `idx_terms_deleted_at` — `deleted_at` 普通索引

### posts 表
- `idx_posts_author_id` — `author_id` 普通索引（按作者查询）
- `idx_posts_status` — `status` 普通索引（按状态筛选）
- `idx_posts_type` — `type` 普通索引（按类型筛选）
- `idx_posts_published_at` — `published_at` 普通索引（时间排序）
- `idx_posts_deleted_at` — `deleted_at` 普通索引

### post_terms 表
- `PRIMARY KEY (post_id, term_id)` — 联合主键（避免重复关联）
- `idx_post_terms_post` — `post_id` 普通索引（按文章查术语）
- `idx_post_terms_term` — `term_id` 普通索引（按术语查文章）

### post_revisions 表
- `idx_post_revisions_post` — `post_id` 普通索引（按文章查历史）
- `idx_post_revisions_revision` — `(post_id, revision_number)` 联合索引（版本排序）

---

## 建议追加的索引（根据查询模式）

```sql
-- 文章列表：按状态 + 类型 + 发布时间排序（最常见查询）
CREATE INDEX idx_posts_status_type_published
  ON posts (status, type, published_at DESC);

-- 文章列表：按作者 + 状态查询
CREATE INDEX idx_posts_author_status
  ON posts (author_id, status);

-- post_terms：按 term 查文章时同时按 post 排序
CREATE INDEX idx_post_terms_term_post
  ON post_terms (term_id, post_id);

-- 术语计数更新频繁，确保 count 字段有索引（如果用于排序）
CREATE INDEX idx_terms_count ON terms (count DESC);
```

---

## 索引设计原则

1. **主键**：所有表使用 `BIGINT UNSIGNED AUTO_INCREMENT` 作为主键
2. **外键不建物理外键**：用逻辑外键 + 索引代替，方便分库分表
3. **多对多中间表**：必须建 `(a_id, b_id)` 联合主键，再加反向索引
4. **软删除**：所有 `deleted_at` 字段加索引，避免全表扫描
5. **查询覆盖**：`WHERE` + `ORDER BY` + `LIMIT` 场景需要联合索引，注意字段顺序
6. **避免过多索引**：写多读少的表（如 revisions）索引从简，revisions 只需 `post_id` 索引

---

## GORM AutoMigrate 说明

开发环境使用 `db.AutoMigrate(...)` 自动建表，但 **GORM 不会自动创建所有索引**，特别是：
- 联合索引需手写 SQL（已在 migration 文件中提供）
- 唯一索引需 GORM tag `uniqueIndex` 或手写 SQL
- 索引命名规范：`idx_<table>_<column(s)>`

生产环境建议：
1. 禁用 AutoMigrate
2. 使用 migration 文件（已在 `migrations/` 提供）
3. 用 `migrate` 工具（如 golang-migrate/migrate）管理版本化迁移
