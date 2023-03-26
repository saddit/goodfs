package registry

type (
	Discovery interface {
		// GetServices server addresses
		GetServices(name string) []string
		// GetServiceMapping server id to address mapping
		GetServiceMapping(name string) map[string]string
		GetService(name string, id string) (string, bool)
	}
)
