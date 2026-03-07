package user

type GetUserMeOut struct {
	UserID          int    `json:"user_id"`
	Email           string `json:"email"`
	Username        string `json:"username"`
	ReputationScore int    `json:"reputation_score"`
	Role            string `json:"role"`
}

type GetUserByIDIn struct {
	UserID int `param:"id" validate:"required"`
}

type GetUserByIDOut struct {
	UserID          int    `json:"user_id"`
	Email           string `json:"email"`
	Username        string `json:"username"`
	ReputationScore int    `json:"reputation_score"`
	Role            string `json:"role"`
}