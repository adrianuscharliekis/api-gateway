package model

import "time"

type Tracelog struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	IP          string    `gorm:"type:varchar(45)"`       // For IPv6 compatibility
	Proses      string    `gorm:"type:varchar(50);index"` // Indexed for faster queries
	CaCode      string    `gorm:"type:varchar(20);index"` // Unique client codes
	ProductType string    `gorm:"type:varchar(30);index"` // E.g. "sindoferry"
	Log         string    `gorm:"type:text"`              // Long message, don't index
	Tracetime   time.Time `gorm:"type:datetime;index"`    // Useful for sorting/filtering
}
