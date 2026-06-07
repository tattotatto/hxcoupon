CREATE TABLE coupon_templates (
    id                 BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name               VARCHAR(128)    NOT NULL COMMENT 'Coupon template name',
    type               ENUM('full_reduction','discount','fixed_amount')
                                       NOT NULL COMMENT 'Coupon type',
    discount_value     DECIMAL(10,2)   NOT NULL COMMENT 'Discount amount (yuan or percentage)',
    threshold_amount   DECIMAL(10,2)   NOT NULL DEFAULT 0.00 COMMENT 'Min order amount (0=no threshold)',
    applicable_scope   ENUM('all','specific') NOT NULL DEFAULT 'all' COMMENT 'Which stores can use',
    stackable          TINYINT(1)      NOT NULL DEFAULT 0 COMMENT 'Can stack with other coupons',
    max_stack_count    TINYINT UNSIGNED DEFAULT 1 COMMENT 'Max coupons when stacking',
    validity_type      ENUM('fixed_date','days_after_receive')
                                       NOT NULL COMMENT 'How validity is calculated',
    validity_days      INT UNSIGNED    DEFAULT NULL COMMENT 'Valid days after receipt',
    valid_start        DATETIME        DEFAULT NULL COMMENT 'Fixed start date',
    valid_end          DATETIME        DEFAULT NULL COMMENT 'Fixed end date',
    total_quantity     INT UNSIGNED    DEFAULT 0 COMMENT 'Total issued limit (0=unlimited)',
    issued_count       INT UNSIGNED    NOT NULL DEFAULT 0 COMMENT 'How many have been issued',
    per_user_limit     INT UNSIGNED    DEFAULT 1 COMMENT 'Max per user for this template',
    product_restriction JSON           DEFAULT NULL COMMENT 'Product/category restrictions',
    status             TINYINT(1)      NOT NULL DEFAULT 0 COMMENT '0=draft 1=enabled 2=disabled',
    created_by         VARCHAR(64)     DEFAULT NULL,
    created_at         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_status (status),
    KEY idx_type (type),
    KEY idx_validity (validity_type, valid_start, valid_end)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Coupon templates';
