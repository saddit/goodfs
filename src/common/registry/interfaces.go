package registry

type (
	Registry interface {
		Register() error
		Unregister() error
	}
	Discovery interface {
		// GetServices server addresses
		GetServices(name string, rpc bool) []string
		// GetServiceMapping server id to address mapping
		GetServiceMapping(name string, rpc bool) map[string]string
		GetService(name string, id string, rpc bool) (string, bool)
	}
)
