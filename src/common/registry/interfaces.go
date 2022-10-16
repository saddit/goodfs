package registry

type (
	Registry interface {
		Register() error
		Unregister() error
	}
	Discovery interface {
		// GetServices server addresses
		GetServices(name string) []string
		// GetServiceMapping server id to address mapping
		GetServiceMapping(name string) map[string]string
	}
)
