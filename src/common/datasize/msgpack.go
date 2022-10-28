package datasize

import "github.com/tinylib/msgp/msgp"

func (d *DataSize) DecodeMsg(ds *msgp.Reader) error {
	v, err := ds.ReadUint64()
	*d = DataSize(v)
	return err
}

func (d *DataSize) EncodeMsg(en *msgp.Writer) error {
	return en.WriteUint64(uint64(*d))
}

func (d *DataSize) MarshalMsg(o []byte) ([]byte, error) {
	o = msgp.AppendUint64(o, uint64(*d))
	return o, nil
}

func (d *DataSize) UnmarshalMsg(bts []byte) ([]byte, error) {
	v, bts, err := msgp.ReadUint64Bytes(bts)
	*d = DataSize(v)
	return bts, err
}

func (d *DataSize) Msgsize() int {
	return msgp.Uint64Size
}
