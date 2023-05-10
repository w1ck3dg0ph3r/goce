package main

import (
	"github.com/w1ck3dg0ph3r/goce/cache"
	"github.com/w1ck3dg0ph3r/goce/shortid"
)

type SharedCodeKey struct {
	id shortid.ID
}

type SharedCodeValue struct {
	Code []byte
}

type SharedCodeStore = cache.Cache[SharedCodeKey, SharedCodeValue]

func NewSharedStore(filename string) (*SharedCodeStore, error) {
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
