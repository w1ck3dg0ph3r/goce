package store

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"

	"github.com/w1ck3dg0ph3r/goce/compilers"
	"github.com/w1ck3dg0ph3r/goce/parsers"
	"github.com/w1ck3dg0ph3r/goce/pkg/cache"
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
	var sum [32]byte
	h := sha256.New()
	br := bufio.NewWriter(h)
	_, _ = br.WriteString(k.CompilerName)
	je := json.NewEncoder(br)
	_ = je.Encode(k.CompilerOptions)
	_, _ = br.Write(k.Code)
	_ = br.Flush()
	return h.Sum(sum[:0])
}
