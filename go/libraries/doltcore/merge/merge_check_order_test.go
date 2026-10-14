// Copyright 2023 Dolthub, Inc.
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
	"encoding/json"
	"io"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/doltdb"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/doltdb/durable"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/merge"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/table/editor"
	"github.com/nustiueudinastea/dolt/go/store/prolly"
	"github.com/stretchr/testify/require"
)

func TestMergeCheckOrder(t *testing.T) {
	t.Run("first evaluation error follows schema order", func(t *testing.T) {
		s := sch("CREATE TABLE t (id INT PRIMARY KEY, j JSON)")
		// Imported schemas can contain erroring expressions. Name order deliberately
		// disagrees with schema order; the first CHECK must always report 'bad'.
		_, err := s.sch.Checks().AddCheck("z_first", "json_extract(j, 'bad')", true, false)
		require.NoError(t, err)
		_, err = s.sch.Checks().AddCheck("a_second", "json_extract(j, '$[')", true, false)
		require.NoError(t, err)
		data := dataTest{
			ancestor: []sql.Row{{0, `{"x":1}`}},
			left:     []sql.Row{{0, `{"x":1}`}, {1, `{"x":1}`}},
			right:    []sql.Row{{0, `{"x":1}`}, {2, `{"x":1}`}},
		}
		ctx := sql.NewContext(context.Background())
		a, l, r, _ := setupDataMergeTest(ctx, t, s, data)
		for i := 0; i < 100; i++ {
			_, err = merge.MergeRoots(ctx, doltdb.SimpleTableResolver{}, l, r, a, rootish{r}, rootish{a}, editor.Options{}, merge.MergeOpts{})
			require.Error(t, err)
			require.Contains(t, err.Error(), "bad", "iteration %d", i)
		}
	})
	t.Run("record every failed check on each row", func(t *testing.T) {
		s := sch("CREATE TABLE t (id INT PRIMARY KEY, a INT, b INT)")
		_, err := s.sch.Checks().AddCheck("z_first", "(a > 0)", true, false)
		require.NoError(t, err)
		_, err = s.sch.Checks().AddCheck("a_second", "(b > 0)", true, false)
		require.NoError(t, err)
		data := dataTest{
			ancestor: []sql.Row{{0, 1, 1}},
			left:     []sql.Row{{0, 1, 1}, {1, -1, -1}},
			right:    []sql.Row{{0, 1, 1}, {2, -2, -2}},
		}
		ctx := sql.NewContext(context.Background())
		a, l, r, _ := setupDataMergeTest(ctx, t, s, data)
		result, err := merge.MergeRoots(ctx, doltdb.SimpleTableResolver{}, l, r, a, rootish{r}, rootish{a}, editor.Options{}, merge.MergeOpts{})
		require.NoError(t, err)
		table, ok, err := result.Root.GetTable(ctx, doltdb.TableName{Name: "t"})
		require.NoError(t, err)
		require.True(t, ok)
		count, err := table.NumConstraintViolations(ctx)
		require.NoError(t, err)
		require.EqualValues(t, 4, count)
		artifacts, err := table.GetArtifacts(ctx)
		require.NoError(t, err)
		iter, err := durable.ProllyMapFromArtifactIndex(artifacts).IterAllCVs(ctx)
		require.NoError(t, err)
		names := make(map[string]int)
		for {
			artifact, err := iter.Next(ctx)
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			var info prolly.ConstraintViolationMeta
			require.NoError(t, json.Unmarshal(artifact.Metadata, &info))
			var meta merge.CheckCVMeta
			require.NoError(t, json.Unmarshal(info.VInfo, &meta))
			names[meta.Name]++
		}
		require.Equal(t, map[string]int{"z_first": 2, "a_second": 2}, names)
	})
}
