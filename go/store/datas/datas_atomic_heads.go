// Atomic publication of commit heads together with working-set heads.
//
// This adds one typed sibling of SetCommitHeadsAtRoot that also carries
// working-set edits. It has no test hooks and no mutable package globals.
//
// The existing primitive validates every target as a COMMIT, so a working set
// can never be passed to it as if it were a commit. This sibling validates
// commit targets exactly as before AND validates working-set targets as working
// sets, then performs the SAME single expected-root publication over one
// address-map edit, with no optimistic retry.

package datas

import (
	"context"
	"errors"
	"fmt"

	"github.com/nustiueudinastea/dolt/go/store/hash"
	"github.com/nustiueudinastea/dolt/go/store/types"
)

// WorkingSetHeadEdit is the typed working-set counterpart of CommitHeadEdit.
// ExpectedHead is a compare-and-set precondition on the working-set dataset.
type WorkingSetHeadEdit struct {
	DatasetID    string
	ExpectedHead hash.Hash
	TargetHead   hash.Hash
}

// SetHeadsAtRootWithWorkingSets publishes commit-head and working-set-head edits
// in ONE expected-root publication, all-or-none. Commit targets must be commits
// and working-set targets must be working sets; existing heads are type-checked
// the same way. Any validation or expectation failure publishes nothing.
func (db *database) SetHeadsAtRootWithWorkingSets(
	ctx context.Context,
	expectedRoot hash.Hash,
	commitEdits []CommitHeadEdit,
	workingSetEdits []WorkingSetHeadEdit,
) (CommitHeadBatchResult, error) {
	var result CommitHeadBatchResult
	if expectedRoot.IsEmpty() {
		return result, errors.New("expected storage root is empty")
	}
	if len(commitEdits)+len(workingSetEdits) == 0 {
		return result, errors.New("head batch is empty")
	}
	seen := make(map[string]struct{}, len(commitEdits)+len(workingSetEdits))
	requireType := func(datasetID string, target hash.Hash, wantType string) (types.Value, error) {
		if err := ValidateDatasetId(datasetID); err != nil {
			return nil, fmt.Errorf("%w: %s", err, datasetID)
		}
		if _, ok := seen[datasetID]; ok {
			return nil, fmt.Errorf("duplicate dataset edit: %s", datasetID)
		}
		seen[datasetID] = struct{}{}
		if target.IsEmpty() {
			return nil, fmt.Errorf("target head is empty: %s", datasetID)
		}
		value, err := db.ReadValue(ctx, target)
		if err != nil {
			return nil, err
		}
		if value == nil {
			return nil, fmt.Errorf("target head is missing: %s", target)
		}
		head, err := newHead(ctx, value, target)
		if err != nil {
			return nil, err
		}
		if head.TypeName() != wantType {
			return nil, fmt.Errorf("target for %s is %s, not a %s", datasetID, head.TypeName(), wantType)
		}
		return value, nil
	}
	for _, edit := range commitEdits {
		// The value read inside requireType is REUSED here. Re-reading it through
		// a helper that discarded its error meant an I/O failure could reach
		// IsCommit as a nil value instead of stopping the publication.
		value, err := requireType(edit.DatasetID, edit.TargetHead, commitName)
		if err != nil {
			return result, err
		}
		isCommit, err := IsCommit(value)
		if err != nil {
			return result, err
		}
		if !isCommit {
			return result, fmt.Errorf("target for %s is not a commit", edit.DatasetID)
		}
	}
	for _, edit := range workingSetEdits {
		if _, err := requireType(edit.DatasetID, edit.TargetHead, workingSetName); err != nil {
			return result, err
		}
	}

	currentRoot, err := db.rt.Root(ctx)
	if err != nil {
		return result, err
	}
	if currentRoot != expectedRoot {
		return result, ErrOptimisticLockFailed
	}
	addressMap, err := db.loadDatasetsRefmap(ctx, expectedRoot)
	if err != nil {
		return result, err
	}
	checkExpected := func(datasetID string, expected hash.Hash, wantType string) error {
		current, err := addressMap.Get(ctx, datasetID)
		if err != nil {
			return err
		}
		if current != expected {
			return fmt.Errorf("%w: %s", ErrExpectedHeadMismatch, datasetID)
		}
		if current.IsEmpty() {
			return nil
		}
		value, err := db.ReadValue(ctx, current)
		if err != nil {
			return err
		}
		head, err := newHead(ctx, value, current)
		if err != nil {
			return err
		}
		if head.TypeName() != wantType {
			return fmt.Errorf("existing head for %s is %s, not a %s", datasetID, head.TypeName(), wantType)
		}
		return nil
	}
	for _, edit := range commitEdits {
		if err := checkExpected(edit.DatasetID, edit.ExpectedHead, commitName); err != nil {
			return result, err
		}
	}
	for _, edit := range workingSetEdits {
		if err := checkExpected(edit.DatasetID, edit.ExpectedHead, workingSetName); err != nil {
			return result, err
		}
	}

	editor := addressMap.Editor()
	for _, edit := range commitEdits {
		if err := editor.Update(ctx, edit.DatasetID, edit.TargetHead); err != nil {
			return result, err
		}
	}
	for _, edit := range workingSetEdits {
		if err := editor.Update(ctx, edit.DatasetID, edit.TargetHead); err != nil {
			return result, err
		}
	}
	updated, err := editor.Flush(ctx)
	if err != nil {
		return result, err
	}
	rootValue := storeroot_flatbuffer(updated)
	rootRef, err := db.WriteValue(ctx, types.SerialMessage(rootValue))
	if err != nil {
		return result, err
	}
	result.EnteredPublication = true
	if err := db.tryCommitChunks(ctx, rootRef.TargetHash(), expectedRoot); err != nil {
		return result, err
	}
	result.Applied = true
	return result, nil
}

// WriteCleanWorkingSetValue writes a working-set value carrying the given
// working and staged roots, the CALLER'S metadata and NO merge or rebase state,
// returning its address. Callers must already have refused any workspace
// that carried merge or rebase state, so nothing is discarded here.
//
// Metadata is a required argument. An earlier revision invented an identity and
// read time.Now() here, which put a wall-clock second into a content address:
// two calls for identical roots produced different objects whenever they
// straddled a second boundary, so any exact expectation about the published
// object was only accidentally stable. The caller owns that choice now.
func WriteCleanWorkingSetValue(
	ctx context.Context, db Database, workingRoot, stagedRoot hash.Hash, meta *WorkingSetMeta,
) (hash.Hash, error) {
	inner, ok := db.(*database)
	if !ok {
		return hash.Hash{}, errors.New("prototype working-set writer needs the concrete database")
	}
	if meta == nil {
		return hash.Hash{}, errors.New("prototype working-set writer requires explicit metadata")
	}
	staged := stagedRoot
	data := workingset_flatbuffer(workingRoot, &staged, nil, nil, meta)
	written, err := inner.WriteValue(ctx, types.SerialMessage(data))
	if err != nil {
		return hash.Hash{}, err
	}
	return written.TargetHash(), nil
}
