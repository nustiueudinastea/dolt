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

package overrides

import (
	"context"
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
)

// The context value lives here so sqlutil's schema importer and expranalysis can
// share validation without introducing an import cycle through sqlfmt.
type expressionValidatorKey struct{}

// WithExpressionValidator returns a copy of ctx carrying v. A nil validator disables validation.
func WithExpressionValidator(ctx *sql.Context, v func(*sql.Context, sql.Expression) error) *sql.Context {
	return ctx.WithContext(context.WithValue(ctx.Context, expressionValidatorKey{}, v))
}

// ExpressionValidatorFromContext returns the optional schema expression validator carried by ctx.
func ExpressionValidatorFromContext(ctx *sql.Context) func(*sql.Context, sql.Expression) error {
	v, _ := ctx.Value(expressionValidatorKey{}).(func(*sql.Context, sql.Expression) error)
	return v
}

// ValidateExpression applies the context's validator to a resolved schema expression before use.
func ValidateExpression(ctx *sql.Context, expr sql.Expression) error {
	if v := ExpressionValidatorFromContext(ctx); v != nil {
		if err := v(ctx, expr); err != nil {
			return fmt.Errorf("invalid schema expression: %w", err)
		}
	}
	return nil
}

// ValidateSchemaExpressions validates all resolved column expressions and CHECKs,
// including virtual columns and unenforced CHECKs, before any can be used.
func ValidateSchemaExpressions(ctx *sql.Context, sch sql.Schema, checks sql.CheckConstraints) error {
	if ExpressionValidatorFromContext(ctx) == nil {
		return nil
	}
	for _, col := range sch {
		for _, expr := range []*sql.ColumnDefaultValue{col.Default, col.Generated, col.OnUpdate} {
			if expr != nil && expr.Expr != nil {
				if err := ValidateExpression(ctx, expr); err != nil {
					return err
				}
			}
		}
	}
	for _, check := range checks {
		if err := ValidateExpression(ctx, check.Expr); err != nil {
			return err
		}
	}
	return nil
}

// PlanningContext connects D3 to engines exposing planning validators, without
// requiring that optional API from upstream GMS. Preserve both caller validators.
func PlanningContext(ctx *sql.Context) *sql.Context {
	if ExpressionValidatorFromContext(ctx) == nil {
		return ctx
	}
	if original, ok := interface{}(ctx).(contextValidators); ok {
		derived := ctx.WithContext(ctx.Context)
		interface{}(derived).(contextValidators).SetValidators(func(_ *sql.Context, node sql.Node) error { return original.ValidateNode(node) }, func(got *sql.Context, expr sql.Expression) error {
			if err := original.ValidateExpression(expr); err != nil {
				return err
			}
			return ValidateExpression(got, expr)
		})
		return derived
	}
	return ctx
}

type contextValidators interface {
	SetValidators(func(*sql.Context, sql.Node) error, func(*sql.Context, sql.Expression) error)
	ValidateNode(sql.Node) error
	ValidateExpression(sql.Expression) error
}

// ValidatingCatalog checks resolved function arguments before a constructor can
// evaluate or discard them. Keeping this adapter here also supports upstream
// engines without context validators.
func ValidatingCatalog(ctx *sql.Context, catalog sql.Catalog) sql.Catalog {
	if ExpressionValidatorFromContext(ctx) == nil {
		return catalog
	}
	if _, ok := interface{}(ctx).(contextValidators); ok {
		return catalog
	}
	return validatingCatalog{Catalog: catalog}
}

type validatingCatalog struct{ sql.Catalog }

func (c validatingCatalog) Function(ctx *sql.Context, schema, name string) (sql.Function, bool) {
	f, ok := c.Catalog.Function(ctx, schema, name)
	if !ok {
		return nil, false
	}
	return validatingFunction{Function: f}, true
}

type validatingFunction struct{ sql.Function }

func (f validatingFunction) NewInstance(ctx *sql.Context, args []sql.Expression) (sql.Expression, error) {
	for _, arg := range args {
		if err := ValidateExpression(ctx, arg); err != nil {
			return nil, err
		}
	}
	return f.Function.NewInstance(ctx, args)
}
