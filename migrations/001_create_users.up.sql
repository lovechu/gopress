-- GoPress: 创建用户表
-- Version: 001
-- Up Migration

CREATE TABLE IF NOT EXISTS `users` (
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `username`      VARCHAR(64)     NOT NULL COMMENT '用户名（唯一）',
    `email`         VARCHAR(128)    NOT NULL COMMENT '邮箱（唯一）',
    `password_hash` VARCHAR(256)    NOT NULL COMMENT 'bcrypt 哈希',
    `display_name`  VARCHAR(128)    NOT NULL DEFAULT '' COMMENT '昵称',
    `avatar`        VARCHAR(512)    NOT NULL DEFAULT '' COMMENT '头像 URL',
    `role`          VARCHAR(32)     NOT NULL DEFAULT 'subscriber'
                    COMMENT '角色: admin|editor|author|subscriber',
    `bio`           TEXT                     COMMENT '个人简介',
    `is_active`     TINYINT(1)      NOT NULL DEFAULT 1 COMMENT '是否启用',
    `last_login_at` DATETIME                 COMMENT '最后登录时间',
    `created_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP
                    ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`    DATETIME                 COMMENT '软删除时间',

    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_users_username` (`username`),
    UNIQUE KEY `uq_users_email`    (`email`),
    KEY `idx_users_role`           (`role`),
    KEY `idx_users_is_active`     (`is_active`),
    KEY `idx_users_deleted_at`    (`deleted_at`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='用户表';

-- 插入默认 admin（密码: Admin@GoPress2026，bcrypt cost=12）
-- 生产环境部署后请立即修改密码
INSERT INTO `users` (
    `username`, `email`, `password_hash`, `display_name`, `role`, `is_active`
) VALUES (
    'admin',
    'admin@gopress.local',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj0rTlExiDvy',
    'Administrator',
    'admin',
    1
);
