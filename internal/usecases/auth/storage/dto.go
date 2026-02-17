package storage

type SignUpFilter struct {
	Username       string
	Email          string
	HashedPassword string
}

type UserInfo struct {
	UserID     int
	Email      string
	Username   string
	Role       string
}
