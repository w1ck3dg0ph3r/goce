package cache

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

type Key interface {
	Hash() []byte
}

type Cache[K Key, V any] struct {
	db *bbolt.DB

	doneCh chan struct{}
	wg     sync.WaitGroup
}

func New[K Key, V any](filename string) (*Cache[K, V], error) {
	db, err := bbolt.Open(filename, 0o660, &bbolt.Options{
		NoFreelistSync: true,
		FreelistType:   bbolt.FreelistMapType,
	})
	if err != nil {
		return nil, fmt.Errorf("open cache file: %w", err)
	}

	cache := &Cache[K, V]{
		db:     db,
		doneCh: make(chan struct{}),
	}

	if err := cache.createBuckets(); err != nil {
		return nil, fmt.Errorf("create buckets: %w", err)
	}

	cache.startCleanup(1 * time.Minute)

	return cache, nil
}

func (cache *Cache[K, V]) Get(k K, v *V) bool {
	found := false
	err := cache.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		key := k.Hash()
		value := b.Get(key)
		found = value != nil
		if found {
			unmarshal(value, v)
		}
		return nil
	})
	if err != nil {
		return false
	}
	return found
}

func (cache *Cache[K, V]) Set(k Key, v *V, ttl time.Duration) error {
	err := cache.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		key := k.Hash()
		value := marshal(v)
		if err := b.Put(key[:], value); err != nil {
			return err
		}

		if ttl > 0 {
			ttlb := tx.Bucket(ttlBucketName)
			expiry := time.Now().Add(ttl)
			var ttlKey [8]byte
			binary.BigEndian.PutUint64(ttlKey[:], uint64(expiry.Unix()))
			if err := ttlb.Put(ttlKey[:], key[:]); err != nil {
				return err
			}
		}

		return nil
	})
	return err
}

func (cache *Cache[K, V]) createBuckets() error {
	err := cache.db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(bucketName); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(ttlBucketName); err != nil {
			return err
		}
		return nil
	})
	return err
}

func (cache *Cache[K, V]) startCleanup(interval time.Duration) {
	cache.wg.Add(1)
	go func() {
		defer cache.wg.Done()
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ticker.C:
				_ = cache.cleanup()
			case <-cache.doneCh:
				return
			}
		}
	}()
}

func (cache *Cache[K, V]) cleanup() error {
	var nowKey [8]byte
	binary.BigEndian.PutUint64(nowKey[:], uint64(time.Now().Unix()))

	var keysToDelete [][]byte
	var ttlKeysToDelete [][]byte

	var err error

	err = cache.db.View(func(tx *bbolt.Tx) error {
		ttlb := tx.Bucket(ttlBucketName)
		c := ttlb.Cursor()
		k, v := c.First()
		for k != nil {
			if bytes.Compare(k, nowKey[:]) >= 0 {
				break
			}
			keysToDelete = append(keysToDelete, v)
			ttlKeysToDelete = append(ttlKeysToDelete, k)
			k, v = c.Next()
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = cache.db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		for _, k := range keysToDelete {
			if err := b.Delete(k); err != nil {
				return err
			}
		}
		ttlb := tx.Bucket(ttlBucketName)
		for _, k := range ttlKeysToDelete {
			if err := ttlb.Delete(k); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (cache *Cache[K, V]) Close() {
	close(cache.doneCh)
	cache.wg.Wait()
	_ = cache.db.Close()
}

var (
	bucketName    = []byte("c")
	ttlBucketName = []byte("t")
)

func marshal(e any) []byte {
	b := &bytes.Buffer{}
	w := gob.NewEncoder(b)
	if err := w.Encode(e); err != nil {
		panic(err)
	}
	return b.Bytes()
}

func unmarshal(b []byte, e any) {
	r := gob.NewDecoder(bytes.NewReader(b))
	if err := r.Decode(e); err != nil {
		panic(err)
	}
}
