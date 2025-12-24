package utils

import (
	"encoding/binary"
	"time"
)

func TimeToBinary(t *time.Time) []byte {
	ts := t.UnixNano()
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(ts))
	return buf
}

func BinaryToTime(b []byte) *time.Time {
	ts := int64(binary.BigEndian.Uint64(b))
	parsedTime := time.Unix(0, ts)
	return &parsedTime
}
