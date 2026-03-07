package roles

type Role string

const (
	RoleUser      Role = "user"
	RoleModerator Role = "moderator"
	RoleAdmin     Role = "admin"
)

var RoleLevel = map[Role]int{
	RoleUser:      100,
	RoleModerator: 200,
	RoleAdmin:     300,
}
