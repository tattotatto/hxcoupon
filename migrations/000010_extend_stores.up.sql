-- Extend stores with user ownership and expanded platform types
ALTER TABLE stores
    MODIFY COLUMN type ENUM('miniprogram','h5','app','web','api','other') NOT NULL DEFAULT 'api'
        COMMENT '平台类型',
    ADD COLUMN user_id BIGINT UNSIGNED NULL COMMENT '所属用户ID',
    ADD COLUMN description VARCHAR(512) NULL COMMENT '应用描述',
    ADD INDEX idx_store_user (user_id),
    ADD CONSTRAINT fk_store_user FOREIGN KEY (user_id) REFERENCES admin_users(id);
