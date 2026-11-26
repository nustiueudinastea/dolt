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
	"testing"

	"github.com/nustiueudinastea/go-mysql-server/sql"
	"github.com/stretchr/testify/require"

	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/doltdb"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/merge"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/table/editor"
)

// TestUniqueValidatorFlushedIndexMove verifies that moving a UNIQUE key to an
// earlier primary key succeeds cleanly even when intervening edits flush the
// merge validator's private secondary index before the old row is deleted or
// reassigned.
func TestUniqueValidatorFlushedIndexMove(t *testing.T) {
	uniqSchema := sch("CREATE TABLE t (pk int PRIMARY KEY, code int, val int, UNIQUE KEY uq_code (code))")

	cases := []struct {
		name string
		data dataTest
	}{
		{
			name: "delete old row after private index flush",
			data: dataTest{
				name: "delete after flush",
				ancestor: []sql.Row{
					sql.NewRow(10, 100, 1),
				},
				left: []sql.Row{
					sql.NewRow(10, 100, 1),
					sql.NewRow(30, 300, 3),
				},
				right: []sql.Row{
					sql.NewRow(5, 100, 5),
					sql.NewRow(6, 106, 6),
					sql.NewRow(7, 107, 7),
					sql.NewRow(8, 108, 8),
				},
			},
		},
		{
			name: "modify old row unique key after private index flush",
			data: dataTest{
				name: "modify after flush",
				ancestor: []sql.Row{
					sql.NewRow(10, 100, 1),
				},
				left: []sql.Row{
					sql.NewRow(10, 100, 1),
					sql.NewRow(30, 300, 3),
				},
				right: []sql.Row{
					sql.NewRow(5, 100, 5),
					sql.NewRow(6, 106, 6),
					sql.NewRow(7, 107, 7),
					sql.NewRow(8, 108, 8),
					sql.NewRow(10, 200, 1),
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			a, l, r, _ := setupDataMergeTest(ctx, t, uniqSchema, tc.data)

			var eo editor.Options
			mo := merge.MergeOpts{
				UniqueIndexPendingLimit: 1,
			}
			res, err := merge.MergeRoots(sql.NewContext(ctx), doltdb.SimpleTableResolver{}, l, r, a, rootish{r}, rootish{a}, eo, mo)
			require.NoError(t, err)
			require.False(t, res.HasMergeArtifacts(), "unexpected merge artifacts: %+v", res.Stats)
			require.Zero(t, res.CountOfTablesWithConstraintViolations())

			tbl, ok, err := res.Root.GetTable(ctx, doltdb.TableName{Name: "t"})
			require.NoError(t, err)
			require.True(t, ok)
			arts, err := tbl.GetArtifacts(ctx)
			require.NoError(t, err)
			cnt, err := arts.Count()
			require.NoError(t, err)
			require.Zero(t, cnt, "expected zero leftover unique violation artifacts")
		})
	}
}
