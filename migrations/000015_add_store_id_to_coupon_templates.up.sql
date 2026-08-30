-- Add store_id to coupon_templates: the store that owns/created the template,
-- used to resolve source_store_name / qr_code_url / mp redirect for its coupons.
ALTER TABLE coupon_templates
    ADD COLUMN store_id BIGINT UNSIGNED NULL COMMENT '创建/所属门店（发券方归属）' AFTER user_id,
    ADD INDEX idx_template_store (store_id),
    ADD CONSTRAINT fk_template_store FOREIGN KEY (store_id) REFERENCES stores(id);

-- Backfill: for existing specific-scope templates, take their first applicable store.
UPDATE coupon_templates t
    JOIN (SELECT template_id, MIN(store_id) AS sid FROM coupon_template_stores GROUP BY template_id) c
      ON c.template_id = t.id
SET t.store_id = c.sid
WHERE t.applicable_scope = 'specific' AND t.store_id IS NULL;