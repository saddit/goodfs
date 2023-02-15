package hashslot

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *SlotInfo) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "id":
			z.GroupID, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "GroupID")
				return
			}
		case "server_id":
			z.ServerID, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "ServerID")
				return
			}
		case "slots":
			var zb0002 uint32
			zb0002, err = dc.ReadArrayHeader()
			if err != nil {
				err = msgp.WrapError(err, "Slots")
				return
			}
			if cap(z.Slots) >= int(zb0002) {
				z.Slots = (z.Slots)[:zb0002]
			} else {
				z.Slots = make([]string, zb0002)
			}
			for za0001 := range z.Slots {
				z.Slots[za0001], err = dc.ReadString()
				if err != nil {
					err = msgp.WrapError(err, "Slots", za0001)
					return
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *SlotInfo) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "id"
	err = en.Append(0x83, 0xa2, 0x69, 0x64)
	if err != nil {
		return
	}
	err = en.WriteString(z.GroupID)
	if err != nil {
		err = msgp.WrapError(err, "GroupID")
		return
	}
	// write "server_id"
	err = en.Append(0xa9, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x5f, 0x69, 0x64)
	if err != nil {
		return
	}
	err = en.WriteString(z.ServerID)
	if err != nil {
		err = msgp.WrapError(err, "ServerID")
		return
	}
	// write "slots"
	err = en.Append(0xa5, 0x73, 0x6c, 0x6f, 0x74, 0x73)
	if err != nil {
		return
	}
	err = en.WriteArrayHeader(uint32(len(z.Slots)))
	if err != nil {
		err = msgp.WrapError(err, "Slots")
		return
	}
	for za0001 := range z.Slots {
		err = en.WriteString(z.Slots[za0001])
		if err != nil {
			err = msgp.WrapError(err, "Slots", za0001)
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *SlotInfo) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "id"
	o = append(o, 0x83, 0xa2, 0x69, 0x64)
	o = msgp.AppendString(o, z.GroupID)
	// string "server_id"
	o = append(o, 0xa9, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x5f, 0x69, 0x64)
	o = msgp.AppendString(o, z.ServerID)
	// string "slots"
	o = append(o, 0xa5, 0x73, 0x6c, 0x6f, 0x74, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Slots)))
	for za0001 := range z.Slots {
		o = msgp.AppendString(o, z.Slots[za0001])
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *SlotInfo) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "id":
			z.GroupID, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "GroupID")
				return
			}
		case "server_id":
			z.ServerID, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "ServerID")
				return
			}
		case "slots":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Slots")
				return
			}
			if cap(z.Slots) >= int(zb0002) {
				z.Slots = (z.Slots)[:zb0002]
			} else {
				z.Slots = make([]string, zb0002)
			}
			for za0001 := range z.Slots {
				z.Slots[za0001], bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "Slots", za0001)
					return
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *SlotInfo) Msgsize() (s int) {
	s = 1 + 3 + msgp.StringPrefixSize + len(z.GroupID) + 10 + msgp.StringPrefixSize + len(z.ServerID) + 6 + msgp.ArrayHeaderSize
	for za0001 := range z.Slots {
		s += msgp.StringPrefixSize + len(z.Slots[za0001])
	}
	return
}
