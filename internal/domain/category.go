package domain

// Category groups medicines. Managed by Admin; visible to all.
type Category struct {
	Base
	Name        string  `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Description *string `gorm:"type:text" json:"description,omitempty"`
	IsActive    bool    `gorm:"not null;default:true" json:"is_active"`
}

// TableName pins the table name.
func (Category) TableName() string { return "categories" }
