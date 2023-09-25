package definitions

type Plugin interface {
	Init() error
	Terminate()
}
