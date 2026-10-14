// Copyright 2020 Dolthub, Inc.
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

package writer_test

import (
	"fmt"
	"testing"

	"context"
	"github.com/dolthub/go-mysql-server/enginetest/queries"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/nustiueudinastea/dolt/go/cmd/dolt/commands/engine"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/dtestutils"
	"github.com/stretchr/testify/require"
)

func TestSecondaryIndexConflictOrder(t *testing.T) {
	for _, script := range []queries.ScriptTest{
		{
			Name: "duplicate update chooses the first index by name",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, a INT, b INT, n INT, UNIQUE KEY ub (b), UNIQUE KEY ua (a))",
				"INSERT INTO t VALUES (1,1,2,0),(2,2,1,0)",
				"INSERT INTO t VALUES (3,1,1,0) ON DUPLICATE KEY UPDATE n=n+1",
			},
			Assertions: []queries.ScriptTestAssertion{
				{Query: "SELECT id,n FROM t ORDER BY id", Expected: []sql.Row{{int32(1), int32(1)}, {int32(2), int32(0)}}},
			},
		},
		{
			Name: "replace with a delete trigger",
			SetUpScript: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, a INT, b INT, n INT, UNIQUE KEY ub (b), UNIQUE KEY ua (a))",
				"CREATE TABLE deleted (seq INT AUTO_INCREMENT PRIMARY KEY, id INT)",
				"CREATE TRIGGER log_delete AFTER DELETE ON t FOR EACH ROW INSERT INTO deleted(id) VALUES (OLD.id)",
				"INSERT INTO t VALUES (1,1,2,0),(2,2,1,0)",
				"REPLACE INTO t VALUES (3,1,1,0)",
			},
			Assertions: []queries.ScriptTestAssertion{
				// The pinned GMS REPLACE path does not invoke DELETE triggers. Keep
				// that behavior while checking both conflicting rows are removed.
				{Query: "SELECT seq,id FROM deleted ORDER BY seq", Expected: nil},
				{Query: "SELECT id,a,b,n FROM t", Expected: []sql.Row{{int32(3), int32(1), int32(1), int32(0)}}},
			},
		},
	} {
		t.Run(script.Name, func(t *testing.T) {
			for i := 0; i < 100; i++ {
				t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
					denv := dtestutils.CreateTestEnv()
					e, db, err := engine.NewSqlEngineForEnv(context.Background(), denv)
					require.NoError(t, err)
					defer e.Close()
					ctx, err := e.NewDefaultContext(context.Background())
					require.NoError(t, err)
					defer sql.SessionEnd(ctx.Session)
					ctx.SetCurrentDatabase(db)
					run := func(query string) []sql.Row {
						sql.SessionCommandBegin(ctx.Session)
						defer sql.SessionCommandEnd(ctx.Session)
						_, iter, _, err := e.Query(ctx, query)
						require.NoError(t, err, query)
						rows, err := sql.RowIterToRows(ctx, iter)
						require.NoError(t, err, query)
						return rows
					}
					for _, query := range script.SetUpScript {
						run(query)
					}
					for _, assertion := range script.Assertions {
						require.EqualValues(t, assertion.Expected, run(assertion.Query), assertion.Query)
					}
				})
			}
		})
	}
}
