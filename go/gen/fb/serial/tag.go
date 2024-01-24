// Copyright 2022-2023 Dolthub, Inc.
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

// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package serial

import (
	flatbuffers "github.com/dolthub/flatbuffers/v23/go"
)

type Tag struct {
	_tab flatbuffers.Table
}

func InitTagRoot(o *Tag, buf []byte, offset flatbuffers.UOffsetT) error {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	return o.Init(buf, n+offset)
}

func TryGetRootAsTag(buf []byte, offset flatbuffers.UOffsetT) (*Tag, error) {
	x := &Tag{}
	return x, InitTagRoot(x, buf, offset)
}

func TryGetSizePrefixedRootAsTag(buf []byte, offset flatbuffers.UOffsetT) (*Tag, error) {
	x := &Tag{}
	return x, InitTagRoot(x, buf, offset+flatbuffers.SizeUint32)
}

func (rcv *Tag) Init(buf []byte, i flatbuffers.UOffsetT) error {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
	if TagNumFields < rcv.Table().NumFields() {
		return flatbuffers.ErrTableHasUnknownFields
	}
	return nil
}

func (rcv *Tag) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Tag) CommitAddr(j int) byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.GetByte(a + flatbuffers.UOffsetT(j*1))
	}
	return 0
}

func (rcv *Tag) CommitAddrLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *Tag) CommitAddrBytes() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Tag) MutateCommitAddr(j int, n byte) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.MutateByte(a+flatbuffers.UOffsetT(j*1), n)
	}
	return false
}

func (rcv *Tag) Name() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Tag) Email() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Tag) Desc() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Tag) TimestampMillis() uint64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		return rcv._tab.GetUint64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Tag) MutateTimestampMillis(n uint64) bool {
	return rcv._tab.MutateUint64Slot(12, n)
}

func (rcv *Tag) UserTimestampMillis() int64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		return rcv._tab.GetInt64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Tag) MutateUserTimestampMillis(n int64) bool {
	return rcv._tab.MutateInt64Slot(14, n)
}

const TagNumFields = 6

func TagStart(builder *flatbuffers.Builder) {
	builder.StartObject(TagNumFields)
}
func TagAddCommitAddr(builder *flatbuffers.Builder, commitAddr flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(commitAddr), 0)
}
func TagStartCommitAddrVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(1, numElems, 1)
}
func TagAddName(builder *flatbuffers.Builder, name flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(name), 0)
}
func TagAddEmail(builder *flatbuffers.Builder, email flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(email), 0)
}
func TagAddDesc(builder *flatbuffers.Builder, desc flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(3, flatbuffers.UOffsetT(desc), 0)
}
func TagAddTimestampMillis(builder *flatbuffers.Builder, timestampMillis uint64) {
	builder.PrependUint64Slot(4, timestampMillis, 0)
}
func TagAddUserTimestampMillis(builder *flatbuffers.Builder, userTimestampMillis int64) {
	builder.PrependInt64Slot(5, userTimestampMillis, 0)
}
func TagEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
