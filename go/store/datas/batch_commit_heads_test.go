package datas

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nustiueudinastea/dolt/go/store/chunks"
	"github.com/nustiueudinastea/dolt/go/store/hash"
	"github.com/nustiueudinastea/dolt/go/store/types"
)

func TestSetCommitHeadsAtRoot(t *testing.T) {
	ctx := context.Background()
	storage := &chunks.TestStorage{}
	db := NewDatabase(storage.NewViewWithDefaultFormat()).(*database)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	_, targetA := addCommit(t, db, "source-a", "a")
	_, targetB := addCommit(t, db, "source-b", "b")
	_, unrelated := addCommit(t, db, "unrelated", "keep")
	rootBefore, err := db.rt.Root(ctx)
	require.NoError(t, err)

	result, err := db.SetCommitHeadsAtRoot(ctx, rootBefore, []CommitHeadEdit{
		{DatasetID: "target-a", TargetHead: targetA},
		{DatasetID: "target-b", TargetHead: targetB},
	})
	require.NoError(t, err)
	require.True(t, result.EnteredPublication)
	require.True(t, result.Applied)
	rootAfter, err := db.rt.Root(ctx)
	require.NoError(t, err)
	require.NotEqual(t, rootBefore, rootAfter)
	for id, want := range map[string]hash.Hash{"target-a": targetA, "target-b": targetB, "unrelated": unrelated} {
		dataset, err := db.GetDataset(ctx, id)
		require.NoError(t, err)
		got, ok := dataset.MaybeHeadAddr()
		require.True(t, ok)
		require.Equal(t, want, got)
	}

	result, err = db.SetCommitHeadsAtRoot(ctx, rootAfter, []CommitHeadEdit{
		{DatasetID: "target-a", ExpectedHead: targetA, TargetHead: targetA},
		{DatasetID: "target-b", ExpectedHead: targetB, TargetHead: targetB},
	})
	require.NoError(t, err)
	require.True(t, result.EnteredPublication)
	require.True(t, result.Applied)
	rootIdempotent, err := db.rt.Root(ctx)
	require.NoError(t, err)
	require.Equal(t, rootAfter, rootIdempotent)

	result, err = db.SetCommitHeadsAtRoot(ctx, rootBefore, []CommitHeadEdit{{DatasetID: "stale", TargetHead: targetA}})
	require.ErrorIs(t, err, ErrOptimisticLockFailed)
	require.False(t, result.EnteredPublication)

	for name, testCase := range map[string]struct {
		edits []CommitHeadEdit
		err   error
	}{
		"empty":             {nil, nil},
		"invalid id":        {[]CommitHeadEdit{{DatasetID: "bad id", TargetHead: targetA}}, ErrInvalidDatasetID},
		"duplicate":         {[]CommitHeadEdit{{DatasetID: "dupe", TargetHead: targetA}, {DatasetID: "dupe", TargetHead: targetB}}, nil},
		"zero target":       {[]CommitHeadEdit{{DatasetID: "zero"}}, nil},
		"expected mismatch": {[]CommitHeadEdit{{DatasetID: "target-a", TargetHead: targetA}}, ErrExpectedHeadMismatch},
	} {
		t.Run(name, func(t *testing.T) {
			before, err := db.rt.Root(ctx)
			require.NoError(t, err)
			result, err := db.SetCommitHeadsAtRoot(ctx, before, testCase.edits)
			require.Error(t, err)
			if testCase.err != nil {
				require.True(t, errors.Is(err, testCase.err), "%v", err)
			}
			require.False(t, result.EnteredPublication)
			after, rootErr := db.rt.Root(ctx)
			require.NoError(t, rootErr)
			require.Equal(t, before, after)
		})
	}

	nonCommitRef, err := db.WriteValue(ctx, types.String("not-a-commit"))
	require.NoError(t, err)
	root, err := db.rt.Root(ctx)
	require.NoError(t, err)
	result, err = db.SetCommitHeadsAtRoot(ctx, root, []CommitHeadEdit{{DatasetID: "non-commit", TargetHead: nonCommitRef.TargetHash()}})
	require.Error(t, err)
	require.False(t, result.EnteredPublication)

	tupleDataset, err := db.GetDataset(ctx, "tuple-head")
	require.NoError(t, err)
	tupleDataset, err = db.SetTuple(ctx, tupleDataset, []byte("tuple"))
	require.NoError(t, err)
	tupleHead, ok := tupleDataset.MaybeHeadAddr()
	require.True(t, ok)
	root, err = db.rt.Root(ctx)
	require.NoError(t, err)
	result, err = db.SetCommitHeadsAtRoot(ctx, root, []CommitHeadEdit{{DatasetID: "tuple-head", ExpectedHead: tupleHead, TargetHead: targetA}})
	require.Error(t, err)
	require.False(t, result.EnteredPublication)
}
