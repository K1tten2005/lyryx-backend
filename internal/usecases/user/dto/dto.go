package dto

import "mime/multipart"

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

type UploadAvatarOpts struct {
	UserID     int
	AvatarFile *multipart.FileHeader
}

type PatchUpdateAvatarOpts struct {
	UserID    int
	AvatarURL string
}
