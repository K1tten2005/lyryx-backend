package song

import "time"

type GetSongByIDIn struct {
	SongID int `param:"id" validate:"required"`
}

type GetSongByIDOut struct {
	SongID      int       `json:"id"`
	Title       string    `json:"title"`
	Lyrics      string    `json:"lyrics"`
	CoverURL    string    `json:"cover_url"`
	ReleaseDate time.Time `json:"release_date"`
	Views       int       `json:"views"`
	Artist      Artist    `json:"artist"`
}

type Artist struct {
	ArtistID  int    `json:"id"`
	Name      string `json:"name"`
	Bio       string `json:"bio"`
	AvatarURL string `json:"avatar_url"`
}

type PostSongIn struct {
	Title       string `json:"title" validate:"required"`
	Lyrics      string `json:"lyrics" validate:"required"`
	CoverURL    string `json:"cover_url,omitempty"`
	ReleaseDate string `json:"release_date" validate:"required"`
	ArtistID    int    `json:"artist_id" validate:"required"`
}

type PostSongOut struct {
	SongID      int       `json:"id"`
	Title       string    `json:"title"`
	Lyrics      string    `json:"lyrics"`
	CoverURL    string    `json:"cover_url"`
	ReleaseDate time.Time `json:"release_date"`
	Views       int       `json:"views"`
	Artist      Artist    `json:"artist"`
}

type PatchUpdateSongIn struct {
	SongID      int     `param:"id" validate:"required"`
	Title       *string `json:"title"`
	ArtistID    *int    `json:"artist_id"`
	Lyrics      *string `json:"lyrics"`
	ReleaseDate *string `json:"release_date"`
}

type PatchUpdateSongOut struct {
	SongID      int       `json:"id"`
	Title       string    `json:"title"`
	Lyrics      string    `json:"lyrics"`
	CoverURL    string    `json:"cover_url"`
	ReleaseDate time.Time `json:"release_date"`
	Views       int       `json:"views"`
	Artist      Artist    `json:"artist"`
}

type PatchUpdateCoverOut struct {
	CoverURL string `json:"cover_url"`
}
