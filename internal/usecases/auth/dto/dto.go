package dto

type SignUpOpts struct {
	Username string
	Email    string
	Password string
}

type UserInfo struct {
	UserID          int
	Email           string
	Username        string
	ReputationScore int
	Role            string
}

type SignInOpts struct {
	Email    string
	Password string
}

type SetNewRefreshTokenOpts struct {
	Email        string
	RefreshToken string
}

type SignOutOpts struct {
	Email string
}
