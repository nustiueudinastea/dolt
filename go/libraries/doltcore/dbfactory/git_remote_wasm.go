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
)

func TeardownGitRemotes(ctx context.Context) {}

// GitRemoteCacheDirName matches the native cache location used by shared env code.
const GitRemoteCacheDirName = "git-remote-cache"

// Browser builds have no native git remote stores to close.
func CloseGitRemotesUnderRoot(root string) error { return nil }

// GitRemoteCacheKey is unavailable with the browser's memory-only factories.
func GitRemoteCacheKey(urlObj *url.URL, params map[string]interface{}) (string, error) {
	return "", fmt.Errorf("git remote database schemes are not available in wasm")
}
