package domain

// UserRole enumerates the access levels in the system.
type UserRole string

const (
	RoleAdmin UserRole = "ADMIN"
	RoleStaff UserRole = "STAFF"
)

// Valid reports whether r is a recognised role.
func (r UserRole) Valid() bool {
	return r == RoleAdmin || r == RoleStaff
}

// User is an authenticated account. Passwords are stored hashed only.
type User struct {
	Base
	Username     string   `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	PasswordHash string   `gorm:"type:varchar(255);not null" json:"-"`
	FullName     string   `gorm:"type:varchar(100);not null" json:"full_name"`
	Role         UserRole `gorm:"type:varchar(20);not null;default:STAFF" json:"role"`
	LineUserID   *string  `gorm:"type:varchar(100)" json:"line_user_id,omitempty"`
	IsActive     bool     `gorm:"not null;default:true" json:"is_active"`
}

// TableName pins the table name.
func (User) TableName() string { return "users" }
