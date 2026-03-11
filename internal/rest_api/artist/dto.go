package artist

type GetArtistByIDIn struct {
	ArtistID int `param:"id" validate:"required"`
}

type GetArtistByIDOut struct {
	ArtistID  int    `json:"id"`
	Name      string `json:"name"`
	Bio       string `json:"bio"`
	AvatarURL string `json:"avatar_url"`
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
