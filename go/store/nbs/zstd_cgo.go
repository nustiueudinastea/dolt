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

//go:build !dolt_purego_zstd

package nbs

import "github.com/dolthub/gozstd"

type zstdCDict = gozstd.CDict
type zstdDDict = gozstd.DDict

func newZstdCDict(dict []byte) (*zstdCDict, error) {
	return gozstd.NewCDict(dict)
}

func newZstdDDict(dict []byte) (*zstdDDict, error) {
	return gozstd.NewDDict(dict)
}

func zstdCompress(dst, src []byte) []byte {
	return gozstd.Compress(dst, src)
}

func zstdCompressDict(dst, src []byte, dict *zstdCDict) []byte {
	return gozstd.CompressDict(dst, src, dict)
}

func zstdDecompress(dst, src []byte) ([]byte, error) {
	return gozstd.Decompress(dst, src)
}

func zstdDecompressDict(dst, src []byte, dict *zstdDDict) ([]byte, error) {
	return gozstd.DecompressDict(dst, src, dict)
}

func zstdBuildDict(samples [][]byte, desiredDictLen int) []byte {
	return gozstd.BuildDict(samples, desiredDictLen)
}
