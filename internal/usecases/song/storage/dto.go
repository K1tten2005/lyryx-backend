package storage

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

type CreateSongFilter struct {
	Title       string
	Lyrics      string
	CoverURL    string
	ReleaseDate time.Time
	ArtistID    int
}

type UpdateSongInfoFilter struct {
	SongID      int
	Title       *string
	ArtistID    *int
	Lyrics      *string
	ReleaseDate *time.Time
}

type UploadCoverFilter struct {
	SongID    int
	CoverFile *multipart.FileHeader
}

type UpdateSongCoverFilter struct {
	SongID   int
	CoverURL string
}
