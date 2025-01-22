// Copyright 2022 Dolthub, Inc.
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

package dprocedures

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/dolt/go/cmd/dolt/cli"
	"github.com/dolthub/dolt/go/libraries/doltcore/branch_control"
	"github.com/dolthub/dolt/go/libraries/doltcore/dconfig"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/types"
)

const (
	cmdFailure = 1
	cmdSuccess = 0
)

func init() {
	if os.Getenv(dconfig.EnvDisableGcProcedure) != "" {
		DoltGCFeatureFlag = false
	}
}

var DoltGCFeatureFlag = true

func NewDoltGCProcedure(gcSafepointController *dsess.GCSafepointController) sql.ExternalStoredProcedureDetails {
	impl := &DoltGCProcedure{
		gcSafepointController: gcSafepointController,
	}
	return sql.ExternalStoredProcedureDetails{
		Name:      "dolt_gc",
		Schema:    int64Schema("status"),
		Function:  impl.Run,
		ReadOnly:  true,
		AdminOnly: true,
	}
}

// doltGC is the stored procedure to run online garbage collection on a database.
func (p *DoltGCProcedure) Run(ctx *sql.Context, args ...string) (sql.RowIter, error) {
	if !DoltGCFeatureFlag {
		return nil, errors.New("DOLT_GC() stored procedure disabled")
	}
	res, err := p.doGC(ctx, args)
	if err != nil {
		return nil, err
	}
	return rowToIter(int64(res)), nil
}

var ErrServerPerformedGC = errors.New("this connection was established when this server performed an online garbage collection. this connection can no longer be used. please reconnect.")

// The original behavior safepoint controller, which kills all connections right as the GC process is being finalized.
// The only connection which is left up is the connection on which dolt_gc is called, but that connection is
// invalidated in such a way that all future queries on it return an error.
type killConnectionsSafepointController struct {
	callCtx   *sql.Context
	origEpoch int
}

func (sc killConnectionsSafepointController) BeginGC(ctx context.Context, keeper func(hash.Hash) bool) error {
	return nil
}

func (sc killConnectionsSafepointController) EstablishPreFinalizeSafepoint(ctx context.Context) error {
	return nil
}

func (sc killConnectionsSafepointController) EstablishPostFinalizeSafepoint(ctx context.Context) error {
	// Here we need to sanity check role and epoch.
	if sc.origEpoch != -1 {
		if _, role, ok := sql.SystemVariables.GetGlobal(dsess.DoltClusterRoleVariable); ok {
			if role.(string) != "primary" {
				return fmt.Errorf("dolt_gc failed: when we began we were a primary in a cluster, but now our role is %s", role.(string))
			}
			_, epoch, ok := sql.SystemVariables.GetGlobal(dsess.DoltClusterRoleEpochVariable)
			if !ok {
				return fmt.Errorf("dolt_gc failed: when we began we were a primary in a cluster, but we can no longer read the cluster role epoch.")
			}
			if sc.origEpoch != epoch.(int) {
				return fmt.Errorf("dolt_gc failed: when we began we were primary in the cluster at epoch %d, but now we are at epoch %d. for gc to safely finalize, our role and epoch must not change throughout the gc.", sc.origEpoch, epoch.(int))
			}
		} else {
			return fmt.Errorf("dolt_gc failed: when we began we were a primary in a cluster, but we can no longer read the cluster role.")
		}
	}

	killed := make(map[uint32]struct{})
	processes := sc.callCtx.ProcessList.Processes()
	for _, p := range processes {
		if p.Connection != sc.callCtx.Session.ID() {
			// Kill any inflight query.
			sc.callCtx.ProcessList.Kill(p.Connection)
			// Tear down the connection itself.
			sc.callCtx.KillConnection(p.Connection)
			killed[p.Connection] = struct{}{}
		}
	}

	// Look in processes until the connections are actually gone.
	params := backoff.NewExponentialBackOff()
	params.InitialInterval = 1 * time.Millisecond
	params.MaxInterval = 25 * time.Millisecond
	params.MaxElapsedTime = 3 * time.Second
	err := backoff.Retry(func() error {
		processes := sc.callCtx.ProcessList.Processes()
		allgood := true
		for _, p := range processes {
			if _, ok := killed[p.Connection]; ok {
				allgood = false
				sc.callCtx.ProcessList.Kill(p.Connection)
			}
		}
		if !allgood {
			return errors.New("unable to establish safepoint.")
		}
		return nil
	}, params)
	if err != nil {
		return err
	}
	sc.callCtx.Session.SetTransaction(nil)
	dsess.DSessFromSess(sc.callCtx.Session).SetValidateErr(ErrServerPerformedGC)
	return nil
}

func (sc killConnectionsSafepointController) CancelSafepoint() {
}

type DoltGCProcedure struct {
	// Used by the implementation to visit existing sessions, find them
	// at a quiesced state and ensure that their in-memory state makes
	// it to the GC process.
	gcSafepointController *dsess.GCSafepointController
}

func (*DoltGCProcedure) doGC(ctx *sql.Context, args []string) (int, error) {
	dbName := ctx.GetCurrentDatabase()

	if len(dbName) == 0 {
		return cmdFailure, fmt.Errorf("Empty database name.")
	}
	if err := branch_control.CheckAccess(ctx, branch_control.Permissions_Write); err != nil {
		return cmdFailure, err
	}

	apr, err := cli.CreateGCArgParser().Parse(args)
	if err != nil {
		return cmdFailure, err
	}

	if apr.NArg() != 0 {
		return cmdFailure, InvalidArgErr
	}

	dSess := dsess.DSessFromSess(ctx.Session)
	ddb, ok := dSess.GetDoltDB(ctx, dbName)
	if !ok {
		return cmdFailure, fmt.Errorf("Could not load database %s", dbName)
	}

	if apr.Contains(cli.ShallowFlag) && apr.Contains(cli.FullFlag) {
		return cmdFailure, fmt.Errorf("cannot supply both --shallow and --full to dolt_gc: %w", InvalidArgErr)
	}

	if apr.Contains(cli.ShallowFlag) {
		err = ddb.ShallowGC(ctx)
		if err != nil {
			return cmdFailure, err
		}
	} else {
		// Currently, if this server is involved in cluster
		// replication, a full GC is only safe to run on the primary.
		// We assert that we are the primary here before we begin, and
		// we assert again that we are the primary at the same epoch as
		// we establish the safepoint.
		origepoch := -1
		if _, role, ok := sql.SystemVariables.GetGlobal(dsess.DoltClusterRoleVariable); ok {
			// TODO: magic constant...
			if role.(string) != "primary" {
				return cmdFailure, fmt.Errorf("cannot run a full dolt_gc() while cluster replication is enabled and role is %s; must be the primary", role.(string))
			}
			_, epoch, ok := sql.SystemVariables.GetGlobal(dsess.DoltClusterRoleEpochVariable)
			if !ok {
				return cmdFailure, fmt.Errorf("internal error: cannot run a full dolt_gc(); cluster replication is enabled but could not read %s", dsess.DoltClusterRoleEpochVariable)
			}
			origepoch = epoch.(int)
		}

		var mode types.GCMode = types.GCModeDefault
		if apr.Contains(cli.FullFlag) {
			mode = types.GCModeFull
		}

		sc := killConnectionsSafepointController{
			origEpoch: origepoch,
			callCtx:   ctx,
		}
		err = ddb.GC(ctx, mode, sc)
		if err != nil {
			return cmdFailure, err
		}
	}

	return cmdSuccess, nil
}
