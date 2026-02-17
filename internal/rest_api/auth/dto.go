package auth

type PostSignUpIn struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type PostSignUpOut struct {
	AccessToken string `json:"access_token"`
}
