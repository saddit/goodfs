package cache

type ICache interface {
	NotifyEvicted() <-chan Entry
	Get(k string) []byte
	HasGet(k string) ([]byte, bool)
	Has(k string) bool
	Set(k string, v []byte) bool
	SetGob(k string, v interface{}) bool
	Refresh(k string)
	Delete(k string)
	Close() error
}
