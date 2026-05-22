package domain

import "time"

type Department struct {
	ID        uint        `gorm:"primaryKey"`
	Name      string      `gorm:"type:varchar(200);not null"`
	ParentID  *uint       `gorm:"index"`
	Parent    *Department `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time
}
