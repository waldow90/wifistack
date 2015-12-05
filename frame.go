package wifistack

import (
	"encoding/binary"
	"hash/crc32"
)

type Frame struct {
	Version         int
	Type            int
	Subtype         int
	FromDS          bool
	ToDS            bool
	MoreFrag        bool
	Retry           bool
	PowerManagement bool
	MoreData        bool
	Encrypted       bool
	Order           bool

	DurationID uint16
	MAC1       [6]byte
	MAC2       [6]byte
	MAC3       [6]byte

	SequenceControl uint16
	MAC4            [6]byte

	Payload []byte
}

func DecodeFrame(data []byte) (*Frame, error) {
	// TODO: some frames may be less than 28 bytes (I think).
	if len(data) < 28 {
		return nil, ErrBufferUnderflow
	}

	calculatedChecksum := crc32.ChecksumIEEE(data[:len(data)-4])
	sentChecksum := binary.LittleEndian.Uint32(data[len(data)-4:])
	if calculatedChecksum != sentChecksum {
		return nil, ErrBadChecksum
	}

	res := Frame{}
	res.Version = int(data[0]) & 3
	res.Type = int(data[0]>>2) & 3
	res.Subtype = int(data[0]>>4) & 0xf
	if (data[1] & 1) != 0 {
		res.FromDS = true
	}
	if (data[1] & 2) != 0 {
		res.ToDS = true
	}
	if (data[1] & 4) != 0 {
		res.MoreFrag = true
	}
	if (data[1] & 8) != 0 {
		res.Retry = true
	}
	if (data[1] & 0x10) != 0 {
		res.PowerManagement = true
	}
	if (data[1] & 0x20) != 0 {
		res.MoreData = true
	}
	if (data[1] & 0x40) != 0 {
		res.Encrypted = true
	}
	if (data[1] & 0x80) != 0 {
		res.Order = true
	}

	res.DurationID = binary.BigEndian.Uint16(data[2:])

	copy(res.MAC1[:], data[4:])
	copy(res.MAC2[:], data[10:])
	copy(res.MAC3[:], data[16:])

	res.SequenceControl = binary.BigEndian.Uint16(data[22:])

	// TODO: figure out if the packet has an extra MAC field.

	res.Payload = data[24 : len(data)-4]

	return &res, nil
}
