package user

type GetUserMeOut struct {
	UserID          int    `json:"user_id"`
	Email           string `json:"email"`
	Username        string `json:"username"`
	Bio             string `json:"bio"`
	AvatarURL       string `json:"avatar_url"`
	ReputationScore int    `json:"reputation_score"`
	Role            string `json:"role"`
}

type GetUserByIDIn struct {
	UserID int `param:"id" validate:"required"`
}

type GetUserByIDOut struct {
	UserID          int    `json:"user_id"`
	Email           string `json:"email"`
	Username        string `json:"username"`
	Bio             string `json:"bio"`
	AvatarURL       string `json:"avatar_url"`
	ReputationScore int    `json:"reputation_score"`
	Role            string `json:"role"`
}

type PatchUpdateUserIn struct {
	Email    *string `json:"email"`
	Username *string `json:"username"`
	Bio      *string `json:"bio"`
	Password *string `json:"password"`
}

type PatchUpdateAvatarIn struct {
	AvatarURL string `json:"avatar_url" validate:"required,url"`
}

type PostAvatarUploadURLIn struct {
	FileExt     string `json:"file_ext" validate:"required"`
	ContentType string `json:"content_type" validate:"required"`
}

type PostAvatarUploadURLOut struct {
	UploadURL    string            `json:"upload_url"`
	AvatarURL    string            `json:"avatar_url"`
	ObjectKey    string            `json:"object_key"`
	Method       string            `json:"method"`
	ExpiresInSec int               `json:"expires_in_sec"`
	Headers      map[string]string `json:"headers"`
}

type UploadURLInfo struct {
	UploadURL    string
	AvatarURL    string
	ObjectKey    string
	Method       string
	ExpiresInSec int
	Headers      map[string]string
}
