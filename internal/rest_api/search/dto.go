package search

type GetSearchIn struct {
	Query  string `query:"q" validate:"required"`
	Limit  int    `query:"limit"`
}

type GetSearchOut struct {
	Songs              []SongInfo   `json:"songs"`
	LyricsMatchedSongs []SongInfo   `json:"lyrics_matched_songs"`
	Artists            []ArtistInfo `json:"artists"`
	Users              []UserInfo   `json:"users"`
}

type SongInfo struct {
	ID            int        `json:"id"`
	Title         string     `json:"title"`
	LyricsSnippet string     `json:"lyrics_snippet"`
	Artist        ArtistInfo `json:"artist"`
	CoverURL      string     `json:"cover_url,omitempty"`
}

type ArtistInfo struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type UserInfo struct {
	UserID          int    `json:"user_id"`
	Username        string `json:"username"`
	AvatarURL       string `json:"avatar_url,omitempty"`
	ReputationScore int    `json:"reputation_score,omitempty"`
}
