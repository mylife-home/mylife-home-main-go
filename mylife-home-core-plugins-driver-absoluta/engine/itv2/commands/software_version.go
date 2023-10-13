package commands

type SoftwareVersion struct {
	VersionFields [11]byte // BCD string
}

func init() {
	registerCommand[SoftwareVersion](1549)
}
