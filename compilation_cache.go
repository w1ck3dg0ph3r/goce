package main

import (
	"bufio"
	"crypto/md5"
	"encoding/json"

	"github.com/w1ck3dg0ph3r/goce/cache"
	"github.com/w1ck3dg0ph3r/goce/compilers"
	"github.com/w1ck3dg0ph3r/goce/parsers"
)

type CompilationCacheKey struct {
	CompilerName    string
	CompilerOptions compilers.CompilerOptions
	Code            []byte
}

type CompilationCacheValue struct {
	BuildFailed bool
	BuildOutput string
	parsers.Result
}

type CompilationCache = cache.Cache[CompilationCacheKey, CompilationCacheValue]

func NewCompilationCache(filename string) (*CompilationCache, error) {
	return cache.New[CompilationCacheKey, CompilationCacheValue](filename)
}

func (k CompilationCacheKey) Hash() []byte {
	var sum [16]byte
	h := md5.New()
	br := bufio.NewWriter(h)
	_, _ = br.WriteString(k.CompilerName)
	je := json.NewEncoder(br)
	_ = je.Encode(k.CompilerOptions)
	_, _ = br.Write(k.Code)
	_ = br.Flush()
	return h.Sum(sum[:0])
}
