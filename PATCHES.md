# Maintained patches

Upstream base: `eaa9030af9088e518040479e92e40b1136f5d758` (merge base of the original fork head
`27fbb50617bf36638130b0ba01521fc3c1ef3239` and `upstream/main`). Branch: `swarmion`.
Each behavior patch has its own commit; this inventory is a separate documentation commit.

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

## Validation

Run from `go/`, using default Go caches and the clone's default commit identity
and signing configuration:

```sh
go test -p 2 -count=1 -tags=dolt_purego_zstd,gms_pure_go ./libraries/doltcore/merge ./libraries/doltcore/sqle/writer
```

Both complete packages pass, including the new SQL and writer tests. Every
changed Go file was formatted with `gofmt`; `git diff --check` passes. Parent
regression runs used detached worktrees inside this clone, with only the relevant
test files copied in. Local logs are retained in `.git/patch-evidence/`.
No pushes, branch rewrites, or changes to other branches were made.
