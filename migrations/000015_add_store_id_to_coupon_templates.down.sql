ALTER TABLE coupon_templates
    DROP FOREIGN KEY fk_template_store,
    DROP INDEX idx_template_store,
    DROP COLUMN store_id;