package definitions

type Plugin interface {
	Init(runtime Runtime) error
	Terminate() // Note: will be executed even if Init() returns an error
}
