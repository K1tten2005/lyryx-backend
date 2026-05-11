package dto

import "time"

type GetAnnotationOpts struct {
	SongID int
	UserID *int
}

type GetAnnotationByIDOpts struct {
	AnnotationID int
	UserID       *int
}

type AnnotationInfo struct {
	ID         int
	Song       SongInfo
	User       UserInfo
	Content    string
	StartIndex int
	EndIndex   int
	Snippet    string
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

type PostAnnotationOpts struct {
	SongID     int
	UserID     int
	Content    string
	StartIndex int
	EndIndex   int
}

type PatchUpdateAnnotationOpts struct {
	AnnotationID int
	Content      string
	UserID       int
}

type DeleteAnnotationOpts struct {
	AnnotationID int
	UserID       int
	Role         string
}

type PostVoteAnnotationOpts struct {
	AnnotationID int
	Value        int
	UserID       int
}

type DeleteVoteOpts struct {
	AnnotationID int
	UserID       int
}

type GetUserAnnotationsOpts struct {
	UserID        int
	CurrentUserID *int
	Limit         int
	Offset        int
}

type GetAiAnnotationOpts struct {
	SongID     int
	Question   string
	StartIndex *int
	EndIndex   *int
}

type AiAnnotationResp struct {
	Response string
}
