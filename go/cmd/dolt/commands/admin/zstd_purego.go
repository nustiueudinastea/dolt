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

//go:build dolt_purego_zstd

package admin

import (
	"context"
	"fmt"

	"github.com/nustiueudinastea/dolt/go/cmd/dolt/cli"
	"github.com/nustiueudinastea/dolt/go/libraries/doltcore/env"
	"github.com/nustiueudinastea/dolt/go/libraries/utils/argparser"
	"github.com/fatih/color"
	kpzstd "github.com/klauspost/compress/zstd"
)

type ZstdCmd struct {
}

func (cmd ZstdCmd) Name() string {
	return "zstd"
}

func (cmd ZstdCmd) Description() string {
	return "A temporary admin command for taking a dependency on zstd and working out tooling dependencies."
}

func (cmd ZstdCmd) RequiresRepo() bool {
	return false
}

func (cmd ZstdCmd) Docs() *cli.CommandDocumentation {
	return nil
}

func (cmd ZstdCmd) ArgParser() *argparser.ArgParser {
	ap := argparser.NewArgParserWithMaxArgs(cmd.Name(), 0)
	return ap
}

func (cmd ZstdCmd) Hidden() bool {
	return true
}

func (cmd ZstdCmd) Exec(ctx context.Context, commandStr string, args []string, dEnv *env.DoltEnv, cliCtx cli.CliContext) int {
	enc, err := kpzstd.NewWriter(nil, kpzstd.WithEncoderConcurrency(1))
	if err != nil {
		fmt.Fprintf(color.Error, "zstd error: %v\n", err)
		return 1
	}
	fmt.Fprintf(color.Error, "Hello, world! compressed is %v\n", enc.EncodeAll([]byte("Hello, world!"), nil))

	return 0
}
