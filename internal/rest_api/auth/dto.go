package auth

// sign-up
type PostSignUpIn struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type PostSignUpOut struct {
	AccessToken string `json:"access_token"`
}

// sign-in
type PostSignInIn struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type PostSignInOut struct {
	AccessToken string `json:"access_token"`
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