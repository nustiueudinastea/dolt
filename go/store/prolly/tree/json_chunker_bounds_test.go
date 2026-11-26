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

package tree

import (
	"context"
	"strings"
	"testing"

	"github.com/nustiueudinastea/go-mysql-server/sql/types"
	"github.com/stretchr/testify/require"
)

func TestJsonChunker_NoEmptyTrailingLeafOnFinalPositionCut(t *testing.T) {
	ctx := context.Background()
	ns := NewTestNodeStore()

	// Construct a JSON document whose final segment exceeds maxChunkSize (16384),
	// forcing crossesBoundary to fire at the final scanner location (endOfValue).
	doc := types.JSONDocument{Val: map[string]any{
		"a": strings.Repeat("x", 17000),
	}}
	root, err := SerializeJsonToAddr(ctx, ns, doc)
	require.NoError(t, err)
	require.NotNil(t, root)

	// Walk every leaf and assert no leaf is empty, and that full JSON round-trips.
	leafCount := 0
	err = WalkNodes(ctx, root, ns, func(_ context.Context, nd *Node) error {
		if nd.IsLeaf() {
			leafCount++
			require.NotEmpty(t, nd.GetValue(0), "JsonChunker emitted an empty leaf chunk")
		}
		return nil
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, leafCount, 1)

	idxDoc := NewIndexedJsonDocument(root, ns)
	got, err := idxDoc.ToInterface(ctx)
	require.NoError(t, err)
	require.Equal(t, doc.Val, got)
}

func TestJsonChunker_OversizedSeekKeyReturnsTypedError(t *testing.T) {
	ctx := context.Background()
	ns := NewTestNodeStore()

	// Member name > 32,747 bytes exceeds 2*(len(key)+hash.ByteLen) <= MaxVectorOffset
	// when a chunk boundary cuts at that member location.
	longKey := strings.Repeat("k", 33000)
	doc := types.JSONDocument{Val: map[string]any{
		longKey: "v",
		"z":     strings.Repeat("y", 17000),
	}}
	_, err := SerializeJsonToAddr(ctx, ns, doc)
	require.ErrorIs(t, err, ErrJsonLocationTooLarge)

	// Deeply nested path whose accumulated jsonLocation key exceeds capacity.
	seg := strings.Repeat("p", 1000)
	var nested any = map[string]any{"leaf": strings.Repeat("y", 17000), "a": 1}
	for range 34 {
		nested = map[string]any{seg: nested}
	}
	_, err = SerializeJsonToAddr(ctx, ns, types.JSONDocument{Val: nested})
	require.ErrorIs(t, err, ErrJsonLocationTooLarge)
}

func TestJsonChunker_SingleLeafTopLevelString(t *testing.T) {
	ctx := context.Background()
	ns := NewTestNodeStore()

	doc := types.JSONDocument{Val: strings.Repeat("s", 5000)}
	root, err := SerializeJsonToAddr(ctx, ns, doc)
	require.NoError(t, err)
	require.True(t, root.IsLeaf())
	require.Equal(t, 5002, len(root.GetValue(0)))

	idxDoc := NewIndexedJsonDocument(root, ns)
	got, err := idxDoc.ToInterface(ctx)
	require.NoError(t, err)
	require.Equal(t, doc.Val, got)
}

func TestJsonChunker_OversizedSingleLeafReturnsTypedError(t *testing.T) {
	ctx := context.Background()
	ns := NewTestNodeStore()

	// A top-level JSON string near 16 MiB forms a single leaf whose Blob
	// FlatBuffer would exceed serial.FinishMessage's 24-bit size field.
	doc := types.JSONDocument{Val: strings.Repeat("a", (1<<24)-2)}
	_, err := SerializeJsonToAddr(ctx, ns, doc)
	require.ErrorIs(t, err, ErrJsonLeafTooLarge)

	// A top-level JSON string at exact maxJsonLeafPayloadBytes succeeds without panicking.
	okDoc := types.JSONDocument{Val: strings.Repeat("b", maxJsonLeafPayloadBytes-2)}
	root, err := SerializeJsonToAddr(ctx, ns, okDoc)
	require.NoError(t, err)
	require.True(t, root.IsLeaf())
	require.Equal(t, maxJsonLeafPayloadBytes, len(root.GetValue(0)))
}

func TestKeySplitter_HighLevelNoPanic(t *testing.T) {
	s := newKeySplitter(20)
	require.NotNil(t, s)
	ks, ok := s.(*keySplitter)
	require.True(t, ok)
	require.Equal(t, saltFromLevel(21), ks.salt)
}

func TestPutBlobBuilder_DetachesNodeStore(t *testing.T) {
	ns := NewTestNodeStore()
	bb := ns.BlobBuilder()
	require.NotNil(t, bb.ns)
	ns.PutBlobBuilder(bb)
	require.Nil(t, bb.ns)
	require.Nil(t, bb.S)
}
