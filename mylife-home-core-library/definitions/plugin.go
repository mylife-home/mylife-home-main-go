package definitions

type Plugin interface {
	Init(runtime Runtime) error
	Terminate()
}
