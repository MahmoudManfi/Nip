-- MySQL schema for users and follows

CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_name VARCHAR(50) NOT NULL,
  email_or_phone VARCHAR(100) NOT NULL,
  is_email BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_users_user_name (user_name),
  UNIQUE KEY uq_users_email_or_phone (email_or_phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS follows (
  follower_id BIGINT UNSIGNED NOT NULL,
  followed_id BIGINT UNSIGNED NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (follower_id, followed_id),
  KEY idx_follows_follower (follower_id),
  KEY idx_follows_followed (followed_id),
  CONSTRAINT fk_follows_follower FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_follows_followed FOREIGN KEY (followed_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT chk_not_self_follow CHECK (follower_id <> followed_id),
  INDEX idx_followed (followed_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

