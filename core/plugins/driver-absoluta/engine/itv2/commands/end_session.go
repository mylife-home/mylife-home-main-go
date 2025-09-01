package commands

type EndSession struct {
}

func init() {
	registerCommand[EndSession](1547)
}
