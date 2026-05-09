CREATE TABLE IF NOT EXISTS posts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  slug VARCHAR(255) NOT NULL UNIQUE,
  content LONGTEXT,
  excerpt TEXT,
  status VARCHAR(32) NOT NULL DEFAULT 'draft',
  type VARCHAR(32) NOT NULL DEFAULT 'post',
  author_id BIGINT UNSIGNED NOT NULL,
  comment_allowed TINYINT(1) NOT NULL DEFAULT 1,
  view_count INT NOT NULL DEFAULT 0,
  published_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  INDEX idx_author (author_id),
  INDEX idx_status (status),
  INDEX idx_type (type),
  INDEX idx_published (published_at),
  INDEX idx_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS post_terms (
  post_id BIGINT UNSIGNED NOT NULL,
  term_id BIGINT UNSIGNED NOT NULL,
  sort INT NOT NULL DEFAULT 0,
  PRIMARY KEY (post_id, term_id),
  INDEX idx_post (post_id),
  INDEX idx_term (term_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS post_revisions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  post_id BIGINT UNSIGNED NOT NULL,
  title VARCHAR(255),
  content LONGTEXT,
  excerpt TEXT,
  status VARCHAR(32),
  revision_number INT NOT NULL,
  changed_by BIGINT UNSIGNED NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_post (post_id),
  INDEX idx_revision (post_id, revision_number)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
