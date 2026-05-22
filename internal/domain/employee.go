package domain

import "time"

type Employee struct {
	ID           uint       `gorm:"primaryKey"`
	DepartmentID uint       `gorm:"not null;index"`
	Department   Department `gorm:"foreignKey:DepartmentID;constraint:OnDelete:CASCADE"`
	FullName     string     `gorm:"type:varchar(200);not null"`
	Position     string     `gorm:"type:varchar(200);not null"`
	HiredAt      *time.Time
	CreatedAt    time.Time
}
