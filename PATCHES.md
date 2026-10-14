# Maintained patches

Upstream base: `eaa9030af9088e518040479e92e40b1136f5d758` (merge base of the original fork head
`27fbb50617bf36638130b0ba01521fc3c1ef3239` and `upstream/main`). Branch: `swarmion`.
Each behavior patch has its own commit. The D1/D2 inventory was added separately;
the D3 inventory is included in its behavior patch.

## Existing fork commits

These are the original 13 commits, oldest first, from `git log --reverse --oneline
$(git merge-base 27fbb5061 upstream/main)..27fbb5061`:

- `38b3aca29` Add pure-Go zstd and WASM build shims
- `80a3dd9a4` cmd wasm
- `52c84853b` wasm: avoid native lock dependencies in browser builds
- `a6153c4a6` Publish maintained Dolt module path
- `8cef6e4df` Use maintained Dolt MCP module
- `d0714559d` Add expected-root batch commit-head publication
- `e2907f329` Add atomic commit-head and working-set-head publication
- `280cf9d84` Use the zstd adapter in new upstream archive tests
- `6228cbc04` Regenerate maintained protobuf descriptors after the upstream generator update
- `b21d4ed0c` Preserve browser lock and subprocess shims with the new upstream APIs
- `f4783f73c` Supply browser shims for upstream Git remote cleanup helpers
- `e8afbcc55` Keep native Git remote cache lookup unavailable in browser builds
- `27fbb5061` Make embedded engine shutdown safe for repeated and concurrent callers

## D1 — Secondary-index writer order

Commit: `75b006f452f0cb7dc5c7d5965be1e04e0c2738c7`.

Files:
- `go/libraries/doltcore/sqle/writer/prolly_table_writer.go`
- `go/libraries/doltcore/sqle/writer/prolly_table_writer_test.go`
- `go/libraries/doltcore/sqle/writer/secondary_index_order_test.go`

The primary index retains its existing priority. Secondary indexes are visited
in bytewise lexical index-name order, matching `schema.AllIndexes()`, for unique
validation, insertion, updates, deletion, unique-key change detection,
commit/discard, GC-root traversal, and index materialization. Reset caches the
names; directly assembled writers initialize the same order lazily. This avoids
map-order-dependent ODKU row selection and REPLACE duplicate retries. Index
names remain stable after schema reloads, unlike a Go map's traversal order.

MySQL does not promise which conflicting row ODKU selects with multiple unique
indexes; this is an explicit engine tie-breaker, rather than a claim that MySQL
uses name order. See [MySQL ODKU](https://dev.mysql.com/doc/refman/8.4/en/insert-on-duplicate.html).

Tests: `TestSecondaryIndexWriterOrder` checks validation/edit/lifecycle order over
100 fresh writers. `TestSecondaryIndexConflictOrder` runs the audit's unique-a,
unique-b ODKU scenario and a REPLACE with an AFTER DELETE trigger 100 times each.
The ODKU consistently increments row 1 and REPLACE removes both conflicting rows.
**Pinned GMS does not invoke DELETE triggers on its REPLACE path**
(`sql/rowexec/insert.go`, replacer loop); the trigger log stays empty. This test
preserves that existing difference from MySQL; implementing missing trigger
execution is outside this ordering patch. The writer test independently observes
ordered deletions.

Parent evidence: copy these two test files onto detached `27fbb5061` and run
`TestSecondaryIndex(ConflictOrder|WriterOrder)`. Both tests fail: writer callbacks
visit `ub` before `ua`, and ODKU increments row 2 in some iterations. The
configured-trigger REPLACE case remains unchanged on the parent.

## D2 — Merge CHECK order

Commit: `bba11cf70faacf270496ab3f20ac12f0fe6af776`.

Files:
- `go/libraries/doltcore/merge/merge_prolly_rows.go`
- `go/libraries/doltcore/merge/merge_check_order_test.go`

Compile and evaluate enforced CHECKs in `AllChecks()` schema order. Store named
expressions in a slice instead of converting the already ordered schema into a
map. This stabilizes evaluation and the first error, including imported schemas
whose expressions can error. Every false CHECK still records its own violation;
NULL continues to satisfy a CHECK. MySQL does not specify evaluation order for
multiple CHECKs; schema order is the fork's explicit rule.

Test: `TestMergeCheckOrder` repeats a merge with two erroring JSON-path CHECKs
100 times, expecting the first schema CHECK's `bad` path error. Check names are
chosen so lexical order disagrees with schema order. A second case merges two
rows failing both pure CHECKs and verifies all four artifacts and both names.

Parent evidence: the test copied onto detached D1 `75b006f45` fails on iteration 1:
it reports the second CHECK's missing-`]` error instead of the first CHECK's
`bad` error. The four-artifact case already passes on the parent, as required
for preservation of existing violation recording.

## D3 — Validate schema expressions resolved outside the analyzer

Commit: this D3 commit, immediately after `87de2a45962b51e45481c62efc943fb737274440`.

Files:
- `go/libraries/doltcore/sqle/expranalysis/{expranalysis,expranalysis_test}.go`
- `go/libraries/doltcore/sqle/overrides/expression_validation.go`
- `go/libraries/doltcore/sqle/sqlutil/schema.go`
- `go/libraries/doltcore/merge/{merge_prolly_rows,merge_expression_validator_test,expression_validation_test}.go`
- `go/libraries/doltcore/sqle/index/key_builder.go`
- `go/libraries/doltcore/sqle/writer/{prolly_index_writer,prolly_index_writer_keyless,prolly_fk_indexer,expression_validation_test}.go`
- `go/libraries/doltcore/table/editor/creation/external_build_index.go`

API: `expranalysis.WithExpressionValidator(ctx, func(*sql.Context, sql.Expression) error)`
returns a copied SQL context with an optional validator in its underlying context.
`ExpressionValidatorFromContext` retrieves it; `ValidateExpression` invokes it and
wraps rejection errors with `%w`, preserving `errors.Is`. Passing nil disables it.
The callback receives the complete resolved expression tree, including the column
value wrapper, and can inspect its children. It must not evaluate expressions and
may be called repeatedly. With no validator, existing behavior is preserved.
The shared implementation lives in `overrides` to avoid the existing
`expranalysis -> sqlfmt -> sqlutil` import cycle. No new GMS API is referenced.

Resolution sites covered:
- `expranalysis.parseCreateTable`, shared by `ResolveSchema`,
  `ResolveDefaultExpression` and `ResolveCheckExpression`: validate every DEFAULT,
  stored/virtual generated value, ON UPDATE value, and enforced/unenforced CHECK.
- `expranalysis.ResolveExpression`: validate the planbuilder's resolved predicate
  before running the analyzer, which can evaluate constants, and validate the
  resulting expression before returning it. Existing analyzer diagnostics remain
  unchanged when the callback accepts the expression or is absent.
- `sqlutil.ParseCreateTableStatement`: schema-file import planbuilds CREATE TABLE
  directly; validate all column expressions and CHECKs before converting to Dolt.
- Existing callers `merge.resolveDefaults`, `merge.newCheckValidator`,
  `writer.resolveVirtualExpressions`, `writer.newWriterSchema` (partial predicates),
  `index.NewSecondaryKeyBuilder`, and `AlterableDoltTable.createIndex` already route
  through those resolvers and now inherit validation.

Direct evaluation sites also validate in the evaluating context, so previously
resolved or session-cached expressions cannot bypass a newly installed validator:
- `merge.checkValidator.validateDiff`: validate all CHECKs before remapping can
  evaluate defaults; `merge.writeTupleExpression`: defaults and stored generated values.
- `index.SecondaryKeyBuilder.SecondaryKeyFromRow`: virtual generated index keys.
- `writer.prollySecondaryIndexWriter.keyPartFromRow` and
  `writer.prollyKeylessSecondaryWriter.keyPartFromRow`: virtual generated index keys.
- `writer.prollySecondaryIndexWriter.matchesPredicate`,
  `writer.prollyKeylessSecondaryWriter.validateUniqueKeyViolation`, and
  `writer.prollyKeylessSecondaryWriter.Insert`: partial-index predicates.
- `writer.prollyFkPkRowIter.Next` and `writer.prollyFkKeylessRowIter.Next`: validate
  all cached virtual projections before delegating to `sql.EvalProjections`.
- `creation.BuildProllyIndexExternal`: partial-index predicate evaluation.

The audit also examined `merge.ColumnsFromCheckDefinition` (AST-only column-name
extraction) and `sqlutil.FromDoltSchema` (unresolved string placeholders); neither
resolves or evaluates a SQL expression. Ordinary analyzed query expressions are
outside this hook's scope.

Tests: `TestResolveSchemaExpressionValidator` checks all six marker-bearing
expressions, individual rejections through all three schema resolvers, zero Eval
calls, context isolation, nil opt-out, and unchanged default/CHECK results.
`TestResolveExpressionValidator` checks rejection before analysis and unchanged
accepted/absent-validator behavior. `TestParseCreateTableExpressionValidator`
checks the import bypass, all four expression kinds, and unchanged imported schema.
`TestMergeExpressionValidator` performs real three-way merges that fill a default
or evaluate CHECKs: rejection preserves the error identity and keeps Eval at zero;
without a validator the functions run and merged rows match the expected root.
`TestMergeValidatesCachedExpressions` and `TestWriterValidatesCachedExpressions`
exercise direct evaluation under a new context, plus no-validator controls.

Parent evidence: copy the four test files onto detached `87de2a459` and run the
six tests above. The exact parent cannot compile the new API. To demonstrate
behavioral failures as well, add only the context setter/getter definitions (no
validation call sites) to the parent. All six tests fail: schema/import callbacks
see zero markers; merge default filling evaluates once, CHECK validation twice,
and cached merge/writer expressions once instead of zero. All six pass with D3.
The compatibility file and both failure logs are retained locally under
`.git/patch-evidence/d3/`, alongside the complete package-check log.

## Validation

Run from `go/`, using default Go caches and the clone's default commit identity
and signing configuration:

```sh
go test -p 2 -count=1 -tags=dolt_purego_zstd,gms_pure_go \
  ./libraries/doltcore/sqle/expranalysis ./libraries/doltcore/merge \
  ./libraries/doltcore/sqle/index ./libraries/doltcore/sqle/writer \
  ./libraries/doltcore/table/editor/creation ./libraries/doltcore/sqle/overrides \
  ./libraries/doltcore/sqle/sqlutil
```

All listed packages pass (the creation package has no matching tests; overrides
and sqlutil have no test files). The import regression runs in expranalysis. Every
changed Go file was formatted with `gofmt`; `git diff --check` passes. Parent
regression runs used detached worktrees inside this clone, with the relevant
test files and the D3 API compatibility definitions described above. Local logs
are retained in `.git/patch-evidence/`.
No pushes, branch rewrites, or changes to other branches were made.

### Review follow-up: caller contexts during resolution

The base-session expression resolver derives its fresh SQL context from the caller,
retaining D3 context values and G7/G8 policies when used with the Swarmion GMS fork.
SQL completion and the SQL context factory preserve the SQL caller when adding
signal cancellation or storage context validation. Other schema, replication and
command helpers already pass the caller to `sql.NewContext`.

`overrides.TestFreshContextInventory` classifies all production fresh-context calls
outside test harnesses. Metadata, JSON serialization, independent login sessions
and server initialization are exempt from caller expression execution. Scheduled
event session factories are independent entry points. The base-session regression
fails before the fix with local GMS, then observes the caller flag, D3 validator
and G8 rejection. It skips when linked to upstream GMS without the optional API.

### Review follow-up: validate before schema function construction

D3 is connected to G8's planbuilder validation when the engine exposes that
optional context API. The adapter copies the context and preserves both original
G8 validators. Against upstream GMS, a catalog adapter validates resolved function
arguments before their constructors run. Both the schema-file importer and all
expranalysis parser entry points use these adapters, with the caller context.

`TestRejectSleepBeforeDefaultConstructor` imports the review's
`DEFAULT (utc_timestamp(SLEEP(0)))` through a counting SLEEP implementation. Before
the fix SLEEP ran once and vanished from the expression tree; after the fix the
validator rejects it and the Eval count is zero, with upstream and local GMS.
Existing schema tests assert coverage without requiring a single callback: G8 may
validate an expression repeatedly during construction and analysis.
