package dto

type User struct {
	UserID          int
	Email           string
	Username        string
	Bio             string
	AvatarURL       string
	ReputationScore int
	Role            string
}

type PatchUpdateUserOpts struct {
	UserID   int
	Email    *string
	Username *string
	Bio      *string
	Password *string
}

type PatchUpdateAvatarOpts struct {
	UserID    int
	AvatarURL string
}
