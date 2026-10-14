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

package expranalysis_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	gms "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/transform"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/stretchr/testify/require"

	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/doltdb"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/dtestutils"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/schema"
	_ "github.com/nustiueudinastea/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/sqle/expranalysis"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/sqle/sqlutil"
	storetypes "github.com/nustiueudinastea/dolt/go/store/types"
)

func TestResolveSchemaExpressionValidator(t *testing.T) {
	var evals atomic.Int32
	ctx := markerContext(&evals)
	id := schema.NewColumn("id", 0, storetypes.IntKind, true)
	def := schema.NewColumn("d", 1, storetypes.IntKind, false)
	def.Default = "(marker_default() + 1)"
	gen := schema.NewColumn("g", 2, storetypes.IntKind, false)
	gen.Generated = "(marker_generated() + id)"
	virtual := schema.NewColumn("v", 3, storetypes.IntKind, false)
	virtual.Generated, virtual.Virtual = "(marker_virtual() + id)", true
	update := schema.NewColumn("u", 4, storetypes.TimestampKind, false)
	// ON UPDATE only accepts time-function syntax; the provider supplies a counting implementation.
	update.OnUpdate = "NOW()"
	sch, err := schema.SchemaFromCols(schema.NewColCollection(id, def, gen, virtual, update))
	require.NoError(t, err)
	_, err = sch.Checks().AddCheck("c1", "(marker_check() > id)", true, false)
	require.NoError(t, err)
	_, err = sch.Checks().AddCheck("c2", "(marker_unenforced() > id)", false, false)
	require.NoError(t, err)

	baseline, err := expranalysis.ResolveSchema(ctx, "t", sch)
	require.NoError(t, err)
	seen := make(map[string]int)
	validating := expranalysis.WithExpressionValidator(ctx, func(got *sql.Context, e sql.Expression) error {
		require.True(t, e.Resolved(), "%s", e)
		require.NotNil(t, expranalysis.ExpressionValidatorFromContext(got))
		for _, name := range markerNames(e) {
			seen[name]++
		}
		return nil
	})
	resolved, err := expranalysis.ResolveSchema(validating, "t", sch)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprint(baseline), fmt.Sprint(resolved))
	require.Equal(t, map[string]int{
		"marker_default": 1, "marker_generated": 1, "marker_virtual": 1,
		"marker_update": 1, "marker_check": 1, "marker_unenforced": 1,
	}, seen)
	require.Zero(t, evals.Load())
	require.Nil(t, expranalysis.ExpressionValidatorFromContext(ctx))
	require.Nil(t, expranalysis.ExpressionValidatorFromContext(expranalysis.WithExpressionValidator(validating, nil)))

	for name := range seen {
		t.Run(name, func(t *testing.T) {
			rejected := errors.New("rejected " + name)
			rejecting := expranalysis.WithExpressionValidator(ctx, func(_ *sql.Context, e sql.Expression) error {
				for _, found := range markerNames(e) {
					if found == name {
						return rejected
					}
				}
				return nil
			})
			_, err := expranalysis.ResolveSchema(rejecting, "t", sch)
			require.ErrorIs(t, err, rejected)
			// These entry points also resolve the rest of the schema, not just the requested expression.
			_, err = expranalysis.ResolveDefaultExpression(rejecting, "t", sch, def)
			require.ErrorIs(t, err, rejected)
			_, err = expranalysis.ResolveCheckExpression(rejecting, "t", sch, "(marker_check() > id)")
			require.ErrorIs(t, err, rejected)
			require.Zero(t, evals.Load())
		})
	}
	// No validator retains the original resolved expressions and evaluation results.
	defExpr, err := expranalysis.ResolveDefaultExpression(ctx, "t", sch, def)
	require.NoError(t, err)
	value, err := defExpr.Eval(ctx, nil)
	require.NoError(t, err)
	require.EqualValues(t, 8, value)
	check, err := expranalysis.ResolveCheckExpression(ctx, "t", sch, "(marker_check() > id)")
	require.NoError(t, err)
	value, err = check.Eval(ctx, sql.Row{int64(0)})
	require.NoError(t, err)
	require.Equal(t, true, value)
	require.EqualValues(t, 2, evals.Load())
}

func TestResolveExpressionValidator(t *testing.T) {
	var evals atomic.Int32
	ctx := markerContext(&evals)
	rejected := errors.New("marker is forbidden")
	rejecting := expranalysis.WithExpressionValidator(ctx, func(_ *sql.Context, e sql.Expression) error {
		if len(markerNames(e)) > 0 {
			return rejected
		}
		return nil
	})
	_, err := expranalysis.ResolveExpression(rejecting, doltdb.TableName{Name: "t"}, "marker_predicate() + 1")
	require.ErrorIs(t, err, rejected)
	require.Zero(t, evals.Load())
	baseline, baselineErr := expranalysis.ResolveExpression(ctx, doltdb.TableName{Name: "t"}, "marker_predicate() + 1")
	allowing := expranalysis.WithExpressionValidator(ctx, func(*sql.Context, sql.Expression) error { return nil })
	allowed, allowedErr := expranalysis.ResolveExpression(allowing, doltdb.TableName{Name: "t"}, "marker_predicate() + 1")
	if baselineErr != nil {
		// Preserve existing analyzer diagnostics as well as successful resolutions.
		require.EqualError(t, allowedErr, baselineErr.Error())
	} else {
		require.NoError(t, allowedErr)
		require.Equal(t, baseline.String(), allowed.String())
	}
}

// Schema-file imports planbuild CREATE TABLE directly rather than using ResolveSchema.
func TestParseCreateTableExpressionValidator(t *testing.T) {
	var evals atomic.Int32
	ctx := markerContext(&evals)
	engine := gms.NewDefault(ctx.Session.(*markerSession).provider)
	defer engine.Close()
	ddb := dtestutils.CreateTestEnv().DoltDB(ctx)
	root, err := doltdb.EmptyRootValue(ctx, ddb.ValueReadWriter(), ddb.NodeStore())
	require.NoError(t, err)
	query := `CREATE TABLE imported (
		id BIGINT PRIMARY KEY,
		d BIGINT DEFAULT (marker_default()),
		g BIGINT AS (marker_generated() + id) STORED,
		u TIMESTAMP ON UPDATE NOW(),
		CONSTRAINT c CHECK (marker_check() > id))`
	_, baseline, err := sqlutil.ParseCreateTableStatement(ctx, root, engine, query)
	require.NoError(t, err)
	seen := make(map[string]int)
	validating := expranalysis.WithExpressionValidator(ctx, func(_ *sql.Context, e sql.Expression) error {
		for _, name := range markerNames(e) {
			seen[name]++
		}
		return nil
	})
	_, allowed, err := sqlutil.ParseCreateTableStatement(validating, root, engine, query)
	require.NoError(t, err)
	require.True(t, schema.SchemasAreEqual(baseline, allowed))
	require.Equal(t, map[string]int{"marker_default": 1, "marker_generated": 1, "marker_update": 1, "marker_check": 1}, seen)
	for name := range seen {
		t.Run(name, func(t *testing.T) {
			rejected := errors.New("import rejects " + name)
			rejecting := expranalysis.WithExpressionValidator(ctx, func(_ *sql.Context, e sql.Expression) error {
				for _, found := range markerNames(e) {
					if found == name {
						return rejected
					}
				}
				return nil
			})
			_, _, err := sqlutil.ParseCreateTableStatement(rejecting, root, engine, query)
			require.ErrorIs(t, err, rejected)
			require.Zero(t, evals.Load())
		})
	}
}

func markerNames(e sql.Expression) (names []string) {
	transform.InspectExpr(nil, e, func(_ *sql.Context, e sql.Expression) bool {
		if marker, ok := e.(*markerExpression); ok {
			names = append(names, marker.name)
		}
		return true
	})
	return
}

type markerExpression struct {
	sql.Expression
	name  string
	evals *atomic.Int32
}

func (e *markerExpression) String() string       { return e.name + "()" }
func (e *markerExpression) FunctionName() string { return e.name }
func (e *markerExpression) Description() string  { return "counts evaluations" }
func (e *markerExpression) WithChildren(_ *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(e, len(children), 0)
	}
	return e, nil
}
func (e *markerExpression) Eval(*sql.Context, sql.Row) (interface{}, error) {
	e.evals.Add(1)
	return int64(7), nil
}

type markerProvider struct {
	*memory.DbProvider
	evals *atomic.Int32
}

func (p *markerProvider) Function(_ *sql.Context, _, name string) (sql.Function, bool) {
	if name == "now" {
		name = "marker_update"
	}
	if !strings.HasPrefix(name, "marker_") {
		return nil, false
	}
	return sql.Function0{Name: name, Fn: func(*sql.Context) sql.Expression {
		return &markerExpression{Expression: expression.NewLiteral(int64(7), types.Int64), name: name, evals: p.evals}
	}}, true
}

type markerSession struct {
	sql.Session
	provider *markerProvider
}

func (s *markerSession) GenericProvider() sql.MutableDatabaseProvider { return s.provider }
func markerContext(evals *atomic.Int32) *sql.Context {
	ctx := sql.NewEmptyContext()
	db := memory.NewDatabase("mydb")
	db.AddTable("t", memory.NewLocalTable(ctx, db, "t", sql.NewPrimaryKeySchema(sql.Schema{
		{Name: "id", Type: types.Int64, PrimaryKey: true, Source: "t"},
	}), nil))
	provider := &markerProvider{DbProvider: memory.NewDBProvider(db), evals: evals}
	ctx = sql.NewContext(context.Background(), sql.WithSession(&markerSession{Session: ctx.Session, provider: provider}))
	ctx.SetCurrentDatabase("mydb")
	return ctx
}
