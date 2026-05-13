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

package nbs

import (
	"fmt"
	"sync"

	kpzstd "github.com/klauspost/compress/zstd"
)

type zstdCDict struct {
	enc *kpzstd.Encoder
	mu  sync.Mutex
}

type zstdDDict struct {
	dec *kpzstd.Decoder
	mu  sync.Mutex
}

var zstdEncoderPool = sync.Pool{
	New: func() any {
		enc, err := kpzstd.NewWriter(
			nil,
			kpzstd.WithEncoderConcurrency(1),
			kpzstd.WithEncoderLevel(kpzstd.SpeedBestCompression),
		)
		if err != nil {
			panic(err)
		}
		return enc
	},
}

var zstdDecoderPool = sync.Pool{
	New: func() any {
		dec, err := kpzstd.NewReader(nil, kpzstd.WithDecoderConcurrency(1))
		if err != nil {
			panic(err)
		}
		return dec
	},
}

func newZstdCDict(dict []byte) (*zstdCDict, error) {
	if len(dict) == 0 {
		return nil, fmt.Errorf("dict cannot be empty")
	}
	enc, err := kpzstd.NewWriter(
		nil,
		kpzstd.WithEncoderConcurrency(1),
		kpzstd.WithEncoderLevel(kpzstd.SpeedBestCompression),
		kpzstd.WithEncoderDict(dict),
	)
	if err != nil {
		return nil, err
	}
	return &zstdCDict{enc: enc}, nil
}

func newZstdDDict(dict []byte) (*zstdDDict, error) {
	if len(dict) == 0 {
		return nil, fmt.Errorf("dict cannot be empty")
	}
	dec, err := kpzstd.NewReader(
		nil,
		kpzstd.WithDecoderConcurrency(1),
		kpzstd.WithDecoderDicts(dict),
	)
	if err != nil {
		return nil, err
	}
	return &zstdDDict{dec: dec}, nil
}

func zstdCompress(dst, src []byte) []byte {
	if len(src) == 0 {
		return dst
	}
	enc := zstdEncoderPool.Get().(*kpzstd.Encoder)
	defer zstdEncoderPool.Put(enc)
	return enc.EncodeAll(src, dst)
}

func zstdCompressDict(dst, src []byte, dict *zstdCDict) []byte {
	if len(src) == 0 {
		return dst
	}
	if dict == nil {
		return zstdCompress(dst, src)
	}
	dict.mu.Lock()
	defer dict.mu.Unlock()
	return dict.enc.EncodeAll(src, dst)
}

func zstdDecompress(dst, src []byte) ([]byte, error) {
	dec := zstdDecoderPool.Get().(*kpzstd.Decoder)
	defer zstdDecoderPool.Put(dec)
	out, err := dec.DecodeAll(src, dst)
	if err != nil {
		return nil, fmt.Errorf("cannot decompress invalid src: %w", err)
	}
	return out, nil
}

func zstdDecompressDict(dst, src []byte, dict *zstdDDict) ([]byte, error) {
	if dict == nil {
		return zstdDecompress(dst, src)
	}
	dict.mu.Lock()
	defer dict.mu.Unlock()
	out, err := dict.dec.DecodeAll(src, dst)
	if err != nil {
		return nil, fmt.Errorf("cannot decompress invalid src: %w", err)
	}
	return out, nil
}

func zstdBuildDict(samples [][]byte, desiredDictLen int) (dict []byte) {
	clean := nonEmptyZstdSamples(samples)
	if len(clean) == 0 {
		return nil
	}
	if desiredDictLen < 8 {
		desiredDictLen = 8
	}
	dict, ok := tryBuildZstdDict(clean, desiredDictLen, false)
	if ok {
		return dict
	}
	dict, ok = tryBuildZstdDict(clean, desiredDictLen, true)
	if ok {
		return dict
	}
	dict, ok = tryBuildZstdDict(fallbackZstdSamples(clean, desiredDictLen), desiredDictLen, true)
	if ok {
		return dict
	}
	return nil
}

func nonEmptyZstdSamples(samples [][]byte) [][]byte {
	clean := make([][]byte, 0, len(samples))
	for _, sample := range samples {
		if len(sample) > 0 {
			clean = append(clean, sample)
		}
	}
	return clean
}

func tryBuildZstdDict(samples [][]byte, desiredDictLen int, repeatFirst bool) (dict []byte, ok bool) {
	defer func() {
		if recover() != nil {
			dict = nil
			ok = false
		}
	}()
	history := zstdDictHistory(samples, desiredDictLen, repeatFirst)
	if len(history) < 8 {
		return nil, false
	}
	dict, err := kpzstd.BuildDict(kpzstd.BuildDictOptions{
		ID:       1,
		Contents: samples,
		History:  history,
		Offsets:  [3]int{1, 4, 8},
		Level:    kpzstd.SpeedBestCompression,
	})
	if err != nil || len(dict) == 0 {
		return nil, false
	}
	if !validZstdDict(dict) {
		return nil, false
	}
	return dict, true
}

func validZstdDict(dict []byte) bool {
	enc, err := kpzstd.NewWriter(nil, kpzstd.WithEncoderConcurrency(1), kpzstd.WithEncoderDict(dict))
	if err != nil {
		return false
	}
	if err := enc.Close(); err != nil {
		return false
	}
	dec, err := kpzstd.NewReader(nil, kpzstd.WithDecoderConcurrency(1), kpzstd.WithDecoderDicts(dict))
	if err != nil {
		return false
	}
	dec.Close()
	return true
}

func zstdDictHistory(samples [][]byte, desiredDictLen int, repeatFirst bool) []byte {
	history := make([]byte, 0, desiredDictLen+len(samples[0]))
	if repeatFirst {
		for len(history) < desiredDictLen {
			history = append(history, samples[0]...)
		}
	} else {
		for _, sample := range samples {
			history = append(history, sample...)
		}
	}
	if len(history) > desiredDictLen {
		history = history[len(history)-desiredDictLen:]
	}
	return history
}

func fallbackZstdSamples(samples [][]byte, desiredDictLen int) [][]byte {
	sample := samples[0]
	if len(sample) > desiredDictLen {
		sample = sample[:desiredDictLen]
	}
	if len(sample) < 8 {
		padded := make([]byte, 8)
		copy(padded, sample)
		for i := len(sample); i < len(padded); i++ {
			padded[i] = byte(i)
		}
		sample = padded
	}
	out := make([][]byte, 32)
	for i := range out {
		row := make([]byte, 0, len(sample)+32)
		row = append(row, sample...)
		row = append(row, "dolt-zstd-fallback-"...)
		row = append(row, byte(i))
		out[i] = row
	}
	return out
}
