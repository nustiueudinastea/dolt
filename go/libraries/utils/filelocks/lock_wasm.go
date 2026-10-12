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

package filelocks

import (
	"context"
	"errors"
	"time"
)

// Lock is a no-op in the browser, where the host coordinates persistence.
// Native builds retain upstream's cross-process advisory locks.
type Lock struct{}

var (
	ErrLocked  = errors.New("file is locked")
	ErrTimeout = errors.New("timed out waiting for the lock")
)

func New(path string) (*Lock, error)                    { return &Lock{}, nil }
func (*Lock) Lock() error                               { return nil }
func (*Lock) TryLock() error                            { return nil }
func (*Lock) Unlock() error                             { return nil }
func (*Lock) Close() error                              { return nil }
func (*Lock) LockWithTimeout(time.Duration) error       { return nil }
func (*Lock) LockWithContext(ctx context.Context) error { return ctx.Err() }
