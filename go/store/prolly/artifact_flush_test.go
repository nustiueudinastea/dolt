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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nustiueudinastea/dolt/go/gen/fb/serial"
	"github.com/nustiueudinastea/dolt/go/store/chunks"
	"github.com/nustiueudinastea/dolt/go/store/hash"
	"github.com/nustiueudinastea/dolt/go/store/prolly/tree"
	"github.com/nustiueudinastea/dolt/go/store/types"
	"github.com/nustiueudinastea/dolt/go/store/val"
)

// TestArtifactsEditorAutoFlushPreservesArtifactFileID verifies that when
// ArtifactsEditor automatically flushes pending edits, it serializes nodes
// with MergeArtifactsFileID (ARTM) rather than ProllyTreeNodeFileID (TUPM).
func TestArtifactsEditorAutoFlushPreservesArtifactFileID(t *testing.T) {
	ctx := context.Background()
	ts := &chunks.TestStorage{}
	cs := ts.NewViewWithFormat(types.Format_DOLT.VersionString())
	ns := tree.NewNodeStore(cs)
	srcKD := val.NewTupleDescriptor(val.Type{Enc: val.Int32Enc, Nullable: false})

	am, err := NewArtifactMapFromTuples(ctx, ns, srcKD)
	require.NoError(t, err)

	ed := am.EditorWithMaxPending(1)
	kb := val.NewTupleBuilder(srcKD, ns)
	for _, id := range []int32{1, 2, 3} {
		kb.PutInt32(0, id)
		k, err := kb.Build(ctx, ns.Pool())
		require.NoError(t, err)
		require.NoError(t, ed.Add(ctx, k, hash.Of([]byte{byte(id)}), ArtifactTypeConflict, []byte(`{"base":"x"}`), nil))
	}

	flushed, err := ed.Flush(ctx)
	require.NoError(t, err)
	cnt, err := flushed.Count()
	require.NoError(t, err)
	require.Equal(t, 3, cnt)

	c, err := cs.Get(ctx, flushed.HashOf())
	require.NoError(t, err)
	require.Equal(t, serial.MergeArtifactsFileID, serial.GetFileID(c.Data()))
}
