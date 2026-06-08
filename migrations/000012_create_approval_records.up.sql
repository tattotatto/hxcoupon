-- Approval audit log table
CREATE TABLE approval_records (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL COMMENT '被审批的用户ID',
    action ENUM('approve','reject','suspend','unsuspend') NOT NULL COMMENT '审批动作',
    reason VARCHAR(512) NULL COMMENT '审批意见/原因',
    operated_by BIGINT UNSIGNED NOT NULL COMMENT '操作人ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
    INDEX idx_user (user_id),
    INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
