package plugin

type runData struct {
	cancel func()
	exit   chan struct{}
}
