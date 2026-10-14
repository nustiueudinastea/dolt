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
