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

package merge_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/transform"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/doltdb"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/dtestutils"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/merge"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/schema"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/sqle/expranalysis"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/table/editor"
)

func TestMergeExpressionValidator(t *testing.T) {
	for _, check := range []bool{false, true} {
		name := "fill default"
		if check {
			name = "validate check"
		}
		t.Run(name, func(t *testing.T) {
			var evals atomic.Int32
			ctx := mergeMarkerContext(&evals)
			denv := dtestutils.CreateTestEnv()
			ddb := denv.DoltDB(ctx)
			base := sch("CREATE TABLE t (id BIGINT PRIMARY KEY)")
			final := base
			if check {
				_, err := final.sch.Checks().AddCheck("marker_check", "(merge_marker() > id)", true, false)
				require.NoError(t, err)
			} else {
				final = sch("CREATE TABLE t (id BIGINT PRIMARY KEY, d BIGINT DEFAULT (7))")
				cols := final.sch.GetAllCols().GetColumns()
				cols[1].Default = "(merge_marker())"
				var err error
				final.sch, err = schema.SchemaFromCols(schema.NewColCollection(cols...))
				require.NoError(t, err)
			}
			var eo editor.Options
			a := makeRootWithTable(t, ddb, eo, *tbl(base, row(0)))
			var l, r, expected doltdb.RootValue
			if check {
				l = makeRootWithTable(t, ddb, eo, *tbl(final, row(0), row(1)))
				r = makeRootWithTable(t, ddb, eo, *tbl(final, row(0), row(2)))
				expected = makeRootWithTable(t, ddb, eo, *tbl(final, row(0), row(1), row(2)))
			} else {
				l = makeRootWithTable(t, ddb, eo, *tbl(final, row(0, 7), row(1, 7)))
				r = makeRootWithTable(t, ddb, eo, *tbl(base, row(0), row(2)))
				expected = makeRootWithTable(t, ddb, eo, *tbl(final, row(0, 7), row(1, 7), row(2, 7)))
			}
			rejected := errors.New("merge marker is forbidden")
			rejecting := expranalysis.WithExpressionValidator(ctx, func(_ *sql.Context, e sql.Expression) error {
				var found bool
				transform.InspectExpr(ctx, e, func(_ *sql.Context, e sql.Expression) bool {
					if _, ok := e.(*mergeMarker); ok {
						found = true
					}
					return true
				})
				if found {
					return rejected
				}
				return nil
			})
			_, err := merge.MergeRoots(rejecting, doltdb.SimpleTableResolver{}, l, r, a, rootish{r}, rootish{a}, eo, merge.MergeOpts{})
			assert.ErrorIs(t, err, rejected)
			require.Zero(t, evals.Load())
			// An unvalidated merge still evaluates the marker and yields the same rows.
			result, err := merge.MergeRoots(ctx, doltdb.SimpleTableResolver{}, l, r, a, rootish{r}, rootish{a}, eo, merge.MergeOpts{})
			require.NoError(t, err)
			require.Positive(t, evals.Load())
			verifyMerge(t, ctx, expected, result, a.NodeStore(), false, nil)
		})
	}
}

type mergeMarker struct {
	sql.Expression
	evals *atomic.Int32
}

func (e *mergeMarker) String() string       { return "merge_marker()" }
func (e *mergeMarker) FunctionName() string { return "merge_marker" }
func (e *mergeMarker) Description() string  { return "counts evaluations" }
func (e *mergeMarker) WithChildren(_ *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(e, len(children), 0)
	}
	return e, nil
}
func (e *mergeMarker) Eval(*sql.Context, sql.Row) (interface{}, error) {
	e.evals.Add(1)
	return int64(7), nil
}

type mergeMarkerProvider struct {
	*memory.DbProvider
	evals *atomic.Int32
}

func (p *mergeMarkerProvider) Function(_ *sql.Context, _, name string) (sql.Function, bool) {
	if name != "merge_marker" {
		return nil, false
	}
	return sql.Function0{Name: name, Fn: func(*sql.Context) sql.Expression {
		return &mergeMarker{Expression: expression.NewLiteral(int64(7), types.Int64), evals: p.evals}
	}}, true
}

type mergeMarkerSession struct {
	sql.Session
	provider *mergeMarkerProvider
}

func (s *mergeMarkerSession) GenericProvider() sql.MutableDatabaseProvider { return s.provider }
func mergeMarkerContext(evals *atomic.Int32) *sql.Context {
	ctx := sql.NewEmptyContext()
	provider := &mergeMarkerProvider{DbProvider: memory.NewDBProvider(memory.NewDatabase("mydb")), evals: evals}
	ctx = sql.NewContext(context.Background(), sql.WithSession(&mergeMarkerSession{Session: ctx.Session, provider: provider}))
	ctx.SetCurrentDatabase("mydb")
	return ctx
}
