package datas

import (
	"flag"

	"github.com/attic-labs/noms/chunks"
	"github.com/attic-labs/noms/ref"
)

// DataStore provides versioned storage for noms values. Each DataStore instance represents one moment in history. Heads() returns the Commit from each active fork at that moment. The Commit() method returns a new DataStore, representing a new moment in history.
type DataStore interface {
	chunks.ChunkStore

	// MaybeHead returns the current Head Commit of this Datastore, which contains the current root of the DataStore's value tree, if available. If not, it returns a new Commit and 'false'.
	MaybeHead(datasetID string) (Commit, bool)

	// Head returns the current head Commit, which contains the current root of the DataStore's value tree.
	Head(datasetID string) Commit

	// Datasets returns the root of the datastore which is a MapOfStringToRefOfCommit where string is a datasetID.
	Datasets() MapOfStringToRefOfCommit

	// Commit updates the commit that a datastore points at. The new Commit is constructed using v and the current Head. If the update cannot be performed, e.g., because of a conflict, error will non-nil. The newest snapshot of the datastore is always returned.
	Commit(datasetID string, commit Commit) (DataStore, error)

	// Copies all chunks reachable from (and including)|r| but not reachable from (and including |exclude| in |source| to |sink|
	CopyReachableChunksP(r, exclude ref.Ref, sink chunks.ChunkSink, concurrency int)

	// Copies all chunks reachable from (and including) |r| in |source| that aren't present in |sink|
	CopyMissingChunksP(r ref.Ref, sink chunks.ChunkStore, concurrency int)
}

func NewDataStore(cs chunks.ChunkStore) DataStore {
	return newLocalDataStore(cs)
}

type Flags struct {
	ldb         chunks.LevelDBStoreFlags
	dynamo      chunks.DynamoStoreFlags
	hflags      chunks.HTTPStoreFlags
	memory      chunks.MemoryStoreFlags
	datastoreID *string
}

func NewFlags() Flags {
	return NewFlagsWithPrefix("")
}

func NewFlagsWithPrefix(prefix string) Flags {
	return Flags{
		chunks.LevelDBFlags(prefix),
		chunks.DynamoFlags(prefix),
		chunks.HTTPFlags(prefix),
		chunks.MemoryFlags(prefix),
		flag.String(prefix+"store", "", "name of datastore to access datasets in"),
	}
}

func (f Flags) CreateDataStore() (DataStore, bool) {
	var cs chunks.ChunkStore
	if cs = f.ldb.CreateStore(*f.datastoreID); cs != nil {
	} else if cs = f.dynamo.CreateStore(*f.datastoreID); cs != nil {
	} else if cs = f.memory.CreateStore(*f.datastoreID); cs != nil {
	}

	if cs != nil {
		return newLocalDataStore(cs), true
	}

	if cs = f.hflags.CreateStore(*f.datastoreID); cs != nil {
		return newRemoteDataStore(cs), true
	}

	return &LocalDataStore{}, false
}
