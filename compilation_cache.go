package main

import (
	"bufio"
	"crypto/md5"

	"github.com/w1ck3dg0ph3r/goce/cache"
	"github.com/w1ck3dg0ph3r/goce/parsers"
)

type CompilationCacheKey struct {
	CompilerName string
	Code         []byte
}

type CompilationCacheValue = parsers.Result

type CompilationCache = cache.Cache[CompilationCacheKey, CompilationCacheValue]

func NewCompilationCache(filename string) (*CompilationCache, error) {
	return cache.New[CompilationCacheKey, CompilationCacheValue](filename)
}

func (k CompilationCacheKey) Hash() []byte {
	var sum [16]byte
	h := md5.New()
	br := bufio.NewWriter(h)
	_, _ = br.WriteString(k.CompilerName)
	_, _ = br.Write(k.Code)
	_ = br.Flush()
	return h.Sum(sum[:0])
}
