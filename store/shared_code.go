package store

import (
	"github.com/w1ck3dg0ph3r/goce/pkg/cache"
	"github.com/w1ck3dg0ph3r/goce/pkg/shortid"
)

type SharedCodeKey struct {
	id shortid.ID
}

type SharedCodeValue struct {
	Code []byte
}

type SharedCode = cache.Cache[SharedCodeKey, SharedCodeValue]

func NewSharedCode(filename string) (*SharedCode, error) {
	return cache.New[SharedCodeKey, SharedCodeValue](filename)
}

func NewSharedCodeKey() SharedCodeKey {
	return SharedCodeKey{id: shortid.New()}
}

func ParseSharedCodeKey(s string) (SharedCodeKey, error) {
	id, err := shortid.Parse(s)
	if err != nil {
		return SharedCodeKey{}, err
	}
	return SharedCodeKey{id: id}, nil
}

func (k SharedCodeKey) String() string {
	return k.id.String()
}

func (k SharedCodeKey) Hash() []byte {
	return k.id[:]
}
