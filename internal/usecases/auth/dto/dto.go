package dto

type SignUpOpts struct {
	Username string
	Email    string
	Password string
}

type UserInfo struct {
	UserID     int
	Email      string
	Username   string
	Role       string
}