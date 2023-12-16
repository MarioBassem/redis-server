package parser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
)

func deserialize(r io.Reader) (Value, error) {
	buf := bufio.NewReader(r)

	return deserializeValue(buf)
}

func deserializeValue(buf *bufio.Reader) (Value, error) {
	typeByte, err := buf.Peek(1)
	if err != nil {
		return nil, fmt.Errorf("unexpected read: %w", err)
	}

	switch typeByte[0] {
	case ':':
		return deserializeInt(buf)
	case '+':
		return deserializeString(buf)
	case '-':
		return deserializeError(buf)
	case '$':
		return deserializeBulkString(buf)
	case '*':
		return deserializeArray(buf)
	}

	return nil, fmt.Errorf("unsupported symbol type '%c'", typeByte)
}

func deserializeArray(buf *bufio.Reader) (*arrayValue, error) {
	typeByte, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("unexpected error: %w", err)
	}

	if typeByte != '*' {
		return nil, fmt.Errorf("unexpected symbol '%c' for array: '*' was expected", typeByte)
	}

	arrayLenBytes, err := buf.ReadBytes('\r')
	if err != nil {
		return nil, fmt.Errorf("failed to read array length: %w", err)
	}

	arrayLenBytes, err = removeDelimiter(buf, arrayLenBytes)
	if err != nil {
		return nil, err
	}

	arrayLen, err := strconv.ParseUint(string(arrayLenBytes), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse array length: %w", err)
	}

	values := make([]Value, arrayLen)
	for idx := range values {
		val, err := deserializeValue(buf)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize array elements: %w", err)
		}
		values[idx] = val
	}

	return &arrayValue{
		val: values,
	}, nil
}

func deserializeString(buf *bufio.Reader) (*strValue, error) {
	typeByte, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("unexpected error: %w", err)
	}

	if typeByte != '+' {
		return nil, fmt.Errorf("unexpected symbol '%c' for string: '+' was expected", typeByte)
	}

	strBytes, err := buf.ReadBytes('\r')
	if err != nil {
		return nil, fmt.Errorf("failed to read string bytes: %w", err)
	}

	strBytes, err = removeDelimiter(buf, strBytes)
	if err != nil {
		return nil, err
	}

	if bytes.Contains(strBytes, []byte("\n")) {
		return nil, errors.New("invalid sytnax: string must not contains '\\n'")
	}

	return &strValue{
		val: string(strBytes),
	}, nil
}

func deserializeInt(buf *bufio.Reader) (*intValue, error) {
	typeByte, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("unexpected error: %w", err)
	}

	if typeByte != ':' {
		return nil, fmt.Errorf("unexpected symbol '%c' for int: ':' was expected", typeByte)
	}

	intBytes, err := buf.ReadBytes('\r')
	if err != nil {
		return nil, fmt.Errorf("failed to read int bytes: %w", err)
	}

	sign := int64(1)
	if len(intBytes) > 0 && (intBytes[0] == '-' || intBytes[0] == '+') {
		if intBytes[0] == '-' {
			sign = -1
		}

		intBytes = intBytes[1:]
	}

	intBytes, err = removeDelimiter(buf, intBytes)
	if err != nil {
		return nil, err
	}

	val, err := strconv.ParseUint(string(intBytes), 10, 63)
	if err != nil {
		return nil, fmt.Errorf("failed to parse int: %w", err)
	}

	return &intValue{
		val: int64(val) * sign,
	}, nil
}

func deserializeBulkString(buf *bufio.Reader) (*bulkValue, error) {
	typeByte, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("unexpected error: %w", err)
	}

	if typeByte != '$' {
		return nil, fmt.Errorf("unexpected symbol '%c' for bulk string: '$' was expected", typeByte)
	}

	bulkBytesLength, err := getLength(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read bulk string length: %w", err)
	}

	bulkBytes := make([]byte, bulkBytesLength)
	_, err = buf.Read(bulkBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to read buld bytes: %w", err)
	}

	delimiterBytes := make([]byte, 2)
	_, err = buf.Read(delimiterBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to read delimiter bytes: %w", err)
	}

	if delimiterBytes[0] != '\r' || delimiterBytes[1] != '\n' {
		return nil, fmt.Errorf("invalid delimiter '%c%c'", delimiterBytes[0], delimiterBytes[1])
	}

	return &bulkValue{
		val: bulkBytes,
	}, nil
}

func deserializeError(buf *bufio.Reader) (*errorValue, error) {
	typeByte, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("unexpected error: %w", err)
	}

	if typeByte != '-' {
		return nil, fmt.Errorf("unexpected symbol '%c' for error: '-' was expected", typeByte)
	}

	errorBytes, err := buf.ReadBytes('\r')
	if err != nil {
		return nil, fmt.Errorf("failed to read error bytes: %w", err)
	}

	errorBytes, err = removeDelimiter(buf, errorBytes)
	if err != nil {
		return nil, err
	}

	if bytes.Contains(errorBytes, []byte("\n")) {
		return nil, errors.New("invalid sytnax: string must not contains '\\n'")
	}
	return &errorValue{
		val: errorBytes,
	}, nil
}

func removeDelimiter(buf *bufio.Reader, b []byte) ([]byte, error) {
	nl, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("faild to read delimiter bytes: %w", err)
	}

	if nl != '\n' {
		return nil, fmt.Errorf("invalid syntax: '\\n' was expects, but '%c' was found", nl)
	}

	return b[:len(b)-1], nil
}

func getLength(buf *bufio.Reader) (uint32, error) {
	lenBytes, err := buf.ReadBytes('\r')
	if err != nil {
		return 0, fmt.Errorf("failed to read bulk bytes: %w", err)
	}

	lenBytes, err = removeDelimiter(buf, lenBytes)
	if err != nil {
		return 0, err
	}

	length, err := strconv.ParseUint(string(lenBytes), 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse length: %w", err)
	}

	return uint32(length), nil
}
