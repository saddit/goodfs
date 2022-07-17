package registry

type (
	Registry interface {
		Register() error
		Unregister() error
	}
	Discovery interface {
		GetServices(name string) []string
	}
)
