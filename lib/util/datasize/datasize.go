package datasize

type DataSize int

const (
	KB DataSize = 1024
	MB          = 1024 * KB
	GB          = 1024 * MB
)

func (d DataSize) IntValue() int {
	return int(d)
}

func (d DataSize) Int64Value() int64 {
	return int64(d)
}
