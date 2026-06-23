//go:build !js || !wasm

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

package blobstore

import "github.com/dolthub/fslock"

func fLock(lockFilePath string) (localFileLock, error) {
	lck, err := fslock.New(lockFilePath)
	if err != nil {
		return nil, err
	}
	if err := lck.Lock(); err != nil {
		_ = lck.Close()
		return nil, err
	}

	return lck, nil
}
