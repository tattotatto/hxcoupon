INSERT INTO admin_users (username, password_hash, role, status, approval_status) VALUES
('admin', '$2b$10$gMuU3pFv0I15som5HdpePuObyQ/HmBh.in8jZyNkr26GBi3wuXXai', 'super_admin', 1, 1)
ON DUPLICATE KEY UPDATE password_hash = VALUES(password_hash), approval_status = 1;
-- Default password: admin123
