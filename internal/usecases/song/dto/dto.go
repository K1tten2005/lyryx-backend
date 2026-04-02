package dto

import (
	"mime/multipart"
	"time"
)

type SongInfo struct {
	SongID      int
	Title       string
	Lyrics      string
	CoverURL    string
	ReleaseDate time.Time
	Views       int
	Artist      Artist
}

type Artist struct {
	ArtistID  int
	Name      string
	Bio       string
	AvatarURL string
}

type PostSongOpts struct {
	Title       string
	Lyrics      string
	CoverURL    string
	ReleaseDate time.Time
	ArtistID    int
}

type PatchUpdateSongOpts struct {
	SongID      int
	Title       *string
	ArtistID    *int
	Lyrics      *string
	ReleaseDate *time.Time
}

type UploadCoverOpts struct {
	SongID    int
	CoverFile *multipart.FileHeader
}

type PatchUpdateCoverOpts struct {
	SongID   int
	CoverURL string
}
