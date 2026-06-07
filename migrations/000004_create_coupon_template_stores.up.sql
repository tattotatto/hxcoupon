CREATE TABLE coupon_template_stores (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    template_id BIGINT UNSIGNED NOT NULL,
    store_id    BIGINT UNSIGNED NOT NULL,
    created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_template_store (template_id, store_id),
    KEY idx_store_id (store_id),
    CONSTRAINT fk_cts_template FOREIGN KEY (template_id) REFERENCES coupon_templates(id),
    CONSTRAINT fk_cts_store FOREIGN KEY (store_id) REFERENCES stores(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Template-to-store mapping for specific-scope coupons';
