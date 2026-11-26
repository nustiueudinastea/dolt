// Copyright 2026 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prolly

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nustiueudinastea/dolt/go/store/prolly/tree"
	"github.com/nustiueudinastea/dolt/go/store/val"
)

// TestMutableMapPrefixRangeWithPendingTombstone verifies that IterRange over a
// PrefixRange merges pending deletes with persisted entries using the map's
// full key descriptor, so deleting a later sibling key with the same prefix
// does not hide an earlier persisted sibling key.
func TestMutableMapPrefixRangeWithPendingTombstone(t *testing.T) {
	ctx := context.Background()
	ns := tree.NewTestNodeStore()

	kd := val.NewTupleDescriptor(
		val.Type{Enc: val.Int32Enc, Nullable: true},
		val.Type{Enc: val.Int32Enc, Nullable: false},
	)
	vd := val.NewTupleDescriptor()
	prefixDesc := kd.PrefixDesc(1)

	kb := val.NewTupleBuilder(kd, ns)
	buildKey := func(code, id int32) val.Tuple {
		kb.PutInt32(0, code)
		kb.PutInt32(1, id)
		tup, err := kb.Build(ctx, ns.Pool())
		require.NoError(t, err)
		return tup
	}

	m, err := NewMapFromTuples(ctx, ns, kd, vd)
	require.NoError(t, err)

	mut := m.Mutate().WithMaxPending(1)
	require.NoError(t, mut.Put(ctx, buildKey(100, 5), val.EmptyTuple))
	require.NoError(t, mut.Put(ctx, buildKey(100, 10), val.EmptyTuple))
	// Trigger a flush so (100, 5) and (100, 10) live in the persisted map.
	require.NoError(t, mut.Put(ctx, buildKey(200, 6), val.EmptyTuple))
	require.NoError(t, mut.Put(ctx, buildKey(200, 7), val.EmptyTuple))

	// Pending delete of the later sibling (100, 10) must leave (100, 5) visible.
	require.NoError(t, mut.Delete(ctx, buildKey(100, 10)))

	rng, err := PrefixRange(ctx, buildKey(100, 10), prefixDesc)
	require.NoError(t, err)
	iter, err := mut.IterRange(ctx, rng)
	require.NoError(t, err)

	var seenIDs []int32
	for {
		k, _, err := iter.Next(ctx)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		id, ok := kd.GetInt32(1, k)
		require.True(t, ok)
		seenIDs = append(seenIDs, id)
	}
	require.Equal(t, []int32{5}, seenIDs)
}
