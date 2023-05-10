package shortid

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

type ID [6]byte

func New() ID {
	var id ID
	binary.BigEndian.PutUint32(id[0:4], uint32(time.Now().UnixMilli()))
	binary.BigEndian.PutUint16(id[4:6], uint16(atomic.AddUint32(&counter, 1)))
	return id
}

func Parse(s string) (ID, error) {
	var id ID
	b := decode(s)
	if len(b) != len(id) {
		return id, ErrInvalid
	}
	return ID(b), nil
}

func (id ID) String() string {
	return encode(id[:])
}

var (
	ErrInvalid = errors.New("invalid shortid")

	counter = readRandomUint32()
)

func readRandomUint32() uint32 {
	var b [4]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(fmt.Errorf("cannot initialize random counter with crypto/rand.Reader: %w", err))
	}
	return (uint32(b[0]) << 0) | (uint32(b[1]) << 8) | (uint32(b[2]) << 16) | (uint32(b[3]) << 24)
}
