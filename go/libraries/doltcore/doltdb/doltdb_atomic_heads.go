// DoltDB wrappers for the typed atomic-publication primitive: a thin
// forwarding layer that makes no decisions and carries no test hooks and no
// mutable package globals.

package doltdb

import (
	"context"
	"errors"

	"github.com/nustiueudinastea/dolt/go/store/datas"
	"github.com/nustiueudinastea/dolt/go/store/hash"
)

// atomicHeadsPublisher is satisfied by the concrete datas database. Asserting
// it here keeps the public Database interface unchanged.
type atomicHeadsPublisher interface {
	SetHeadsAtRootWithWorkingSets(
		ctx context.Context,
		expectedRoot hash.Hash,
		commitEdits []datas.CommitHeadEdit,
		workingSetEdits []datas.WorkingSetHeadEdit,
	) (datas.CommitHeadBatchResult, error)
}

// SetHeadsAtRootWithWorkingSets publishes commit heads and working-set heads in
// one expected-root, all-or-none publication.
func (ddb *DoltDB) SetHeadsAtRootWithWorkingSets(
	ctx context.Context,
	expectedRoot hash.Hash,
	commitEdits []datas.CommitHeadEdit,
	workingSetEdits []datas.WorkingSetHeadEdit,
) (datas.CommitHeadBatchResult, error) {
	publisher, ok := ddb.db.Database.(atomicHeadsPublisher)
	if !ok {
		return datas.CommitHeadBatchResult{},
			errors.New("atomic head publication prototype is not present in this build")
	}
	return publisher.SetHeadsAtRootWithWorkingSets(ctx, expectedRoot, commitEdits, workingSetEdits)
}

// WriteCleanWorkingSet writes a merge/rebase-free working-set value
// at the given roots with the CALLER'S metadata, and returns its address. The
// metadata is part of the object's content address, so the caller — not this
// wrapper and not the store — decides it.
func (ddb *DoltDB) WriteCleanWorkingSet(
	ctx context.Context, workingRoot, stagedRoot hash.Hash, meta *datas.WorkingSetMeta,
) (hash.Hash, error) {
	return datas.WriteCleanWorkingSetValue(
		ctx, ddb.db.Database, workingRoot, stagedRoot, meta)
}
