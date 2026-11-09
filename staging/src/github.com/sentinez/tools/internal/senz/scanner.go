// Copyright 2026 Duc-Hung Ho.
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

package senz

import (
	"fmt"
	"strings"

	typepb "github.com/sentinez/sentinez/api/proto/sentinez/types/v1"
	"github.com/sentinez/tools/internal/utils"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	_wellKnownPackage protoreflect.FullName = "google.protobuf"

	_protoPackage protogen.GoImportPath = "google.golang.org/protobuf/proto"

	_typepbPackage protogen.GoImportPath = "github.com/sentinez/sentinez/" +
		"api/proto/sentinez/types/v1"
)

// The row constraints of the generated ScanOne and Scan functions, met by
// dbx.Row, dbx.Rows, pgx.Row and pgx.Rows.
const (
	_rowConstraint  = "interface{ Scan(dest ...any) error }"
	_rowsConstraint = `interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}`
)

var (
	_timeTime = protogen.GoIdent{
		GoName: "Time", GoImportPath: "time",
	}
	_timeDuration = protogen.GoIdent{
		GoName: "Duration", GoImportPath: "time",
	}
	_timestampNew = protogen.GoIdent{
		GoName:       "New",
		GoImportPath: "google.golang.org/protobuf/types/known/timestamppb",
	}
	_durationNew = protogen.GoIdent{
		GoName:       "New",
		GoImportPath: "google.golang.org/protobuf/types/known/durationpb",
	}
	_protoUnmarshal = protogen.GoIdent{
		GoName: "Unmarshal", GoImportPath: _protoPackage,
	}
	_protoClone = protogen.GoIdent{
		GoName: "Clone", GoImportPath: _protoPackage,
	}
	_protoMarshal = protogen.GoIdent{
		GoName: "Marshal", GoImportPath: _protoPackage,
	}
	_protojsonMarshal = protogen.GoIdent{
		GoName:       "Marshal",
		GoImportPath: "google.golang.org/protobuf/encoding/protojson",
	}
	_protojsonUnmarshalOptions = protogen.GoIdent{
		GoName:       "UnmarshalOptions",
		GoImportPath: "google.golang.org/protobuf/encoding/protojson",
	}
)

// _scalarGoTypes the Go type of each scalar kind, an enum's is its own.
var _scalarGoTypes = map[protoreflect.Kind]string{
	protoreflect.BoolKind:     "bool",
	protoreflect.Int32Kind:    "int32",
	protoreflect.Sint32Kind:   "int32",
	protoreflect.Sfixed32Kind: "int32",
	protoreflect.Uint32Kind:   "uint32",
	protoreflect.Fixed32Kind:  "uint32",
	protoreflect.Int64Kind:    "int64",
	protoreflect.Sint64Kind:   "int64",
	protoreflect.Sfixed64Kind: "int64",
	protoreflect.Uint64Kind:   "uint64",
	protoreflect.Fixed64Kind:  "uint64",
	protoreflect.FloatKind:    "float32",
	protoreflect.DoubleKind:   "float64",
	protoreflect.StringKind:   "string",
	protoreflect.BytesKind:    "[]byte",
}

// leafKind is how a column is scanned into its field.
type leafKind int

const (
	// leafDirect scans the column straight into the field.
	leafDirect leafKind = iota + 1
	// leafTimestamp scans a time.Time, set as a timestamppb.Timestamp.
	leafTimestamp
	// leafDuration scans a time.Duration, set as a durationpb.Duration.
	leafDuration
	// leafProto scans the bytes of a message, unmarshaled into the field.
	leafProto
	// leafProtoJSON scans the protobuf JSON of a message, unmarshaled
	// into the field.
	leafProtoJSON
	// leafNullable scans a pointer to the scalar, set unless NULL.
	leafNullable
)

// scanColumn is a column of a message with the scanner option.
type scanColumn struct {
	column
	kind leafKind
	// local is the variable the column is scanned into, unless leafDirect.
	local string
}

// parentNode is a flattened message field, allocated before the scan so
// that the columns of its fields can be scanned into it.
type parentNode struct {
	field    *protogen.Field
	children []*parentNode
}

// generateScanners writes, for each message with (x_message).scanner, the
// order of its columns and the functions scanning a row of them.
func generateScanners(
	gen *protogen.Plugin, g *protogen.GeneratedFile, file *protogen.File,
) {
	for _, msg := range file.Messages {
		xmsg := getXMessage(msg)
		if xmsg == nil || !xmsg.GetScanner() {
			continue
		}

		cols, err := scanColumns(msg)
		if err != nil {
			gen.Error(err)
			return
		}

		writeColumnsFunc(g, msg, cols)
		writeValuesFunc(g, msg, cols)
		writeScanOne(g, msg, cols)
		writeScan(g, msg)
	}
}

// scanColumns lists the stored columns of msg, every nested message
// flattened unless given a codec. A oneof member can't be scanned into,
// so it is an error.
func scanColumns(msg *protogen.Message) ([]scanColumn, error) {
	cols := columns(msg, true /* flattenAll */)

	scanCols := make([]scanColumn, 0, len(cols))
	for _, col := range cols {
		if getXField(col.leaf()).GetIgnore() {
			continue
		}

		sc, err := newScanColumn(col)
		if err != nil {
			return nil, fmt.Errorf("%s: scanner: %w",
				msg.Desc.FullName(), err)
		}
		scanCols = append(scanCols, sc)
	}
	return scanCols, nil
}

func newScanColumn(col column) (scanColumn, error) {
	if err := checkOneof(col); err != nil {
		return scanColumn{}, err
	}

	kind, err := leafKindOf(col.leaf())
	if err != nil {
		return scanColumn{}, err
	}

	sc := scanColumn{column: col, kind: kind}
	if sc.kind != leafDirect {
		sc.local = "col" + utils.SnakeToPascal(col.name)
	}
	return sc, nil
}

func checkOneof(col column) error {
	for _, field := range col.path {
		if field.Oneof != nil && !field.Oneof.Desc.IsSynthetic() {
			return fmt.Errorf("oneof field %s is not supported",
				field.Desc.FullName())
		}
	}
	return nil
}

func leafKindOf(field *protogen.Field) (leafKind, error) {
	xfield := getXField(field)
	if xfield.GetNullable() {
		if !isNullableScalar(field) {
			return 0, fmt.Errorf("nullable field %s is not a scalar",
				field.Desc.FullName())
		}
		return leafNullable, nil
	}

	if !isMessage(field) {
		if xfield.GetCodec() != typepb.ColumnCodec_COLUMN_CODEC_UNSPECIFIED {
			return 0, fmt.Errorf("codec field %s is not a message",
				field.Desc.FullName())
		}
		return leafDirect, nil
	}

	return messageLeafKind(field, xfield.GetCodec()), nil
}

// isNullableScalar reports whether field is a singular scalar, not a
// proto3 optional one, whose pointer already holds NULL.
func isNullableScalar(field *protogen.Field) bool {
	_, ok := _scalarGoTypes[field.Desc.Kind()]
	return (ok || field.Enum != nil) && !field.Desc.IsList() &&
		!field.Desc.HasPresence()
}

func messageLeafKind(
	field *protogen.Field, codec typepb.ColumnCodec,
) leafKind {
	switch codec {
	case typepb.ColumnCodec_COLUMN_CODEC_PROTOJSON:
		return leafProtoJSON
	case typepb.ColumnCodec_COLUMN_CODEC_DRIVER:
		return leafDirect
	case typepb.ColumnCodec_COLUMN_CODEC_PROTO:
		return leafProto
	}

	switch field.Message.Desc.FullName() {
	case "google.protobuf.Timestamp":
		return leafTimestamp
	case "google.protobuf.Duration":
		return leafDuration
	default:
		return leafProto
	}
}

func (c column) leaf() *protogen.Field {
	return c.path[len(c.path)-1]
}

// goPath is the Go selector of the field of c in the message m.
func (c column) goPath() string {
	return c.selector(".", "")
}

// selector joins m and the Go names of the fields of c, each between
// prefix and suffix.
func (c column) selector(prefix, suffix string) string {
	var b strings.Builder
	b.WriteString("m")
	for _, field := range c.path {
		b.WriteString(prefix)
		b.WriteString(field.GoName)
		b.WriteString(suffix)
	}
	return b.String()
}

func writeColumnsFunc(
	g *protogen.GeneratedFile, msg *protogen.Message, cols []scanColumn,
) {
	name := msg.GoIdent.GoName

	g.P("// ", name, "Columns returns the columns of ", name,
		", in the order ScanOne", name, " scans them.")
	g.P("func ", name, "Columns() []string {")
	g.P("	return []string{")
	for _, col := range cols {
		g.P("		", columnConstName(name, col.name), ",")
	}
	g.P("	}")
	g.P("}")
	g.P()
}

// getterPath is the getter chain of the field of c in the message m, nil
// safe; a proto3 optional scalar of m itself is its pointer, so that an
// unset one is stored as NULL.
func (c column) getterPath() string {
	leaf := c.leaf()
	if len(c.path) == 1 && leaf.Message == nil && leaf.Desc.HasPresence() {
		return c.goPath()
	}
	return c.selector(".Get", "()")
}

func writeValuesFunc(
	g *protogen.GeneratedFile, msg *protogen.Message, cols []scanColumn,
) {
	name := msg.GoIdent.GoName

	g.P("// ", name, "Values returns the values of m, in the order of ",
		name, "Columns.")
	g.P("func ", name, "Values(m *", name, ") ([]any, error) {")
	for _, col := range cols {
		writeValueLocal(g, col)
	}
	g.P("	return []any{")
	for _, col := range cols {
		if col.kind == leafDirect || col.kind == leafNullable {
			g.P("		", col.getterPath(), ",")
			continue
		}
		g.P("		", col.local, ",")
	}
	g.P("	}, nil")
	g.P("}")
	g.P()
}

// writeValueLocal sets the local of a column whose value is converted
// from its field, left nil, stored as NULL, when the field is unset.
func writeValueLocal(g *protogen.GeneratedFile, col scanColumn) {
	switch col.kind {
	case leafTimestamp:
		writeValueConvert(g, col, "v.AsTime()")
	case leafDuration:
		writeValueConvert(g, col, "v.AsDuration()")
	case leafProto:
		writeValueMarshal(g, col, g.QualifiedGoIdent(_protoMarshal))
	case leafProtoJSON:
		writeValueMarshal(g, col, g.QualifiedGoIdent(_protojsonMarshal))
	}
}

func writeValueConvert(
	g *protogen.GeneratedFile, col scanColumn, convert string,
) {
	g.P("	var ", col.local, " any")
	g.P("	if v := ", col.getterPath(), "; v != nil {")
	g.P("		", col.local, " = ", convert)
	g.P("	}")
}

func writeValueMarshal(
	g *protogen.GeneratedFile, col scanColumn, marshal string,
) {
	g.P("	var ", col.local, " []byte")
	g.P("	if v := ", col.getterPath(), "; v != nil {")
	g.P("		b, err := ", marshal, "(v)")
	g.P("		if err != nil {")
	g.P("			return nil, err")
	g.P("		}")
	g.P("		", col.local, " = b")
	g.P("	}")
}

func writeScanOne(
	g *protogen.GeneratedFile, msg *protogen.Message, cols []scanColumn,
) {
	name := msg.GoIdent.GoName

	g.P("// ScanOne", name, " scans row, holding the columns of ",
		name, "Columns, into a ", name, ".")
	g.P("func ScanOne", name, "[R ", _rowConstraint, "](row R) (*",
		name, ", error) {")
	g.P("	m := ", messageLiteral(g, msg.GoIdent, parentNodes(cols)))
	writeLocals(g, cols)
	g.P("	if err := row.Scan(")
	for _, col := range cols {
		if col.kind == leafDirect {
			g.P("		&", col.goPath(), ",")
			continue
		}
		g.P("		&", col.local, ",")
	}
	g.P("	); err != nil {")
	g.P("		return nil, err")
	g.P("	}")
	writeAssigns(g, cols)
	g.P("	return m, nil")
	g.P("}")
	g.P()
}

func writeLocals(g *protogen.GeneratedFile, cols []scanColumn) {
	var locals []scanColumn
	for _, col := range cols {
		if col.kind != leafDirect {
			locals = append(locals, col)
		}
	}
	if len(locals) == 0 {
		return
	}

	g.P("	var (")
	for _, col := range locals {
		g.P("		", col.local, " ", localType(g, col))
	}
	g.P("	)")
}

func localType(g *protogen.GeneratedFile, col scanColumn) string {
	switch col.kind {
	case leafTimestamp:
		return g.QualifiedGoIdent(_timeTime)
	case leafDuration:
		return g.QualifiedGoIdent(_timeDuration)
	case leafNullable:
		return "*" + scalarGoType(g, col.leaf())
	default:
		return "[]byte"
	}
}

func scalarGoType(g *protogen.GeneratedFile, field *protogen.Field) string {
	if field.Enum != nil {
		return g.QualifiedGoIdent(field.Enum.GoIdent)
	}
	return _scalarGoTypes[field.Desc.Kind()]
}

// writeAssigns sets the fields of the columns scanned into a local.
func writeAssigns(g *protogen.GeneratedFile, cols []scanColumn) {
	for _, col := range cols {
		switch col.kind {
		case leafTimestamp, leafDuration:
			newFn := _timestampNew
			if col.kind == leafDuration {
				newFn = _durationNew
			}
			g.P("	", col.goPath(), " = ",
				g.QualifiedGoIdent(newFn), "(", col.local, ")")
		case leafProto:
			writeUnmarshal(g, col, g.QualifiedGoIdent(_protoUnmarshal))
		case leafProtoJSON:
			writeUnmarshal(g, col, "("+g.QualifiedGoIdent(
				_protojsonUnmarshalOptions)+
				"{DiscardUnknown: true}).Unmarshal")
		case leafNullable:
			g.P("	if ", col.local, " != nil {")
			g.P("		", col.goPath(), " = *", col.local)
			g.P("	}")
		}
	}
}

// writeUnmarshal sets the message field of col from its scanned bytes
// with unmarshal, left nil when there are none.
func writeUnmarshal(
	g *protogen.GeneratedFile, col scanColumn, unmarshal string,
) {
	g.P("	if len(", col.local, ") != 0 {")
	g.P("		", col.goPath(), " = &",
		g.QualifiedGoIdent(col.leaf().Message.GoIdent), "{}")
	g.P("		if err := ", unmarshal, "(",
		col.local, ", ", col.goPath(), "); err != nil {")
	g.P("			return nil, err")
	g.P("		}")
	g.P("	}")
}

func writeScan(g *protogen.GeneratedFile, msg *protogen.Message) {
	name := msg.GoIdent.GoName

	g.P("// Scan", name, " scans each row of rows with ScanOne", name, ".")
	g.P("func Scan", name, "[R ", _rowsConstraint, "](rows R) ([]*",
		name, ", error) {")
	g.P("	var ms []*", name)
	g.P("	for rows.Next() {")
	g.P("		m, err := ScanOne", name, "(rows)")
	g.P("		if err != nil {")
	g.P("			return nil, err")
	g.P("		}")
	g.P("		ms = append(ms, m)")
	g.P("	}")
	g.P("	return ms, rows.Err()")
	g.P("}")
	g.P()
}

// parentNodes is the tree of the message fields the columns are flattened
// from, in field order.
func parentNodes(cols []scanColumn) []*parentNode {
	var roots []*parentNode
	for _, col := range cols {
		nodes := &roots
		for _, field := range col.path[:len(col.path)-1] {
			nodes = &childNode(nodes, field).children
		}
	}
	return roots
}

// childNode returns the node of field in nodes, appended if missing.
func childNode(nodes *[]*parentNode, field *protogen.Field) *parentNode {
	for _, node := range *nodes {
		if node.field == field {
			return node
		}
	}

	node := &parentNode{field: field}
	*nodes = append(*nodes, node)
	return node
}

// messageLiteral is the composite literal of a message of ident, with the
// message fields of nodes allocated.
func messageLiteral(
	g *protogen.GeneratedFile, ident protogen.GoIdent, nodes []*parentNode,
) string {
	var b strings.Builder
	b.WriteString("&" + g.QualifiedGoIdent(ident) + "{")
	for i, node := range nodes {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(node.field.GoName + ": ")
		b.WriteString(
			messageLiteral(g, node.field.Message.GoIdent, node.children))
	}
	b.WriteString("}")
	return b.String()
}
