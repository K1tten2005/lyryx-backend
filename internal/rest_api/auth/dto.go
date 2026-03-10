package auth

type UserInfo struct {
	UserID          int    `json:"user_id"`
	Email           string `json:"email"`
	Username        string `json:"username"`
	ReputationScore int    `json:"reputation_score"`
	Role            string `json:"role"`
}

// sign-up
type PostSignUpIn struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type PostSignUpOut struct {
	User        UserInfo `json:"user"`
	AccessToken string   `json:"access_token"`
}

// sign-in
type PostSignInIn struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type PostSignInOut struct {
	User        UserInfo `json:"user"`
	AccessToken string   `json:"access_token"`
}

// refresh
type PostRefreshTokenOut struct {
	AccessToken string `json:"access_token"`
}

// update
type PutUpdateUserIn struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required"`
}
