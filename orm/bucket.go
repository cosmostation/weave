/*
package orm provides an easy to use db wrapper

Break state space into prefixed sections called Buckets.
* Each bucket contains only one type of object.
* It has a primary index (which may be composite),
and may possess secondary indexes.
* It may possess one or more secondary indexes (1:1 or 1:N)
* Easy queries for one and iteration.

For inspiration, look at [storm](https://github.com/asdine/storm) built on top of [bolt kvstore](https://github.com/boltdb/bolt#using-buckets).
* Do not use so much reflection magic. Better do stuff compile-time static, even if it is a bit of boilerplate.
* Consider general usability flow from that project
*/
package orm

import (
	"fmt"
	"regexp"

	"github.com/confio/weave"
	"github.com/pkg/errors"
)

const (
	// SeqID is a constant to use to get a default ID sequence
	SeqID = "id"
)

var (
	isBucketName = regexp.MustCompile(`^[a-z_]{3,8}$`).MatchString
)

// Bucket is a generic holder that stores data as well
// as references to secondary indexes and sequences.
//
// This is a generic building block that should generally
// be embedded in a type-safe wrapper to ensure all data
// is the same type.
// Bucket is a prefixed subspace of the DB
// proto defines the default Model, all elements of this type
type Bucket struct {
	name    string
	prefix  []byte
	proto   Cloneable
	indexes map[string]Index
}

// NewBucket creates a bucket to store data
func NewBucket(name string, proto Cloneable) Bucket {
	if !isBucketName(name) {
		panic(fmt.Sprintf("Illegal bucket: %s", name))
	}

	return Bucket{
		name:   name,
		prefix: append([]byte(name), ':'),
		proto:  proto,
	}
}

// DBKey is the full key we store in the db, including prefix
func (b Bucket) DBKey(key []byte) []byte {
	return append(b.prefix, key...)
}

// Get one element
func (b Bucket) Get(db weave.KVStore, key []byte) (Object, error) {
	dbkey := b.DBKey(key)
	bz := db.Get(dbkey)
	if bz == nil {
		return nil, nil
	}

	obj := b.proto.Clone()
	err := obj.Value().Unmarshal(bz)
	if err != nil {
		return nil, err
	}
	obj.SetKey(key)
	return obj, nil
}

// Save will write a model, it must be of the same type as proto
func (b Bucket) Save(db weave.KVStore, model Object) error {
	err := model.Validate()
	if err != nil {
		return err
	}

	bz, err := model.Value().Marshal()
	if err != nil {
		return err
	}
	err = b.updateIndexes(db, model.Key(), model)
	if err != nil {
		return err
	}

	// now save this one
	dbkey := append(b.prefix, model.Key()...)
	db.Set(dbkey, bz)
	return nil
}

// Delete will remove the value at a key
func (b Bucket) Delete(db weave.KVStore, key []byte) error {
	err := b.updateIndexes(db, key, nil)
	if err != nil {
		return err
	}

	// now save this one
	dbkey := b.DBKey(key)
	db.Delete(dbkey)
	return nil
}

func (b Bucket) updateIndexes(db weave.KVStore, key []byte, model Object) error {
	// update all indexes
	if len(b.indexes) > 0 {
		prev, err := b.Get(db, key)
		if err != nil {
			return err
		}
		for _, idx := range b.indexes {
			err = idx.Update(db, prev, model)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Sequence returns a Sequence by name
func (b Bucket) Sequence(name string) Sequence {
	id := append(b.prefix, []byte(name)...)
	return NewSequence(id)
}

// WithIndex returns a copy of this bucket with given index,
// panics if it an index with that name is already registered.
//
// Designed to be chained.
func (b Bucket) WithIndex(name string, indexer Indexer, unique bool) Bucket {
	// no duplicate indexes! (panic on init)
	if _, ok := b.indexes[name]; ok {
		panic(fmt.Sprintf("Index %s registered twice", name))
	}

	iname := b.name + "_" + name
	add := NewIndex(iname, indexer, unique)
	indexes := make(map[string]Index, len(b.indexes)+1)
	for n, i := range b.indexes {
		indexes[n] = i
	}
	indexes[name] = add
	b.indexes = indexes
	return b
}

// GetIndexed querys the named index for the given key
func (b Bucket) GetIndexed(db weave.KVStore, name string, key []byte) ([]Object, error) {
	idx, ok := b.indexes[name]
	if !ok {
		return nil, errors.Errorf("No such index: %s", name)
	}
	refs, err := idx.GetAt(db, key)
	if err != nil {
		return nil, err
	}
	return b.readRefs(db, refs)
}

// GetIndexedLike querys the named index with the given pattern
func (b Bucket) GetIndexedLike(db weave.KVStore, name string, pattern Object) ([]Object, error) {
	idx, ok := b.indexes[name]
	if !ok {
		return nil, errors.Errorf("No such index: %s", name)
	}
	refs, err := idx.GetLike(db, pattern)
	if err != nil {
		return nil, err
	}
	return b.readRefs(db, refs)
}

func (b Bucket) readRefs(db weave.KVStore, refs [][]byte) ([]Object, error) {
	if len(refs) == 0 {
		return nil, nil
	}

	var err error
	objs := make([]Object, len(refs))
	for i, key := range refs {
		objs[i], err = b.Get(db, key)
		if err != nil {
			return nil, err
		}
	}
	return objs, nil
}
