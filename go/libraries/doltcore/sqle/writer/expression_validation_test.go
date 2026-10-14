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

package writer

import (
	"errors"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/sqle/expranalysis"
)

func TestWriterValidatesCachedExpressions(t *testing.T) {
	ctx := sql.NewEmptyContext()
	rejected := errors.New("cached expression is forbidden")
	rejecting := expranalysis.WithExpressionValidator(ctx, func(*sql.Context, sql.Expression) error { return rejected })
	evals := 0
	virtual := &countedWriterExpression{Expression: expression.NewLiteral(int64(7), types.Int64), evals: &evals}
	predicate := &countedWriterExpression{Expression: expression.NewLiteral(false, types.Boolean), evals: &evals}
	for _, keyless := range []bool{false, true} {
		name := "keyed"
		if keyless {
			name = "keyless"
		}
		t.Run(name, func(t *testing.T) {
			evals = 0
			keyed := prollySecondaryIndexWriter{virtualExprs: []sql.Expression{virtual}, predicate: predicate}
			unkeyed := prollyKeylessSecondaryWriter{virtualExprs: []sql.Expression{virtual}, predicate: predicate}
			keyPart := keyed.keyPartFromRow
			if keyless {
				keyPart = unkeyed.keyPartFromRow
			}
			_, err := keyPart(rejecting, 0, nil)
			assert.ErrorIs(t, err, rejected)
			require.Zero(t, evals)
			value, err := keyPart(ctx, 0, nil)
			require.NoError(t, err)
			require.EqualValues(t, 7, value)
			require.Equal(t, 1, evals)
			evals = 0
			if keyless {
				assert.ErrorIs(t, unkeyed.validateUniqueKeyViolation(rejecting, nil), rejected)
				require.Zero(t, evals)
				require.NoError(t, unkeyed.validateUniqueKeyViolation(ctx, nil))
			} else {
				_, err = keyed.matchesPredicate(rejecting, nil)
				assert.ErrorIs(t, err, rejected)
				require.Zero(t, evals)
				matches, err := keyed.matchesPredicate(ctx, nil)
				require.NoError(t, err)
				require.False(t, matches)
			}
			require.Equal(t, 1, evals)
		})
	}
}

type countedWriterExpression struct {
	sql.Expression
	evals *int
}

func (e *countedWriterExpression) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	*e.evals++
	return e.Expression.Eval(ctx, row)
}
