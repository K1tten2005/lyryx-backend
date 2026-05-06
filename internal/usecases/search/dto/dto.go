package dto

type GetSearchOpts struct {
	Query string
	Limit int
}

type SearchResult struct {
	Songs              []SongInfo
	LyricsMatchedSongs []SongInfo
	Artists            []ArtistInfo
	Users              []UserInfo
}

type SongInfo struct {
	ID            int
	Title         string
	LyricsSnippet string
	Artist        ArtistInfo
	CoverURL      string
}

type ArtistInfo struct {
	ID        int
	Name      string
	AvatarURL string
}

type UserInfo struct {
	UserID          int
	Username        string
	AvatarURL       string
	ReputationScore int
}
