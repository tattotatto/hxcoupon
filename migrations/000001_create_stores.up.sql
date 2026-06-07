CREATE TABLE stores (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name          VARCHAR(128)    NOT NULL COMMENT 'Store display name',
    code          VARCHAR(5)      NOT NULL COMMENT '5-char code for coupon prefix',
    app_id        VARCHAR(64)     NOT NULL COMMENT 'Unique application ID',
    type          ENUM('miniprogram','h5') NOT NULL COMMENT 'Channel type',
    status        TINYINT(1)      NOT NULL DEFAULT 1 COMMENT '1=active 0=disabled',
    contact_name  VARCHAR(64)     DEFAULT NULL,
    contact_phone VARCHAR(20)     DEFAULT NULL,
    remark        VARCHAR(512)    DEFAULT NULL,
    created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_code (code),
    UNIQUE KEY uk_app_id (app_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Stores / channels (tenants)';
