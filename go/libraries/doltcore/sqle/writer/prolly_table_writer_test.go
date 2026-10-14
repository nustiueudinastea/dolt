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
	"context"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"
)

// Embedding indexWriter leaves the storage methods unused; these tests exercise
// the table writer's ordering of validation, edits and statement completion.
type recordingIndexWriter struct {
	indexWriter
	name  string
	calls *[]string
}

func (w recordingIndexWriter) record(op string) error {
	*w.calls = append(*w.calls, op+":"+w.name)
	return nil
}
func (w recordingIndexWriter) ValidateKeyViolations(context.Context, sql.Row) error {
	return w.record("validate")
}
func (w recordingIndexWriter) Insert(context.Context, sql.Row) error { return w.record("insert") }
func (w recordingIndexWriter) Update(context.Context, sql.Row, sql.Row) error {
	return w.record("update")
}
func (w recordingIndexWriter) Delete(context.Context, sql.Row) error { return w.record("delete") }
func (w recordingIndexWriter) Commit(context.Context) error          { return w.record("commit") }
func (w recordingIndexWriter) Discard(context.Context) error         { return w.record("discard") }

func TestSecondaryIndexWriterOrder(t *testing.T) {
	ctx := sql.NewEmptyContext()
	for i := 0; i < 100; i++ {
		var calls []string
		w := prollyTableWriter{
			primary: recordingIndexWriter{name: "primary", calls: &calls},
			secondary: map[string]indexWriter{
				"ub": recordingIndexWriter{name: "ub", calls: &calls},
				"ua": recordingIndexWriter{name: "ua", calls: &calls},
			},
		}
		require.NoError(t, w.Insert(ctx, nil))
		require.NoError(t, w.Update(ctx, nil, nil))
		require.NoError(t, w.Delete(ctx, nil))
		require.NoError(t, w.StatementComplete(ctx))
		require.NoError(t, w.DiscardChanges(ctx, nil))
		require.Equal(t, []string{
			"validate:primary", "validate:ua", "validate:ub", "insert:ua", "insert:ub", "insert:primary",
			"update:ua", "update:ub", "update:primary", "delete:ua", "delete:ub", "delete:primary",
			"commit:primary", "commit:ua", "commit:ub", "discard:primary", "discard:ua", "discard:ub",
		}, calls, "iteration %d", i)
	}
}
