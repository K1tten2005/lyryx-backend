package annotation

type GetAnnotationsIn struct {
	SongID int `param:"id" validate:"required"`
}

type GetAnnotationsOut struct {
	SongID      int          `json:"song_id"`
	Annotations []Annotation `json:"annotations"`
}

type Annotation struct {
	ID         int      `json:"id"`
	User       UserInfo `json:"user"`
	Content    string   `json:"content"`
	StartIndex int      `json:"start_index"`
	EndIndex   int      `json:"end_index"`
	Rating     int      `json:"rating"`
	CreatedAt  string   `json:"created_at"`
	MyVote     *int     `json:"my_vote"` // -1, 1 или null (опционально)
}

type UserInfo struct {
	UserID          int    `json:"user_id"`
	Username        string `json:"username"`
	AvatarURL       string `json:"avatar_url,omitempty"`
	ReputationScore int    `json:"reputation_score,omitempty"`
}

type GetAnnotationByIDIn struct {
	AnnotationID int `param:"id" validate:"required"`
}

type GetAnnotationByIDOut struct {
	ID         int      `json:"id"`
	Song       SongInfo `json:"song"`
	User       UserInfo `json:"user"`
	Content    string   `json:"content"`
	StartIndex int      `json:"start_index"`
	EndIndex   int      `json:"end_index"`
	Rating     int      `json:"rating"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
	MyVote     *int     `json:"my_vote"`
}

type SongInfo struct {
	ID       int        `json:"id"`
	Title    string     `json:"title"`
	Artist   ArtistInfo `json:"artist"`
	CoverURL string     `json:"cover_url,omitempty"`
}

type ArtistInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type PostAnnotationIn struct {
	SongID     int    `param:"id" validate:"required"`
	Content    string `json:"content" validate:"required"`
	StartIndex *int   `json:"start_index" validate:"required"`
	EndIndex   *int   `json:"end_index" validate:"required"`
}

type PostAnnotationOut struct {
	ID         int      `json:"id"`
	SongID     int      `json:"song_id"`
	User       UserInfo `json:"user"`
	Content    string   `json:"content"`
	StartIndex int      `json:"start_index"`
	EndIndex   int      `json:"end_index"`
	Rating     int      `json:"rating"`
	CreatedAt  string   `json:"created_at"`
}

type PatchUpdateAnnotationIn struct {
	AnnotationID int     `param:"id" validate:"required"`
	Content      *string `json:"content,omitempty" validate:"omitempty"`
}

type PatchUpdateAnnotationOut struct {
	ID        int    `json:"id"`
	Content   string `json:"content"`
	UpdatedAt string `json:"updated_at"`
	Rating    int    `json:"rating"`
}

type DeleteAnnotationIn struct {
	AnnotationID int `param:"id" validate:"required"`
}

type PostVoteAnnotationIn struct {
	AnnotationID int `param:"id" validate:"required"`
	Value        int `json:"value" validate:"required"`
}

type PostVoteAnnotationOut struct {
	AnnotationID int  `json:"annotation_id"`
	NewRating    int  `json:"new_rating"`
	MyVote       *int `json:"my_vote"`
}

type DeleteVoteIn struct {
	AnnotationID int `param:"id" validate:"required"`
}

type GetUserAnnotationsIn struct {
	UserID int `param:"id" validate:"required"`
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}

type UserAnnotation struct {
	ID         int      `json:"id"`
	Song       SongInfo `json:"song"`
	Content    string   `json:"content"`
	StartIndex int      `json:"start_index"`
	EndIndex   int      `json:"end_index"`
	Snippet    string   `json:"snippet"`
	Rating     int      `json:"rating"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
	MyVote     *int     `json:"my_vote"`
}

type GetUserAnnotationsOut struct {
	UserID      int              `json:"user_id"`
	Annotations []UserAnnotation `json:"annotations"`
	Total       int              `json:"total"`
	HasMore     bool             `json:"has_more"`
}

type GetAiAnnotationIn struct {
	SongID     int    `param:"id" validate:"required"`
	Question   string `query:"question" validate:"required"`
	StartIndex *int   `query:"start_index" validate:"required"`
	EndIndex   *int   `query:"end_index" validate:"required"`
}

type GetAiAnnotationOut struct {
	SongID   int    `json:"song_id"`
	Response string `json:"response"`
}
