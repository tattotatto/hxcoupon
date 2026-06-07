CREATE TABLE admin_users (
    id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username       VARCHAR(64)     NOT NULL,
    password_hash  VARCHAR(256)    NOT NULL,
    role           ENUM('super_admin','admin','viewer') NOT NULL DEFAULT 'admin',
    status         TINYINT(1)      NOT NULL DEFAULT 1,
    last_login_at  DATETIME        DEFAULT NULL,
    created_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Admin dashboard users';
