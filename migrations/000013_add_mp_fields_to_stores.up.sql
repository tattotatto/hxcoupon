ALTER TABLE stores
    ADD COLUMN mp_appid     VARCHAR(64)  DEFAULT NULL COMMENT 'Mini-program AppID for use-coupon redirect',
    ADD COLUMN mp_page_path VARCHAR(256) DEFAULT NULL COMMENT 'Mini-program page path for use-coupon redirect';
