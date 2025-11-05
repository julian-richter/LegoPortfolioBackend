package models

type User struct {
	BaseModel
	Username     string `json:"username" db:"username"`
	PasswordHash string `json:",omitempty" db:"password_hash"`
	FirstName    string `json:"first_name" db:"first_name"`
	LastName     string `json:"last_name" db:"last_name"`
}

// TableName returns the database table name
func (User) TableName() string {
	return "users"
}

// FullName returns the user's full name
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}
