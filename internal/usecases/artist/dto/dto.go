package dto

type Artist struct {
	ArtistID          int
	Name              string
	Bio               string
	AvatarURL         string
}

type PostArtistOpts struct {
	Name string
	Bio  string
}