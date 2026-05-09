-- GoPress Media Module Migration
-- Migration: 004_create_media
-- Description: Create media table for file storage

-- Create media table
CREATE TABLE IF NOT EXISTS `media` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `uuid` VARCHAR(36) NOT NULL,
    `file_name` VARCHAR(255) NOT NULL,
    `original_name` VARCHAR(500) NOT NULL,
    `file_size` BIGINT NOT NULL DEFAULT 0,
    `mime_type` VARCHAR(100) NOT NULL,
    `media_type` VARCHAR(20) NOT NULL DEFAULT 'other',
    `width` INT NOT NULL DEFAULT 0,
    `height` INT NOT NULL DEFAULT 0,
    `alt` VARCHAR(255) DEFAULT '',
    `caption` TEXT,
    `storage_key` VARCHAR(500) NOT NULL,
    `thumbnail_keys` VARCHAR(1000) DEFAULT '',
    `user_id` INT UNSIGNED NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE INDEX `idx_uuid` (`uuid`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_media_type` (`media_type`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add foreign key constraint (optional, uncomment if you want enforced FK)
-- ALTER TABLE `media` ADD CONSTRAINT `fk_media_user_id` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE;
