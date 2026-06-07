CREATE TABLE coupon_usage_records (
    id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    coupon_id      BIGINT UNSIGNED NOT NULL,
    user_phone     VARCHAR(20)     NOT NULL,
    store_id       BIGINT UNSIGNED NOT NULL COMMENT 'Store that performed the action',
    action         ENUM('consume','refund','expire','freeze','unfreeze')
                                    NOT NULL,
    order_info     JSON            DEFAULT NULL COMMENT 'Related order data',
    operator       VARCHAR(64)     DEFAULT NULL,
    ip_address     VARCHAR(45)     DEFAULT NULL,
    created_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY idx_coupon_id (coupon_id),
    KEY idx_user_phone (user_phone),
    KEY idx_store_id (store_id),
    KEY idx_created_at (created_at),
    CONSTRAINT fk_cur_coupon FOREIGN KEY (coupon_id) REFERENCES coupon_instances(id),
    CONSTRAINT fk_cur_store FOREIGN KEY (store_id) REFERENCES stores(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Coupon usage audit records';
