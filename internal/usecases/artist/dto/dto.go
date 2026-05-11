package dto

import "mime/multipart"

type Artist struct {
	ArtistID  int
	Name      string
	Bio       string
	AvatarURL string
}

type GetArtistByIDOpts struct {
	ArtistID int
	Limit    int
	Offset   int
}

type GetArtistByIDResp struct {
	ArtistID  int
	Name      string
	Bio       string
	AvatarURL string
	Songs     []Song
}

type Song struct {
	ID          int
	Title       string
	CoverURL    string
	Views       int
	ReleaseDate string
}

type PostArtistOpts struct {
	Name string
	Bio  string
}

type PatchUpdateArtistOpts struct {
	ArtistID int
	Name     *string
	Bio      *string
}

type UploadAvatarOpts struct {
	ArtistID   int
	AvatarFile *multipart.FileHeader
}

type PatchUpdateAvatarOpts struct {
	ArtistID  int
	AvatarURL string
}
