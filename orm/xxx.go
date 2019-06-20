package orm

import (
	"context"
	"fmt"

	"github.com/iov-one/weave"
)

type UnboundModelBucket interface {
	Bind(context.Context) ModelBucket
}

func WithLastModified(b ModelBucket) UnboundModelBucket {
	return &unboundLastModifiedBucket{bucket: b}
}

type unboundLastModifiedBucket struct {
	bucket ModelBucket
}

func (u *unboundLastModifiedBucket) Bind(ctx context.Context) ModelBucket {
	blockHeight, ok := weave.GetHeight(ctx)
	if !ok {
		panic("block height not present in the context")
	}
	return &lastModifiedBucket{
		blockHeight: blockHeight,
		bucket:      u.bucket,
	}
}

type lastModifiedBucket struct {
	blockHeight int64
	bucket      ModelBucket
}

func (b *lastModifiedBucket) One(db weave.ReadOnlyKVStore, key []byte, dest Model) error {
	return b.bucket.One(db, key, dest)
}

func (b *lastModifiedBucket) ByIndex(db weave.ReadOnlyKVStore, indexName string, key []byte, dest ModelSlicePtr) (keys [][]byte, err error) {
	return b.bucket.ByIndex(db, indexName, key, dest)
}

func (b *lastModifiedBucket) Put(db weave.KVStore, key []byte, m Model) ([]byte, error) {
	type metadator interface {
		GetMetadata() *weave.Metadata
	}
	if m, ok := m.(metadator); ok {
		meta := m.GetMetadata()
		// TODO: set block height
		fmt.Println("meta = b.blockHeight", meta)
	}
	return b.bucket.Put(db, key, m)
}

func (b *lastModifiedBucket) Delete(db weave.KVStore, key []byte) error {
	return b.bucket.Delete(db, key)
}

func (b *lastModifiedBucket) Has(db weave.KVStore, key []byte) error {
	return b.bucket.Has(db, key)
}

func (b *lastModifiedBucket) Register(name string, r weave.QueryRouter) {
	b.bucket.Register(name, r)
}
