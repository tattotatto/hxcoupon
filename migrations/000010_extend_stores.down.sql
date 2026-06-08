ALTER TABLE stores
    DROP FOREIGN KEY fk_store_user,
    DROP INDEX idx_store_user,
    DROP COLUMN description,
    DROP COLUMN user_id,
    MODIFY COLUMN type ENUM('miniprogram','h5') NOT NULL;
