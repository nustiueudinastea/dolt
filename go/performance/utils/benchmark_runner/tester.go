// Copyright 2019-2022 Dolthub, Inc.
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

package benchmark_runner

import "context"

type Tester interface {
	Test(ctx context.Context) (*Result, error)
}

type Test interface {
	GetId() string
	GetName() string
	GetParamsToSlice() []string
	GetPrepareArgs(serverConfig ServerConfig) []string
	GetRunArgs(serverConfig ServerConfig) []string
	GetCleanupArgs(serverConfig ServerConfig) []string
}

type SysbenchTest interface {
	Test
	GetFromScript() bool
}

type TestParams interface {
	ToSlice() []string
}

type SysbenchTestParams interface {
	TestParams
	Append(params ...string)
}

type TpccTestParams interface {
	TestParams
	GetNumThreads() int
	GetScaleFactor() int
	GetTables() int
	GetTrxLevel() string
	GetReportCSV() bool
	GetReportInterval() int
	GetTime() int
}
