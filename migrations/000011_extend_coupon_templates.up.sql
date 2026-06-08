-- Extend coupon_templates with user ownership
ALTER TABLE coupon_templates
    ADD COLUMN user_id BIGINT UNSIGNED NULL COMMENT '创建者ID（发券方）',
    ADD COLUMN description TEXT NULL COMMENT '模板使用说明',
    ADD INDEX idx_template_user (user_id),
    ADD CONSTRAINT fk_template_user FOREIGN KEY (user_id) REFERENCES admin_users(id);
