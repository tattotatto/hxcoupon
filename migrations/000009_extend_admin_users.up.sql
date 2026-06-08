-- Extend admin_users for platform member registration and approval system
ALTER TABLE admin_users
    MODIFY COLUMN role ENUM('super_admin','admin','member') NOT NULL DEFAULT 'member'
        COMMENT '角色: super_admin=超管 admin=运营 member=注册用户',
    ADD COLUMN member_type ENUM('issuer','consumer','both') NULL
        COMMENT '用户类型: issuer=发券方 consumer=用券方 both=两者皆是',
    ADD COLUMN approval_status TINYINT NOT NULL DEFAULT 0
        COMMENT '审批状态: 0=待审批 1=已通过 2=已驳回 3=已冻结',
    ADD COLUMN company_name VARCHAR(128) NULL COMMENT '企业名称',
    ADD COLUMN contact_name VARCHAR(64) NULL COMMENT '联系人',
    ADD COLUMN contact_phone VARCHAR(20) NULL COMMENT '联系电话',
    ADD COLUMN email VARCHAR(128) NULL COMMENT '邮箱',
    ADD COLUMN reject_reason VARCHAR(512) NULL COMMENT '驳回原因',
    ADD COLUMN registered_at DATETIME NULL COMMENT '注册时间',
    ADD COLUMN approved_at DATETIME NULL COMMENT '审批通过时间',
    ADD COLUMN approved_by BIGINT UNSIGNED NULL COMMENT '审批人ID',
    ADD COLUMN business_license VARCHAR(256) NULL COMMENT '营业执照URL';

-- Update existing super_admin/admin accounts to approved status
UPDATE admin_users SET approval_status = 1 WHERE role IN ('super_admin', 'admin');
