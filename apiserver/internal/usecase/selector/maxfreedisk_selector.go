package selector

type MaxFreeDiskSelector struct{}

const MaxFreeDisk SelectStrategy = "maxfreedisk"

func (s *MaxFreeDiskSelector) Pop(ds []string) ([]string, string) {
	return ds[1:], ds[0]
}

func (s *MaxFreeDiskSelector) Select(ds []string) string {
	return ds[0]
}

func (s *MaxFreeDiskSelector) Strategy() SelectStrategy {
	return MaxFreeDisk
}
