ALTER TABLE admin_users
    DROP COLUMN business_license,
    DROP COLUMN approved_by,
    DROP COLUMN approved_at,
    DROP COLUMN registered_at,
    DROP COLUMN reject_reason,
    DROP COLUMN email,
    DROP COLUMN contact_phone,
    DROP COLUMN contact_name,
    DROP COLUMN company_name,
    DROP COLUMN approval_status,
    DROP COLUMN member_type,
    MODIFY COLUMN role ENUM('super_admin','admin','viewer') NOT NULL DEFAULT 'admin';
