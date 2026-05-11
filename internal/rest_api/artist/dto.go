package artist

type GetArtistByIDIn struct {
	ArtistID int `param:"id" validate:"required"`
	Limit    int `query:"limit"`
	Offset   int `query:"offset"`
}

type GetArtistByIDOut struct {
	ArtistID  int    `json:"id"`
	Name      string `json:"name"`
	Bio       string `json:"bio"`
	AvatarURL string `json:"avatar_url"`
	Songs     []Song `json:"songs"`
}

type Song struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	CoverURL    string `json:"cover_url"`
	Views       int    `json:"views"`
	ReleaseDate string `json:"release_date"`
}

type PostArtistIn struct {
	Name string `json:"name" validate:"required"`
	Bio  string `json:"bio"`
}

type PostArtistOut struct {
	ArtistID  int    `json:"id"`
	Name      string `json:"name"`
	Bio       string `json:"bio"`
	AvatarURL string `json:"avatar_url"`
}

type PatchUpdateArtistIn struct {
	ArtistID int     `param:"id" validate:"required"`
	Name     *string `json:"name"`
	Bio      *string `json:"bio"`
}

type PatchUpdateArtistOut struct {
	ArtistID  int    `json:"id"`
	Name      string `json:"name"`
	Bio       string `json:"bio"`
	AvatarURL string `json:"avatar_url"`
}

type PatchUpdateAvatarOut struct {
	AvatarURL string `json:"avatar_url"`
}
