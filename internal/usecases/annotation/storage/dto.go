package storage

import "time"

type AnnotationInfo struct {
	ID         int
	Song       SongInfo
	User       UserInfo
	Content    string
	StartIndex int
	EndIndex   int
	Rating     int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	MyVote     *int
}

type UserInfo struct {
	UserID          int
	Username        string
	AvatarURL       string
	ReputationScore int
}

type SongInfo struct {
	ID       int
	Title    string
	Artist   ArtistInfo
	CoverURL string
}

type ArtistInfo struct {
	ID   int
	Name string
}

type GetAnnotationsFilter struct {
	SongID int
	UserID *int // nil, если пользователь не авторизован
}

type GetAnnotationByIDFilter struct {
	AnnotationID int
	UserID       *int // nil, если пользователь не авторизован
}

type CreateAnnotationFilter struct {
	AuthorID   int
	SongID     int
	Content    string
	StartIndex int
	EndIndex   int
}

type UpdateAnnotationFilter struct {
	AnnotationID int
	UserID       int
	Content      string
}

type DeleteAnnotationFilter struct {
	AnnotationID int
	UserID       int
	Role         string
}

type VoteAnnotationFilter struct {
	AnnotationID int
	UserID       int
	Value        int // 1 или -1
}

type RemoveVoteFilter struct {
	AnnotationID int
	UserID       int
}

type GetUserAnnotationsFilter struct {
	UserID        int
	CurrentUserID *int
	Limit         int
	Offset        int
}
