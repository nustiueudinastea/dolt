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

package merge

import (
	"errors"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	sqltypes "github.com/dolthub/go-mysql-server/sql/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/schema"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/sqle/expranalysis"
	"github.com/nustiueudinastea/dolt/go/store/prolly/tree"
	"github.com/nustiueudinastea/dolt/go/store/types"
	"github.com/nustiueudinastea/dolt/go/store/val"
)

// Cached expressions must use the evaluating context's validator, even if they
// were resolved earlier without one.
func TestMergeValidatesCachedExpressions(t *testing.T) {
	ctx := sql.NewEmptyContext()
	ns := tree.NewTestNodeStore()
	id := schema.NewColumn("id", 0, types.IntKind, true)
	col := schema.NewColumn("d", 1, types.IntKind, false)
	sch, err := schema.SchemaFromCols(schema.NewColCollection(id, col))
	require.NoError(t, err)
	tm := &TableMerger{ns: ns, leftSch: sch}
	keyBuilder := val.NewTupleBuilder(sch.GetKeyDescriptor(ns), ns)
	keyBuilder.PutInt64(0, 1)
	key, err := keyBuilder.Build(ctx, ns.Pool())
	require.NoError(t, err)
	tb := val.NewTupleBuilder(sch.GetValueDescriptor(ns), ns)
	tb.PutInt64(0, 1)
	value, err := tb.Build(ctx, ns.Pool())
	require.NoError(t, err)
	evals := 0
	expr := &countedMergeExpression{Expression: expression.NewLiteral(int64(7), sqltypes.Int64), evals: &evals}
	rejected := errors.New("cached expression is forbidden")
	rejecting := expranalysis.WithExpressionValidator(ctx, func(_ *sql.Context, e sql.Expression) error {
		require.Same(t, expr, e)
		return rejected
	})
	t.Run("default", func(t *testing.T) {
		evals = 0
		err := writeTupleExpression(rejecting, key, value, expr, col, sch, tm, tb, 0)
		assert.ErrorIs(t, err, rejected)
		require.Zero(t, evals)
		require.NoError(t, writeTupleExpression(ctx, key, value, expr, col, sch, tm, tb, 0))
		require.Equal(t, 1, evals)
		evals = 0
	})
	t.Run("check", func(t *testing.T) {
		evals = 0
		cv := checkValidator{
			checkExpressions: []checkExpression{{name: "cached", expr: expr}},
			tableMerger:      tm, valueMerger: &valueMerger{ns: ns}, sch: sch, leftMatchesFinal: true,
		}
		diff := tree.ThreeWayDiff{Op: tree.DiffOpLeftAdd, Key: key, Left: value}
		_, err := cv.validateDiff(rejecting, diff)
		assert.ErrorIs(t, err, rejected)
		require.Zero(t, evals)
		_, err = cv.validateDiff(ctx, diff)
		require.NoError(t, err)
		require.Equal(t, 1, evals)
	})
}

type countedMergeExpression struct {
	sql.Expression
	evals *int
}

func (e *countedMergeExpression) Eval(*sql.Context, sql.Row) (interface{}, error) {
	*e.evals++
	return int64(7), nil
}
