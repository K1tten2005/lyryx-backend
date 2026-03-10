package user

type GetUserMeOut struct {
	UserID          int    `json:"user_id"`
	Email           string `json:"email"`
	Username        string `json:"username"`
	Bio             string `json:"bio"`
	AvatarURL       string `json:"avatar_url"`
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
	Bio             string `json:"bio"`
	AvatarURL       string `json:"avatar_url"`
	ReputationScore int    `json:"reputation_score"`
	Role            string `json:"role"`
}

type PatchUpdateUserIn struct {
	Email    *string `json:"email"`
	Username *string `json:"username"`
	Bio      *string `json:"bio"`
	Password *string `json:"password"`
}

type PatchUpdateUserOut struct {
	UserID          int    `json:"user_id"`
	Email           string `json:"email"`
	Username        string `json:"username"`
	Bio             string `json:"bio"`
	AvatarURL       string `json:"avatar_url"`
	ReputationScore int    `json:"reputation_score"`
	Role            string `json:"role"`
}

type PatchUpdateAvatarOut struct {
	AvatarURL string `json:"avatar_url"`
}
