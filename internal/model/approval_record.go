package model

import "time"

type ApprovalRecord struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64    `gorm:"not null;index" json:"user_id"`
	Action     string    `gorm:"type:enum('approve','reject','suspend','unsuspend');not null" json:"action"`
	Reason     *string   `gorm:"type:varchar(512)" json:"reason"`
	OperatedBy uint64    `gorm:"not null" json:"operated_by"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ApprovalRecord) TableName() string { return "approval_records" }
