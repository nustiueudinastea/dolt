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
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Review every fresh SQL context. A caller's execution policies must survive;
// independent entry points and helpers that cannot execute SQL are explicit exemptions.
func TestFreshContextInventory(t *testing.T) {
	root := "../../../.."
	const inherits = "propagates caller policies"
	want := map[string][]string{
		"cmd/dolt/commands/sqlserver/server.go:newSessionBuilder":                                    {"independent client login session; no SQL caller"},
		"libraries/doltcore/sqle/testutil.go:NewTestSQLCtxWithProvider":                              {"test harness"},
		"libraries/doltcore/dtestutils/testcommands/multienv.go:CommitWithWorkingSet":                {"test harness"},
		"libraries/doltcore/sqle/dsess/session_state_adapter.go:TempTableFilesDir":                   {"session metadata lookup; no expression resolution or evaluation"},
		"libraries/doltcore/sqle/expranalysis/expranalysis.go:parseQuery":                            {inherits},
		"libraries/doltcore/sqle/show_create_table.go:PrepareCreateTableStmt":                        {inherits},
		"libraries/doltcore/sqle/binlogreplication/binlog_replica_applier.go:executeQueryWithEngine": {inherits},
		"libraries/doltcore/table/typed/json/writer.go:<package>":                                    {"JSON type serialization only"},
		"cmd/dolt/commands/sql.go:newCompleter":                                                      {inherits},
		"cmd/dolt/commands/engine/sqlengine.go:sqlContextFactory":                                    {inherits},
		"cmd/dolt/commands/schcmds/tags.go:Exec":                                                     {inherits},
		"cmd/dolt/commands/indexcmds/rebuild.go:Exec":                                                {inherits},
		"cmd/dolt/commands/indexcmds/cat.go:prettyPrintResults":                                      {"index row formatting only"},
		"cmd/dolt/commands/filter-branch.go:rebaseSqlEngine":                                         {inherits},
		"cmd/dolt/commands/sqlserver/queryist_utils.go:BuildConnectionStringQueryist":                {inherits},
		"cmd/dolt/commands/sqlserver/server.go:ConfigureServices":                                    {"server global variable initialization", "server global variable initialization", inherits},
	}
	found := make(map[string][]string)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		require.NoError(t, err)
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		require.NoError(t, err)
		rel = filepath.ToSlash(rel)
		if strings.Contains(rel, "/enginetest/") || strings.Contains(rel, "/logictest/") {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		require.NoError(t, err)
		for _, decl := range f.Decls {
			name := "<package>"
			if fn, ok := decl.(*ast.FuncDecl); ok {
				name = fn.Name.Name
			}
			ast.Inspect(decl, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				fresh := false
				switch fn := call.Fun.(type) {
				case *ast.Ident:
					fresh = fn.Name == "NewContext" || fn.Name == "NewEmptyContext"
				case *ast.SelectorExpr:
					fresh = fn.Sel.Name == "NewContext" || fn.Sel.Name == "NewEmptyContext"
				}
				if !fresh {
					return true
				}
				key := rel + ":" + name
				entries := want[key]
				classification := "UNCLASSIFIED"
				if len(found[key]) < len(entries) {
					classification = entries[len(found[key])]
				}
				if classification == inherits {
					require.NotEmpty(t, call.Args, key)
					arg, ok := call.Args[0].(*ast.Ident)
					require.True(t, ok, "%s must pass the caller itself, not its embedded context", key)
					if ok {
						require.Contains(t, []string{"ctx", "subCtx"}, arg.Name, key)
					}
				}
				found[key] = append(found[key], classification)
				return true
			})
		}
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, want, found, "classify every new context creation site")
}
