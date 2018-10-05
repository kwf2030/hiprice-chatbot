package conv

import (
  "encoding/binary"
)

func Uint64ToBytes(i uint64) []byte {
  b := make([]byte, 8)
  binary.BigEndian.PutUint64(b, i)
  return b
}

func BytesToUint64(b []byte) uint64 {
  return binary.BigEndian.Uint64(b)
}
