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

package val

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nustiueudinastea/dolt/go/store/pool"
)

func TestAdaptiveValue_OutOfBandSizeMatchesActualEncoding(t *testing.T) {
	ctx := context.Background()
	vs := &mockValueStore{}
	for _, n := range []int{1, 20, 240, 241, 1500, 2287, 2288, 10000} {
		inline := AdaptiveValue(AdaptiveValueInlineBytes(bytes.Repeat([]byte("a"), n)))
		oob, err := inline.convertToOutOfBand(ctx, vs, nil)
		require.NoError(t, err)
		require.Equal(t, int64(len(oob)), inline.outOfBandSize(), "payload length %d", n)
	}
}

func TestTupleBuilder_BuildPermissiveSpillRespectsTargetWithMultiByteVarint(t *testing.T) {
	ctx := context.Background()
	vs := &mockValueStore{}
	bp := pool.NewBuffPool()
	desc := NewTupleDescriptor(
		Type{Enc: StringAdaptiveEnc, Nullable: false},
		Type{Enc: StringAdaptiveEnc, Nullable: false},
		Type{Enc: StringAdaptiveEnc, Nullable: false},
	)
	tb := NewTupleBuilder(desc, vs)
	require.NoError(t, tb.PutAdaptiveStringFromInline(ctx, 0, string(bytes.Repeat([]byte("a"), 1500))))
	require.NoError(t, tb.PutAdaptiveStringFromInline(ctx, 1, string(bytes.Repeat([]byte("b"), 1020))))
	require.NoError(t, tb.PutAdaptiveStringFromInline(ctx, 2, string(bytes.Repeat([]byte("c"), 1005))))

	tup, err := tb.BuildPermissive(ctx, bp)
	require.NoError(t, err)

	// Tuple framing for 3 non-NULL fields is 2 offsets (4 bytes) + count (2 bytes) = 6 bytes.
	// Data bytes must not exceed DefaultTupleLengthTarget (2048).
	dataBytes := len(tup) - (desc.Count()-1)*2 - 2
	require.LessOrEqual(t, dataBytes, int(DefaultTupleLengthTarget))
}
