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

package table

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

type contextProbeIter struct{ next, close *sql.Context }

func (p *contextProbeIter) Next(ctx *sql.Context) (sql.Row, error) { p.next = ctx; return nil, io.EOF }
func (p *contextProbeIter) Close(ctx *sql.Context) error           { p.close = ctx; return nil }

func TestRowIteratorPreservesSQLContext(t *testing.T) {
	ctx := sql.NewEmptyContext()
	probe := &contextProbeIter{}
	iter := NewRowIter(probe)
	_, err := iter.Next(ctx)
	require.ErrorIs(t, err, io.EOF)
	require.NoError(t, iter.Close(ctx))
	require.Same(t, ctx, probe.next)
	require.Same(t, ctx, probe.close)
}
