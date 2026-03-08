package roles

type Role string

const (
	RoleUser      Role = "user"
	RoleModerator Role = "moderator"
)

var RoleLevel = map[Role]int{
	RoleUser:      100,
	RoleModerator: 200,
}
