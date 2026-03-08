package storage

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

type UpdateUserInfoFilter struct {
	UserID   int
	Email    *string
	Username *string
	Bio      *string
	Password *string
}

type UpdateUserAvatarFilter struct {
	UserID    int
	AvatarURL string
}

type UploadAvatarFilter struct {
	UserID     int
	AvatarFile *multipart.FileHeader
}
