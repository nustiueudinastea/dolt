// Unit coverage for SetHeadsAtRootWithWorkingSets and
// WriteCleanWorkingSetValue.
//
// It mirrors the base primitive's own batch_commit_heads_test.go: type
// validation, invalid and duplicate dataset ids, expected-head CAS,
// expected-root CAS, all-or-none and success. There is no entered-error case
// here: producing one would need an injectable hook inside the publication, and
// this file adds no hook and no mutable package global.

package datas

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nustiueudinastea/dolt/go/store/chunks"
	"github.com/nustiueudinastea/dolt/go/store/hash"
)

// protoTestMeta is fixed, so every address this file asserts about is a
// function of its inputs alone.
func protoTestMeta() *WorkingSetMeta {
	return &WorkingSetMeta{
		Name:        "prototype test",
		Email:       "prototype@swarmion.invalid",
		Timestamp:   1700000000,
		Description: "typed sibling unit coverage",
	}
}

// writeProtoTestWorkingSet writes a clean working-set value and returns its
// address, so a test can supply a genuinely typed working-set target.
func writeProtoTestWorkingSet(t *testing.T, db *database, working, staged hash.Hash) hash.Hash {
	t.Helper()
	addr, err := WriteCleanWorkingSetValue(
		context.Background(), db, working, staged, protoTestMeta())
	require.NoError(t, err)
	return addr
}

// TestWriteCleanWorkingSetValueMetadataIsCallerSupplied pins the property the
// owner and its fixtures depend on: the published object is a function of the
// roots AND the caller's metadata, and of nothing else. Identical inputs give
// identical addresses however much wall-clock time passes between calls, and a
// changed timestamp gives a different address.
func TestWriteCleanWorkingSetValueMetadataIsCallerSupplied(t *testing.T) {
	ctx := context.Background()
	storage := &chunks.TestStorage{}
	db := NewDatabase(storage.NewViewWithDefaultFormat()).(*database)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	commitValue, _ := addCommit(t, db, "meta-source", "a")
	rootValue, err := db.WriteValue(ctx, commitValue)
	require.NoError(t, err)
	root := rootValue.TargetHash()

	first, err := WriteCleanWorkingSetValue(ctx, db, root, root, protoTestMeta())
	require.NoError(t, err)
	second, err := WriteCleanWorkingSetValue(ctx, db, root, root, protoTestMeta())
	require.NoError(t, err)
	require.Equal(t, first, second, "identical roots and metadata must give one address")

	later := protoTestMeta()
	later.Timestamp++
	shifted, err := WriteCleanWorkingSetValue(ctx, db, root, root, later)
	require.NoError(t, err)
	require.NotEqual(t, first, shifted,
		"a different timestamp must give a different address; this is exactly why the writer "+
			"must not read the clock itself")

	_, err = WriteCleanWorkingSetValue(ctx, db, root, root, nil)
	require.Error(t, err, "metadata is required")
}

func TestSetHeadsAtRootWithWorkingSets(t *testing.T) {
	ctx := context.Background()
	storage := &chunks.TestStorage{}
	db := NewDatabase(storage.NewViewWithDefaultFormat()).(*database)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	commitValue, commitAddr := addCommit(t, db, "proto-source", "a")
	_, otherCommit := addCommit(t, db, "proto-other", "b")
	_, untouched := addCommit(t, db, "proto-untouched", "keep")
	require.NotNil(t, commitValue)

	rootValue, err := db.WriteValue(ctx, commitValue)
	require.NoError(t, err)
	workingRoot := rootValue.TargetHash()
	workingSet := writeProtoTestWorkingSet(t, db, workingRoot, workingRoot)
	otherWorkingSet := writeProtoTestWorkingSet(t, db, workingRoot, otherCommit)

	rootBefore, err := db.rt.Root(ctx)
	require.NoError(t, err)

	t.Run("expected storage root CAS", func(t *testing.T) {
		stale := hash.Parse("48louhk7jn8afqo9ug3uk9mdv0hu8kud")
		result, err := db.SetHeadsAtRootWithWorkingSets(ctx, stale,
			[]CommitHeadEdit{{DatasetID: "proto-k", TargetHead: commitAddr}},
			[]WorkingSetHeadEdit{{DatasetID: "workingSets/heads/proto", TargetHead: workingSet}})
		require.ErrorIs(t, err, ErrOptimisticLockFailed)
		require.False(t, result.EnteredPublication)
		require.False(t, result.Applied)
		current, rootErr := db.rt.Root(ctx)
		require.NoError(t, rootErr)
		require.Equal(t, rootBefore, current)
	})

	t.Run("empty expected root", func(t *testing.T) {
		_, err := db.SetHeadsAtRootWithWorkingSets(ctx, hash.Hash{},
			[]CommitHeadEdit{{DatasetID: "proto-k", TargetHead: commitAddr}}, nil)
		require.Error(t, err)
	})

	t.Run("empty batch", func(t *testing.T) {
		_, err := db.SetHeadsAtRootWithWorkingSets(ctx, rootBefore, nil, nil)
		require.Error(t, err)
	})

	t.Run("type validation", func(t *testing.T) {
		// A working set may not be published as a commit.
		_, err := db.SetHeadsAtRootWithWorkingSets(ctx, rootBefore,
			[]CommitHeadEdit{{DatasetID: "proto-k", TargetHead: workingSet}}, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not a "+commitName)

		// A commit may not be published as a working set.
		_, err = db.SetHeadsAtRootWithWorkingSets(ctx, rootBefore, nil,
			[]WorkingSetHeadEdit{{DatasetID: "workingSets/heads/proto", TargetHead: commitAddr}})
		require.Error(t, err)
		require.Contains(t, err.Error(), "not a "+workingSetName)
	})

	t.Run("invalid and duplicate dataset ids", func(t *testing.T) {
		_, err := db.SetHeadsAtRootWithWorkingSets(ctx, rootBefore,
			[]CommitHeadEdit{{DatasetID: "bad id", TargetHead: commitAddr}}, nil)
		require.ErrorIs(t, err, ErrInvalidDatasetID)

		_, err = db.SetHeadsAtRootWithWorkingSets(ctx, rootBefore,
			[]CommitHeadEdit{
				{DatasetID: "proto-dupe", TargetHead: commitAddr},
				{DatasetID: "proto-dupe", TargetHead: otherCommit},
			}, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate dataset edit")

		// A commit edit and a working-set edit may not name the same dataset.
		_, err = db.SetHeadsAtRootWithWorkingSets(ctx, rootBefore,
			[]CommitHeadEdit{{DatasetID: "proto-shared", TargetHead: commitAddr}},
			[]WorkingSetHeadEdit{{DatasetID: "proto-shared", TargetHead: workingSet}})
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate dataset edit")
	})

	t.Run("empty target head", func(t *testing.T) {
		_, err := db.SetHeadsAtRootWithWorkingSets(ctx, rootBefore, nil,
			[]WorkingSetHeadEdit{{DatasetID: "workingSets/heads/proto"}})
		require.Error(t, err)
		require.Contains(t, err.Error(), "target head is empty")
	})

	t.Run("all or none on a failing edit", func(t *testing.T) {
		// One good commit edit and one working-set edit whose expected head is
		// wrong: nothing at all may be published.
		result, err := db.SetHeadsAtRootWithWorkingSets(ctx, rootBefore,
			[]CommitHeadEdit{{DatasetID: "proto-k", TargetHead: commitAddr}},
			[]WorkingSetHeadEdit{{
				DatasetID: "workingSets/heads/proto", ExpectedHead: otherWorkingSet,
				TargetHead: workingSet,
			}})
		require.ErrorIs(t, err, ErrExpectedHeadMismatch)
		require.False(t, result.EnteredPublication)
		require.False(t, result.Applied)
		dataset, dsErr := db.GetDataset(ctx, "proto-k")
		require.NoError(t, dsErr)
		_, present := dataset.MaybeHeadAddr()
		require.False(t, present, "a refused batch published one of its edits")
		current, rootErr := db.rt.Root(ctx)
		require.NoError(t, rootErr)
		require.Equal(t, rootBefore, current)
	})

	t.Run("success publishes commits and working sets together", func(t *testing.T) {
		result, err := db.SetHeadsAtRootWithWorkingSets(ctx, rootBefore,
			[]CommitHeadEdit{{DatasetID: "proto-k", TargetHead: commitAddr}},
			[]WorkingSetHeadEdit{{DatasetID: "workingSets/heads/proto", TargetHead: workingSet}})
		require.NoError(t, err)
		require.True(t, result.EnteredPublication)
		require.True(t, result.Applied)

		rootAfter, rootErr := db.rt.Root(ctx)
		require.NoError(t, rootErr)
		require.NotEqual(t, rootBefore, rootAfter)

		for id, want := range map[string]hash.Hash{
			"proto-k":                 commitAddr,
			"workingSets/heads/proto": workingSet,
			"proto-untouched":         untouched,
		} {
			dataset, dsErr := db.GetDataset(ctx, id)
			require.NoError(t, dsErr)
			got, present := dataset.MaybeHeadAddr()
			require.True(t, present, id)
			require.Equal(t, want, got, id)
		}

		// Expected-head CAS on the now-established datasets.
		_, err = db.SetHeadsAtRootWithWorkingSets(ctx, rootAfter,
			[]CommitHeadEdit{{DatasetID: "proto-k", TargetHead: otherCommit}}, nil)
		require.ErrorIs(t, err, ErrExpectedHeadMismatch)

		result, err = db.SetHeadsAtRootWithWorkingSets(ctx, rootAfter,
			[]CommitHeadEdit{{DatasetID: "proto-k", ExpectedHead: commitAddr, TargetHead: otherCommit}},
			[]WorkingSetHeadEdit{{
				DatasetID: "workingSets/heads/proto", ExpectedHead: workingSet,
				TargetHead: otherWorkingSet,
			}})
		require.NoError(t, err)
		require.True(t, result.Applied)

		// An existing head keeps its type: a commit dataset cannot become a
		// working set.
		rootNow, rootErr := db.rt.Root(ctx)
		require.NoError(t, rootErr)
		_, err = db.SetHeadsAtRootWithWorkingSets(ctx, rootNow, nil,
			[]WorkingSetHeadEdit{{
				DatasetID: "proto-k", ExpectedHead: otherCommit, TargetHead: workingSet,
			}})
		require.Error(t, err)
		require.Contains(t, err.Error(), "existing head for proto-k")
	})
}
