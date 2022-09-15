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
		case "location":
			z.Location, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Location")
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
		case "peers":
			var zb0003 uint32
			zb0003, err = dc.ReadArrayHeader()
			if err != nil {
				err = msgp.WrapError(err, "Peers")
				return
			}
			if cap(z.Peers) >= int(zb0003) {
				z.Peers = (z.Peers)[:zb0003]
			} else {
				z.Peers = make([]string, zb0003)
			}
			for za0002 := range z.Peers {
				z.Peers[za0002], err = dc.ReadString()
				if err != nil {
					err = msgp.WrapError(err, "Peers", za0002)
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
	// write "location"
	err = en.Append(0x83, 0xa8, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e)
	if err != nil {
		return
	}
	err = en.WriteString(z.Location)
	if err != nil {
		err = msgp.WrapError(err, "Location")
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
	// write "peers"
	err = en.Append(0xa5, 0x70, 0x65, 0x65, 0x72, 0x73)
	if err != nil {
		return
	}
	err = en.WriteArrayHeader(uint32(len(z.Peers)))
	if err != nil {
		err = msgp.WrapError(err, "Peers")
		return
	}
	for za0002 := range z.Peers {
		err = en.WriteString(z.Peers[za0002])
		if err != nil {
			err = msgp.WrapError(err, "Peers", za0002)
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *SlotInfo) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "location"
	o = append(o, 0x83, 0xa8, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e)
	o = msgp.AppendString(o, z.Location)
	// string "slots"
	o = append(o, 0xa5, 0x73, 0x6c, 0x6f, 0x74, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Slots)))
	for za0001 := range z.Slots {
		o = msgp.AppendString(o, z.Slots[za0001])
	}
	// string "peers"
	o = append(o, 0xa5, 0x70, 0x65, 0x65, 0x72, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Peers)))
	for za0002 := range z.Peers {
		o = msgp.AppendString(o, z.Peers[za0002])
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
		case "location":
			z.Location, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Location")
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
		case "peers":
			var zb0003 uint32
			zb0003, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Peers")
				return
			}
			if cap(z.Peers) >= int(zb0003) {
				z.Peers = (z.Peers)[:zb0003]
			} else {
				z.Peers = make([]string, zb0003)
			}
			for za0002 := range z.Peers {
				z.Peers[za0002], bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "Peers", za0002)
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
	s = 1 + 9 + msgp.StringPrefixSize + len(z.Location) + 6 + msgp.ArrayHeaderSize
	for za0001 := range z.Slots {
		s += msgp.StringPrefixSize + len(z.Slots[za0001])
	}
	s += 6 + msgp.ArrayHeaderSize
	for za0002 := range z.Peers {
		s += msgp.StringPrefixSize + len(z.Peers[za0002])
	}
	return
}
