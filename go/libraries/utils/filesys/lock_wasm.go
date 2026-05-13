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

package filesys

import (
	"errors"
	"sync/atomic"
)

const unlockedStateValue int32 = 0
const lockedStateValue int32 = 1

var errLockUnlock = errors.New("unable to unlock the lock")

// FilesysLock is an interface for locking and unlocking filesystems.
type FilesysLock interface {
	TryLock() (bool, error)
	Unlock() error
}

// CreateFilesysLock returns an in-process lock for browser builds. Browser
// storage is single-process for the current in-memory target, so OS file locks
// are neither available nor useful.
func CreateFilesysLock(fs Filesys, filename string) FilesysLock {
	return NewInMemFileLock(fs)
}

type InMemFileLock struct {
	state int32
}

func NewInMemFileLock(fs Filesys) *InMemFileLock {
	return &InMemFileLock{unlockedStateValue}
}

func (memLock *InMemFileLock) TryLock() (bool, error) {
	if atomic.CompareAndSwapInt32(&memLock.state, unlockedStateValue, lockedStateValue) {
		return true, nil
	}
	return false, nil
}

func (memLock *InMemFileLock) Unlock() error {
	if memLock.state == 0 {
		return nil
	}
	new := atomic.AddInt32(&memLock.state, -lockedStateValue)
	if new != 0 {
		return errLockUnlock
	}
	return nil
}

type LocalFileLock = InMemFileLock

func NewLocalFileLock(fs Filesys, filename string) *LocalFileLock {
	return NewInMemFileLock(fs)
}
