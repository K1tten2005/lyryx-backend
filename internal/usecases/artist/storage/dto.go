package storage

type Artist struct {
	ArtistID          int
	Name              string
	Bio               string
	AvatarURL         string
}

type CreateArtistFilter struct {
	Name        string
	Bio         string
}