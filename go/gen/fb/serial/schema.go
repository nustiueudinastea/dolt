// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package serial

import (
	"strconv"

	flatbuffers "github.com/google/flatbuffers/go"
)

type ForeignKeyReferenceOption byte

const (
	ForeignKeyReferenceOptionDefaultAction ForeignKeyReferenceOption = 0
	ForeignKeyReferenceOptionCascade       ForeignKeyReferenceOption = 1
	ForeignKeyReferenceOptionNoAction      ForeignKeyReferenceOption = 2
	ForeignKeyReferenceOptionRestrict      ForeignKeyReferenceOption = 3
	ForeignKeyReferenceOptionSetNull       ForeignKeyReferenceOption = 4
)

var EnumNamesForeignKeyReferenceOption = map[ForeignKeyReferenceOption]string{
	ForeignKeyReferenceOptionDefaultAction: "DefaultAction",
	ForeignKeyReferenceOptionCascade:       "Cascade",
	ForeignKeyReferenceOptionNoAction:      "NoAction",
	ForeignKeyReferenceOptionRestrict:      "Restrict",
	ForeignKeyReferenceOptionSetNull:       "SetNull",
}

var EnumValuesForeignKeyReferenceOption = map[string]ForeignKeyReferenceOption{
	"DefaultAction": ForeignKeyReferenceOptionDefaultAction,
	"Cascade":       ForeignKeyReferenceOptionCascade,
	"NoAction":      ForeignKeyReferenceOptionNoAction,
	"Restrict":      ForeignKeyReferenceOptionRestrict,
	"SetNull":       ForeignKeyReferenceOptionSetNull,
}

func (v ForeignKeyReferenceOption) String() string {
	if s, ok := EnumNamesForeignKeyReferenceOption[v]; ok {
		return s
	}
	return "ForeignKeyReferenceOption(" + strconv.FormatInt(int64(v), 10) + ")"
}

type Column struct {
	_tab flatbuffers.Table
}

func GetRootAsColumn(buf []byte, offset flatbuffers.UOffsetT) *Column {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Column{}
	x.Init(buf, n+offset)
	return x
}

func GetSizePrefixedRootAsColumn(buf []byte, offset flatbuffers.UOffsetT) *Column {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &Column{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func (rcv *Column) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Column) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Column) Name() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Column) StorageOrder() uint16 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetUint16(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Column) MutateStorageOrder(n uint16) bool {
	return rcv._tab.MutateUint16Slot(6, n)
}

func (rcv *Column) SchemaOrder() uint16 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.GetUint16(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Column) MutateSchemaOrder(n uint16) bool {
	return rcv._tab.MutateUint16Slot(8, n)
}

func (rcv *Column) Nullable() bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.GetBool(o + rcv._tab.Pos)
	}
	return false
}

func (rcv *Column) MutateNullable(n bool) bool {
	return rcv._tab.MutateBoolSlot(10, n)
}

func (rcv *Column) PrimaryKey() bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		return rcv._tab.GetBool(o + rcv._tab.Pos)
	}
	return false
}

func (rcv *Column) MutatePrimaryKey(n bool) bool {
	return rcv._tab.MutateBoolSlot(12, n)
}

func (rcv *Column) AutoIncrement() bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		return rcv._tab.GetBool(o + rcv._tab.Pos)
	}
	return false
}

func (rcv *Column) MutateAutoIncrement(n bool) bool {
	return rcv._tab.MutateBoolSlot(14, n)
}

func (rcv *Column) Default(obj *ColumnDefault) *ColumnDefault {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(16))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(ColumnDefault)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func (rcv *Column) Constraints(obj *ColumnConstraint, j int) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(18))
	if o != 0 {
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	}
	return false
}

func (rcv *Column) ConstraintsLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(18))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *Column) Comment() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(20))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func ColumnStart(builder *flatbuffers.Builder) {
	builder.StartObject(9)
}
func ColumnAddName(builder *flatbuffers.Builder, name flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(name), 0)
}
func ColumnAddStorageOrder(builder *flatbuffers.Builder, storageOrder uint16) {
	builder.PrependUint16Slot(1, storageOrder, 0)
}
func ColumnAddSchemaOrder(builder *flatbuffers.Builder, schemaOrder uint16) {
	builder.PrependUint16Slot(2, schemaOrder, 0)
}
func ColumnAddNullable(builder *flatbuffers.Builder, nullable bool) {
	builder.PrependBoolSlot(3, nullable, false)
}
func ColumnAddPrimaryKey(builder *flatbuffers.Builder, primaryKey bool) {
	builder.PrependBoolSlot(4, primaryKey, false)
}
func ColumnAddAutoIncrement(builder *flatbuffers.Builder, autoIncrement bool) {
	builder.PrependBoolSlot(5, autoIncrement, false)
}
func ColumnAddDefault(builder *flatbuffers.Builder, default_ flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(6, flatbuffers.UOffsetT(default_), 0)
}
func ColumnAddConstraints(builder *flatbuffers.Builder, constraints flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(7, flatbuffers.UOffsetT(constraints), 0)
}
func ColumnStartConstraintsVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func ColumnAddComment(builder *flatbuffers.Builder, comment flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(8, flatbuffers.UOffsetT(comment), 0)
}
func ColumnEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}

type ColumnDefault struct {
	_tab flatbuffers.Table
}

func GetRootAsColumnDefault(buf []byte, offset flatbuffers.UOffsetT) *ColumnDefault {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &ColumnDefault{}
	x.Init(buf, n+offset)
	return x
}

func GetSizePrefixedRootAsColumnDefault(buf []byte, offset flatbuffers.UOffsetT) *ColumnDefault {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &ColumnDefault{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func (rcv *ColumnDefault) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *ColumnDefault) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *ColumnDefault) Expression() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func ColumnDefaultStart(builder *flatbuffers.Builder) {
	builder.StartObject(1)
}
func ColumnDefaultAddExpression(builder *flatbuffers.Builder, expression flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(expression), 0)
}
func ColumnDefaultEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}

type ColumnConstraint struct {
	_tab flatbuffers.Table
}

func GetRootAsColumnConstraint(buf []byte, offset flatbuffers.UOffsetT) *ColumnConstraint {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &ColumnConstraint{}
	x.Init(buf, n+offset)
	return x
}

func GetSizePrefixedRootAsColumnConstraint(buf []byte, offset flatbuffers.UOffsetT) *ColumnConstraint {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &ColumnConstraint{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func (rcv *ColumnConstraint) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *ColumnConstraint) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *ColumnConstraint) Name() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *ColumnConstraint) Expression() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *ColumnConstraint) Enforced() bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.GetBool(o + rcv._tab.Pos)
	}
	return false
}

func (rcv *ColumnConstraint) MutateEnforced(n bool) bool {
	return rcv._tab.MutateBoolSlot(8, n)
}

func ColumnConstraintStart(builder *flatbuffers.Builder) {
	builder.StartObject(3)
}
func ColumnConstraintAddName(builder *flatbuffers.Builder, name flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(name), 0)
}
func ColumnConstraintAddExpression(builder *flatbuffers.Builder, expression flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(expression), 0)
}
func ColumnConstraintAddEnforced(builder *flatbuffers.Builder, enforced bool) {
	builder.PrependBoolSlot(2, enforced, false)
}
func ColumnConstraintEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}

type TableSchema struct {
	_tab flatbuffers.Table
}

func GetRootAsTableSchema(buf []byte, offset flatbuffers.UOffsetT) *TableSchema {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &TableSchema{}
	x.Init(buf, n+offset)
	return x
}

func GetSizePrefixedRootAsTableSchema(buf []byte, offset flatbuffers.UOffsetT) *TableSchema {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &TableSchema{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func (rcv *TableSchema) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *TableSchema) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *TableSchema) Columns(obj *Column, j int) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	}
	return false
}

func (rcv *TableSchema) ColumnsLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *TableSchema) Indexes(obj *IndexSchema, j int) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	}
	return false
}

func (rcv *TableSchema) IndexesLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func TableSchemaStart(builder *flatbuffers.Builder) {
	builder.StartObject(2)
}
func TableSchemaAddColumns(builder *flatbuffers.Builder, columns flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(columns), 0)
}
func TableSchemaStartColumnsVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func TableSchemaAddIndexes(builder *flatbuffers.Builder, indexes flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(indexes), 0)
}
func TableSchemaStartIndexesVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func TableSchemaEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}

type IndexSchema struct {
	_tab flatbuffers.Table
}

func GetRootAsIndexSchema(buf []byte, offset flatbuffers.UOffsetT) *IndexSchema {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &IndexSchema{}
	x.Init(buf, n+offset)
	return x
}

func GetSizePrefixedRootAsIndexSchema(buf []byte, offset flatbuffers.UOffsetT) *IndexSchema {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &IndexSchema{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func (rcv *IndexSchema) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *IndexSchema) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *IndexSchema) Name() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *IndexSchema) Columns(j int) []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.ByteVector(a + flatbuffers.UOffsetT(j*4))
	}
	return nil
}

func (rcv *IndexSchema) ColumnsLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *IndexSchema) Unique() bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.GetBool(o + rcv._tab.Pos)
	}
	return false
}

func (rcv *IndexSchema) MutateUnique(n bool) bool {
	return rcv._tab.MutateBoolSlot(8, n)
}

func (rcv *IndexSchema) SystemDefined() bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.GetBool(o + rcv._tab.Pos)
	}
	return false
}

func (rcv *IndexSchema) MutateSystemDefined(n bool) bool {
	return rcv._tab.MutateBoolSlot(10, n)
}

func (rcv *IndexSchema) Comment() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func IndexSchemaStart(builder *flatbuffers.Builder) {
	builder.StartObject(5)
}
func IndexSchemaAddName(builder *flatbuffers.Builder, name flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(name), 0)
}
func IndexSchemaAddColumns(builder *flatbuffers.Builder, columns flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(columns), 0)
}
func IndexSchemaStartColumnsVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func IndexSchemaAddUnique(builder *flatbuffers.Builder, unique bool) {
	builder.PrependBoolSlot(2, unique, false)
}
func IndexSchemaAddSystemDefined(builder *flatbuffers.Builder, systemDefined bool) {
	builder.PrependBoolSlot(3, systemDefined, false)
}
func IndexSchemaAddComment(builder *flatbuffers.Builder, comment flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(4, flatbuffers.UOffsetT(comment), 0)
}
func IndexSchemaEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}

type ForeignKey struct {
	_tab flatbuffers.Table
}

func GetRootAsForeignKey(buf []byte, offset flatbuffers.UOffsetT) *ForeignKey {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &ForeignKey{}
	x.Init(buf, n+offset)
	return x
}

func GetSizePrefixedRootAsForeignKey(buf []byte, offset flatbuffers.UOffsetT) *ForeignKey {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &ForeignKey{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func (rcv *ForeignKey) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *ForeignKey) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *ForeignKey) Name() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *ForeignKey) ChildTable() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *ForeignKey) ChildColumns(j int) []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.ByteVector(a + flatbuffers.UOffsetT(j*4))
	}
	return nil
}

func (rcv *ForeignKey) ChildColumnsLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *ForeignKey) ChildIndex() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *ForeignKey) ParentTable() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *ForeignKey) ParentColumns(j int) []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.ByteVector(a + flatbuffers.UOffsetT(j*4))
	}
	return nil
}

func (rcv *ForeignKey) ParentColumnsLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *ForeignKey) ParentIndex() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(16))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *ForeignKey) OnUpdate() ForeignKeyReferenceOption {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(18))
	if o != 0 {
		return ForeignKeyReferenceOption(rcv._tab.GetByte(o + rcv._tab.Pos))
	}
	return 0
}

func (rcv *ForeignKey) MutateOnUpdate(n ForeignKeyReferenceOption) bool {
	return rcv._tab.MutateByteSlot(18, byte(n))
}

func (rcv *ForeignKey) OnDelete() ForeignKeyReferenceOption {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(20))
	if o != 0 {
		return ForeignKeyReferenceOption(rcv._tab.GetByte(o + rcv._tab.Pos))
	}
	return 0
}

func (rcv *ForeignKey) MutateOnDelete(n ForeignKeyReferenceOption) bool {
	return rcv._tab.MutateByteSlot(20, byte(n))
}

func ForeignKeyStart(builder *flatbuffers.Builder) {
	builder.StartObject(9)
}
func ForeignKeyAddName(builder *flatbuffers.Builder, name flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(name), 0)
}
func ForeignKeyAddChildTable(builder *flatbuffers.Builder, childTable flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(childTable), 0)
}
func ForeignKeyAddChildColumns(builder *flatbuffers.Builder, childColumns flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(childColumns), 0)
}
func ForeignKeyStartChildColumnsVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func ForeignKeyAddChildIndex(builder *flatbuffers.Builder, childIndex flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(3, flatbuffers.UOffsetT(childIndex), 0)
}
func ForeignKeyAddParentTable(builder *flatbuffers.Builder, parentTable flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(4, flatbuffers.UOffsetT(parentTable), 0)
}
func ForeignKeyAddParentColumns(builder *flatbuffers.Builder, parentColumns flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(5, flatbuffers.UOffsetT(parentColumns), 0)
}
func ForeignKeyStartParentColumnsVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func ForeignKeyAddParentIndex(builder *flatbuffers.Builder, parentIndex flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(6, flatbuffers.UOffsetT(parentIndex), 0)
}
func ForeignKeyAddOnUpdate(builder *flatbuffers.Builder, onUpdate ForeignKeyReferenceOption) {
	builder.PrependByteSlot(7, byte(onUpdate), 0)
}
func ForeignKeyAddOnDelete(builder *flatbuffers.Builder, onDelete ForeignKeyReferenceOption) {
	builder.PrependByteSlot(8, byte(onDelete), 0)
}
func ForeignKeyEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
