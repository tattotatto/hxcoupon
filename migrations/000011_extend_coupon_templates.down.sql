ALTER TABLE coupon_templates
    DROP FOREIGN KEY fk_template_user,
    DROP INDEX idx_template_user,
    DROP COLUMN description,
    DROP COLUMN user_id;
