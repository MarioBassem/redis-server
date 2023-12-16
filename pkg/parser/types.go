package parser

import (
	"fmt"
)

const (
	Int = iota
	String
	Error
	Bulk
	Array
)

type Value interface {
	serialize() []byte
	getType() int
}

type intValue struct {
	val int64
}

func (i *intValue) serialize() []byte {
	b := append([]byte(":"), []byte(fmt.Sprintf("%d", i.val))...)

	return append(b, []byte("\r\n")...)
}

func (i *intValue) getType() int {
	return Int
}

type strValue struct {
	val string
}

func (i *strValue) serialize() []byte {
	b := append([]byte("+"), i.val...)

	return append(b, []byte("\r\n")...)
}

func (i *strValue) getType() int {
	return String
}

type errorValue struct {
	val []byte
}

func (i *errorValue) serialize() []byte {
	b := append([]byte("-"), i.val...)

	return append(b, []byte("\r\n")...)
}

func (i *errorValue) getType() int {
	return Error
}

type bulkValue struct {
	val []byte
}

func (i *bulkValue) serialize() []byte {
	valLen := fmt.Sprintf("%d", len(i.val))
	b := append([]byte("$"), []byte(valLen)...)
	b = append(b, []byte("\r\n")...)

	b = append(b, i.val...)
	b = append(b, []byte("\r\n")...)

	return b
}

func (i *bulkValue) getType() int {
	return Bulk
}

type arrayValue struct {
	val []Value
}

func (i *arrayValue) serialize() []byte {
	valLen := fmt.Sprintf("%d", len(i.val))
	b := append([]byte("*"), []byte(valLen)...)
	b = append(b, []byte("\r\n")...)

	for _, v := range i.val {
		b = append(b, v.serialize()...)
	}

	return b
}

func (i *arrayValue) getType() int {
	return Array
}
