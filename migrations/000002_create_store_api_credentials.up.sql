CREATE TABLE store_api_credentials (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    store_id      BIGINT UNSIGNED NOT NULL,
    app_key       VARCHAR(64)     NOT NULL COMMENT 'API Key / Access Key',
    app_secret    VARCHAR(256)    NOT NULL COMMENT 'Secret key (bcrypt hashed)',
    status        TINYINT(1)      NOT NULL DEFAULT 1 COMMENT '1=active 0=disabled',
    created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_app_key (app_key),
    KEY idx_store_id (store_id),
    CONSTRAINT fk_cred_store FOREIGN KEY (store_id) REFERENCES stores(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Store API credentials for HMAC auth';
