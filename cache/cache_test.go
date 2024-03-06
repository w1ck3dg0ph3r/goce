package cache_test

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/w1ck3dg0ph3r/goce/cache"
)

func TestCache(t *testing.T) {
	tmpdir := t.TempDir()
	c, err := cache.New[key, value](
		filepath.Join(tmpdir, "cache.db"),
		cache.WithCleanupInterval[key, value](500*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	t.Run("set", func(t *testing.T) {
		err := c.Set(key("aaa"), value{"foo", []string{"one", "two", "three"}}, 0)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("get", func(t *testing.T) {
		var v value
		ok, err := c.Get(key("aaa"), &v)
		if !ok {
			t.Errorf("expected to find key")
		}
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if !reflect.DeepEqual(v, value{"foo", []string{"one", "two", "three"}}) {
			t.Errorf("expected %v, got %v", "foo", v)
		}
	})

	t.Run("get not found", func(t *testing.T) {
		var v value
		ok, err := c.Get(key("bbb"), &v)
		if ok {
			t.Errorf("expected to not find key")
		}
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("set with ttl", func(t *testing.T) {
		err := c.Set(key("bbb"), value{"bar", []string{"one", "two", "three"}}, 1*time.Second)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("get with ttl", func(t *testing.T) {
		var v value
		ok, err := c.Get(key("bbb"), &v)
		if !ok {
			t.Errorf("expected to find key")
		}
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if !reflect.DeepEqual(v, value{"bar", []string{"one", "two", "three"}}) {
			t.Errorf("expected %v, got %v", "foo", v)
		}
	})

	t.Run("expire", func(t *testing.T) {
		time.Sleep(3 * time.Second)
		var v value
		ok, err := c.Get(key("bbb"), &v)
		if ok {
			t.Errorf("expected to not find key")
		}
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

type key string

func (k key) Hash() []byte {
	return []byte(k)
}

type value struct {
	Name   string
	Values []string
}
