package storage

import "mime/multipart"

type Artist struct {
	ArtistID  int
	Name      string
	Bio       string
	AvatarURL string
}

type CreateArtistFilter struct {
	Name string
	Bio  string
}

type UpdateArtistInfoFilter struct {
	ArtistID int
	Name     *string
	Bio      *string
}

type UploadAvatarFilter struct {
	ArtistID     int
	AvatarFile *multipart.FileHeader
}

type UpdateArtistAvatarFilter struct {
	ArtistID   int
	AvatarURL  string
}
