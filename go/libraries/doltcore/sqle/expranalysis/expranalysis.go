// Copyright 2024 Dolthub, Inc.
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

package expranalysis

import (
	"fmt"

	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/planbuilder"
	"github.com/dolthub/go-mysql-server/sql/transform"

	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/doltdb"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/schema"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/sqle/overrides"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/sqle/sqlfmt"
)

// WithExpressionValidator returns a copy of ctx carrying v. A nil validator disables validation.
// The validator receives each resolved schema expression's complete tree, and must not evaluate it.
// Validation may occur more than once, including when a cached expression is evaluated in a new context.
func WithExpressionValidator(ctx *sql.Context, v func(*sql.Context, sql.Expression) error) *sql.Context {
	return overrides.WithExpressionValidator(ctx, v)
}

// ExpressionValidatorFromContext returns the optional schema expression validator carried by ctx.
func ExpressionValidatorFromContext(ctx *sql.Context) func(*sql.Context, sql.Expression) error {
	return overrides.ExpressionValidatorFromContext(ctx)
}

// ValidateExpression applies the context's validator to a resolved schema expression before use.
func ValidateExpression(ctx *sql.Context, expr sql.Expression) error {
	return overrides.ValidateExpression(ctx, expr)
}

// ResolveDefaultExpression returns a sql.Expression for the column default or generated expression for the
// column provided
func ResolveDefaultExpression(ctx *sql.Context, tableName string, sch schema.Schema, col schema.Column) (sql.Expression, error) {
	resolved, err := ResolveSchema(ctx, tableName, sch)
	if err != nil {
		return nil, err
	}

	colIdx := resolved.IndexOfColName(col.Name)
	if colIdx < 0 {
		return nil, fmt.Errorf("unable to find column %s in analyzed query", col.Name)
	}

	sqlCol := resolved[colIdx]
	expr := sqlCol.Default
	if expr == nil || expr.Expr == nil {
		expr = sqlCol.Generated
	}

	if expr == nil || expr.Expr == nil {
		return nil, fmt.Errorf("unable to find default or generated expression")
	}

	return expr, nil
}

// ResolveSchema returns the sql.Schema for |sch| with its column expressions resolved.
// All CHECK, DEFAULT, generated and ON UPDATE expressions are validated before returning the schema.
func ResolveSchema(ctx *sql.Context, tableName string, sch schema.Schema) (sql.Schema, error) {
	ct, err := parseCreateTable(ctx, tableName, sch)
	if err != nil {
		return nil, err
	}
	return ct.PkSchema().Schema, nil
}

// ResolveCheckExpression returns a sql.Expression for the check provided
func ResolveCheckExpression(ctx *sql.Context, tableName string, sch schema.Schema, checkExpr string) (sql.Expression, error) {
	ct, err := parseCreateTable(ctx, tableName, sch)
	if err != nil {
		return nil, err
	}

	formatter := overrides.SchemaFormatterFromContext(ctx)
	for _, check := range ct.Checks() {
		// Check definitions created before v1.55.3 may not have backquotes around identifiers
		quotedExpr := stripTableNamesFromExpression(ctx, formatter, check.Expr, true).String()
		unquotedExpr := stripTableNamesFromExpression(ctx, formatter, check.Expr, false).String()
		if quotedExpr == checkExpr || unquotedExpr == checkExpr {
			return check.Expr, nil
		}
	}

	return nil, fmt.Errorf("unable to find check expression")
}

// ResolveExpression compiles a predicate or check constraint string into a sql.Expression resolved.
// Used to re-hydrate partial index predicates and check expressions from their string representation.
// It uses SELECT statement on the expression FROM given table.
func ResolveExpression(ctx *sql.Context, tableName doltdb.TableName, predicateStr string) (sql.Expression, error) {
	// TODO: added * in SELECT stmt to avoid pruning columns during analyzer.
	parsed, err := parseQuery(ctx, fmt.Sprintf("SELECT *, %s FROM %s", predicateStr, queryTableName(ctx, tableName)))
	if err != nil {
		return nil, err
	}

	// The analyzer can evaluate constant expressions. Validate the planbuilder's resolved
	// expression first, while the original function tree is still available.
	if err = validateProjection(ctx, parsed); err != nil {
		return nil, err
	}

	sess, ok := ctx.Session.(SessionDbProvider)
	if ok {
		// field index in GetField is incorrect, it needs `fix_exec_indexes` analyzer rule
		a := analyzer.NewDefault(sess.GenericProvider())
		parsed, err = a.Analyze(ctx, parsed, nil, nil)
		if err != nil {
			return nil, err
		}
		// comes out in ContextRootFinalizer, so unwrap it
		parsed = parsed.Children()[0]
	}

	proj, ok := parsed.(*plan.Project)
	if !ok {
		return nil, fmt.Errorf("expected *plan.Project parsing predicate, got %T", parsed)
	}
	expr := proj.Projections[len(proj.Projections)-1]
	if alias, ok := expr.(*expression.Alias); ok {
		expr = alias.Child
	}
	if err = ValidateExpression(ctx, expr); err != nil {
		return nil, err
	}
	return expr, nil
}

func validateProjection(ctx *sql.Context, node sql.Node) error {
	// Avoid changing the no-validator path, including its analyzer diagnostics.
	if ExpressionValidatorFromContext(ctx) == nil {
		return nil
	}
	proj, ok := node.(*plan.Project)
	if !ok {
		return fmt.Errorf("expected *plan.Project parsing predicate, got %T", node)
	}
	return ValidateExpression(ctx, proj.Projections[len(proj.Projections)-1])
}

// queryTableName renders |tableName| for use in the FROM clause of a query built for the analyzer. The schema, where
// there is one, must survive into the parsed statement: an unqualified name is resolved against the session's
// search_path, which need not reach the table's own schema, and may name a different table of the same name first.
func queryTableName(ctx *sql.Context, tableName doltdb.TableName) string {
	if tableName.Schema == "" {
		return tableName.Name
	}
	formatter := overrides.SchemaFormatterFromContext(ctx)
	return fmt.Sprintf("%s.%s", formatter.QuoteIdentifier(tableName.Schema), formatter.QuoteIdentifier(tableName.Name))
}

func stripTableNamesFromExpression(ctx *sql.Context, formatter sql.SchemaFormatter, expr sql.Expression, quoted bool) sql.Expression {
	e, _, _ := transform.Expr(ctx, expr, func(ctx *sql.Context, e sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
		if col, ok := e.(*expression.GetField); ok {
			return col.WithTable("").WithQuotedNames(formatter, quoted), transform.NewTree, nil
		}
		return e, transform.SameTree, nil
	})
	return e
}

// SessionDbProvider is a way to access the Dolt database provider without import cycles.
type SessionDbProvider interface {
	GenericProvider() sql.MutableDatabaseProvider
}

func parseCreateTable(ctx *sql.Context, tableName string, sch schema.Schema) (*plan.CreateTable, error) {
	// This is a pseudo CREATE TABLE, used only to resolve default/generated/check expressions via the
	// engine's own SQL parser and analyzer, so SystemHidden columns are included here.
	createTable, err := sqlfmt.GenerateCreateTableStatement(ctx, tableName, sch, nil, nil, true)
	if err != nil {
		return nil, err
	}

	pseudoAnalyzedQuery, err := parseQuery(ctx, createTable)
	if err != nil {
		return nil, err
	}

	ct, ok := pseudoAnalyzedQuery.(*plan.CreateTable)
	if !ok {
		return nil, fmt.Errorf("expected a *plan.CreateTable node, but got %T", pseudoAnalyzedQuery)
	}
	if err = overrides.ValidateSchemaExpressions(ctx, ct.PkSchema().Schema, ct.Checks()); err != nil {
		return nil, err
	}
	return ct, nil
}

func parseQuery(ctx *sql.Context, query string) (sql.Node, error) {
	// Doltgres must use the existing provider due to custom functions, so we split the logic here depending on whether
	// this is being called from a query context (which will have a valid Dolt session) or some other context (which
	// will need to construct a new provider).
	var b *planbuilder.Builder
	engOverrides := overrides.EngineOverridesFromContext(ctx)
	sess, ok := ctx.Session.(SessionDbProvider)
	if ok {
		catalog := analyzer.NewCatalog(sess.GenericProvider(), engOverrides)
		catalog.AuthHandler = sql.NoopAuthorizationHandler{}
		b = planbuilder.New(overrides.PlanningContext(ctx), overrides.ValidatingCatalog(ctx, catalog), nil)
	} else {
		mockDatabase := memory.NewDatabase("mydb")
		mockProvider := memory.NewDBProvider(mockDatabase)
		catalog := analyzer.NewCatalog(mockProvider, engOverrides)
		catalog.AuthHandler = sql.NoopAuthorizationHandler{}
		// We need a new context for this operation
		parseCtx := sql.NewContext(ctx)
		parseCtx.SetCurrentDatabase("mydb")

		b = planbuilder.New(overrides.PlanningContext(parseCtx), overrides.ValidatingCatalog(parseCtx, catalog), nil)
	}

	pseudoAnalyzedQuery, _, _, _, err := b.Parse(query, nil, false)
	if err != nil {
		return nil, err
	}
	return pseudoAnalyzedQuery, nil
}
