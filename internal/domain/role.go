package domain

type Role string

const (
	RoleUser    Role = "user"
	RoleAdmin   Role = "admin"
	RoleAnalyst Role = "analyst"
)

func (r Role) Valid() bool {
	switch r {
	case RoleUser, RoleAdmin, RoleAnalyst:
		return true
	default:
		return false
	}
}