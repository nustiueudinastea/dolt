//go:build js && wasm

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

package dbfactory

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"google.golang.org/grpc"

	"github.com/dolthub/dolt/go/libraries/doltcore/grpcendpoint"
	"github.com/dolthub/dolt/go/libraries/utils/earl"
	"github.com/dolthub/dolt/go/store/datas"
	"github.com/dolthub/dolt/go/store/prolly/tree"
	"github.com/dolthub/dolt/go/store/types"
)

const (
	AWSScheme      = "aws"
	GSScheme       = "gs"
	AzScheme       = "az"
	OCIScheme      = "oci"
	FileScheme     = "file"
	MemScheme      = "mem"
	HTTPSScheme    = "https"
	HTTPScheme     = "http"
	SSHScheme      = "ssh"
	LocalBSScheme  = "localbs"
	OSSScheme      = "oss"
	GitFileScheme  = "git+file"
	GitHTTPScheme  = "git+http"
	GitHTTPSScheme = "git+https"
	GitSSHScheme   = "git+ssh"

	defaultScheme       = MemScheme
	defaultMemTableSize = 256 * 1024 * 1024
)

type DBFactory interface {
	CreateDB(context.Context, *types.NomsBinFormat, *url.URL, map[string]interface{}) (datas.Database, types.ValueReadWriter, tree.NodeStore, error)
	PrepareDB(context.Context, *types.NomsBinFormat, *url.URL, map[string]interface{}) error
}

var DBFactories = map[string]DBFactory{
	MemScheme: MemFactory{},
}

var GRPCDialProviderParam = "__DOLT__grpc_dial_provider"
var GRPCUsernameAuthParam = "__DOLT__grpc_username"
var NoCachingParameter = "__dolt__NO_CACHING"
var GitCacheRootParam = "git_cache_root"
var GitRemoteNameParam = "git_remote_name"
var GitRefParam = "git_ref"
var AWSRegionParam = "aws-region"
var AWSCredsTypeParam = "aws-creds-type"
var AWSCredsFileParam = "aws-creds-file"
var AWSCredsProfile = "aws-creds-profile"
var AWSCredTypes = []string{"role", "env", "file"}
var OSSCredsFileParam = "oss-creds-file"
var OSSCredsProfile = "oss-creds-profile"

type GRPCRemoteConfig struct {
	Endpoint    string
	DialOptions []grpc.DialOption
	HTTPFetcher grpcendpoint.HTTPFetcher
}

type GRPCDialProvider interface {
	GetGRPCDialParams(grpcendpoint.Config) (GRPCRemoteConfig, error)
}

type DoltRemoteFactory struct {
	insecure bool
}

type GitCacheRootProvider interface {
	GitCacheRoot() (string, bool)
}

func NewDoltRemoteFactory(insecure bool) DoltRemoteFactory {
	return DoltRemoteFactory{insecure: insecure}
}

func (fact DoltRemoteFactory) PrepareDB(ctx context.Context, nbf *types.NomsBinFormat, urlObj *url.URL, params map[string]interface{}) error {
	return fmt.Errorf("remote Dolt database schemes are not available in wasm")
}

func (fact DoltRemoteFactory) CreateDB(ctx context.Context, nbf *types.NomsBinFormat, urlObj *url.URL, params map[string]interface{}) (datas.Database, types.ValueReadWriter, tree.NodeStore, error) {
	return nil, nil, nil, fmt.Errorf("remote Dolt database schemes are not available in wasm")
}

func CreateDB(ctx context.Context, nbf *types.NomsBinFormat, urlStr string, params map[string]interface{}) (datas.Database, types.ValueReadWriter, tree.NodeStore, error) {
	urlObj, err := earl.Parse(urlStr)
	if err != nil {
		return nil, nil, nil, err
	}
	scheme := urlObj.Scheme
	if scheme == "" {
		scheme = defaultScheme
	}
	if fact, ok := DBFactories[strings.ToLower(scheme)]; ok {
		return fact.CreateDB(ctx, nbf, urlObj, params)
	}
	return nil, nil, nil, fmt.Errorf("unsupported wasm database url scheme: %q", urlObj.Scheme)
}

func PrepareDB(ctx context.Context, nbf *types.NomsBinFormat, urlStr string, params map[string]interface{}) error {
	urlObj, err := earl.Parse(urlStr)
	if err != nil {
		return err
	}
	scheme := urlObj.Scheme
	if scheme == "" {
		scheme = defaultScheme
	}
	if fact, ok := DBFactories[strings.ToLower(scheme)]; ok {
		return fact.PrepareDB(ctx, nbf, urlObj, params)
	}
	return fmt.Errorf("unsupported wasm database url scheme: %q", urlObj.Scheme)
}
