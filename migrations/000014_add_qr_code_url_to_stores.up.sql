ALTER TABLE stores
    ADD COLUMN qr_code_url VARCHAR(512) DEFAULT NULL COMMENT 'QR code image URL for store entry';
