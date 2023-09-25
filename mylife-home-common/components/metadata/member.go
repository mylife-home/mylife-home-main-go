package metadata

type MemberType string

const (
	Action MemberType = "action"
	State  MemberType = "state"
)

type Member struct {
	name        string
	description string
	memberType  MemberType
	valueType   Type
}

func (member *Member) Name() string {
	return member.name
}

func (member *Member) Description() string {
	return member.description
}

func (member *Member) MemberType() MemberType {
	return member.memberType
}

func (member *Member) ValueType() Type {
	return member.valueType
}
