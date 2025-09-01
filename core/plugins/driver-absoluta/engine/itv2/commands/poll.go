package commands

type Poll struct {
}

func init() {
	registerCommand[Poll](1536)
}
