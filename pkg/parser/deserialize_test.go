package parser

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeserializeInt(t *testing.T) {
	tests := map[string]struct {
		input     []byte
		exptected int64
		valid     bool
	}{
		"invalid_nil": {
			input: nil,
		},
		"invalid_empty": {
			input: []byte(""),
		},
		"invalid_sybmol": {
			input: []byte("*1234\r\n"),
		},
		"invalid_number": {
			input: []byte(":abcd\r\n"),
		},
		"invalid_large_number": {
			input: []byte(":9223372036854775808\r\n"),
		},
		"invalid_small_number": {
			input: []byte(":-9223372036854775808\r\n"),
		},
		"invalid_number_without_delimiter": {
			input: []byte(":1234"),
		},
		"valid_large_positive_number": {
			input:     []byte(":9223372036854775807\r\n"),
			exptected: 9223372036854775807,
			valid:     true,
		},
		"valid_small_negative_number": {
			input:     []byte(":-9223372036854775807\r\n"),
			exptected: -9223372036854775807,
			valid:     true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			buf := bufio.NewReader(bytes.NewReader(tc.input))
			val, err := deserializeInt(buf)
			if !tc.valid {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.exptected, val.val)
		})
	}
}

func TestDeserializeString(t *testing.T) {
	tests := map[string]struct {
		input     []byte
		exptected string
		valid     bool
	}{
		"invalid_nil": {
			input: nil,
		},
		"invalid_empty": {
			input: []byte(""),
		},
		"invalid_sybmol": {
			input: []byte(":simple string\r\n"),
		},
		"invalid_string_with_carriage_return": {
			input: []byte("+ab\rcd\r\n"),
		},
		"invalid_string_with_new_line": {
			input: []byte("+ab\ncd\r\n"),
		},
		"invalid_string_without_delimiter": {
			input: []byte("+simple string"),
		},
		"valid_string": {
			input:     []byte("+hello world\r\n"),
			exptected: "hello world",
			valid:     true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			buf := bufio.NewReader(bytes.NewReader(tc.input))
			val, err := deserializeString(buf)
			if !tc.valid {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.exptected, val.val)
		})
	}
}

func TestDeserializeBulkString(t *testing.T) {
	tests := map[string]struct {
		input     []byte
		exptected []byte
		valid     bool
	}{
		"invalid_nil": {
			input: nil,
		},
		"invalid_empty": {
			input: []byte(""),
		},
		"invalid_sybmol": {
			input: []byte(":11\r\nbulk string\r\n"),
		},
		"invalid_missing_length": {
			input: []byte("$bulk string\r\n"),
		},
		"invalid_incorrect_length_smaller_than_actual": {
			input: []byte("$5\r\nbulk string\r\n"),
		},
		"invalid_incorrect_length_larger_than_actual": {
			input: []byte("$12\r\nbulk string\r\n"),
		},
		"invalid_length_without_delimiter": {
			input: []byte("$12bulk string\r\n"),
		},
		"invalid_bulk_string_with_invalid_delimiter_without_new_line": {
			input: []byte("$11\r\nbulk string\r"),
		},
		"invalid_bulk_string_with_invalid_delimiter_without_carriage_return": {
			input: []byte("$11\r\nbulk string\n"),
		},
		"invalid_string_without_delimiter": {
			input: []byte("$11\r\nbulk string"),
		},
		"valid_string": {
			input:     []byte("$12\r\nabcd\r\nab\ncd\r\r\n"),
			exptected: []byte("abcd\r\nab\ncd\r"),
			valid:     true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			buf := bufio.NewReader(bytes.NewReader(tc.input))
			val, err := deserializeBulkString(buf)
			if !tc.valid {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.exptected, val.val)
		})
	}
}

func TestDeserializeArray(t *testing.T) {
	tests := map[string]struct {
		input     []byte
		exptected []Value
		valid     bool
	}{
		"invalid_nil": {
			input: nil,
		},
		"invalid_empty": {
			input: []byte(""),
		},
		"invalid_sybmol": {
			input: []byte("$12\r\nabcd\r\nab\ncd\r\r\n"),
		},
		"invalid_missing_length": {
			input: []byte("*+hello world\r\n+hello back\r\n"),
		},
		"invalid_incorrect_length_smaller_than_actual": {
			input: []byte("*0\r\n+hello world\r\n"),
		},
		"invalid_incorrect_length_larger_than_actual": {
			input: []byte("*2\r\n+hello world\r\n"),
		},
		"invalid_length_without_delimiter": {
			input: []byte("$1+hello world\r\n"),
		},
		"invalid_array_element_delimiter_without_newline": {
			input: []byte("$1\r\n+hello world\r"),
		},
		"invalid_bulk_string_with_invalid_delimiter_without_carriage_return": {
			input: []byte("$1\r\n+hello world\n"),
		},
		"invalid_string_without_delimiter": {
			input: []byte("$1\r\n+hello world"),
		},
		"valid_arrray_multiple_elements": {
			input: []byte("*5\r\n+hello world\r\n-Error\r\n$5\r\nabcd\r\r\n:9223372036854775807\r\n"),
			exptected: []Value{
				&strValue{
					val: "hello world",
				},
				&errorValue{
					val: []byte("Error"),
				},
				&bulkValue{
					val: []byte("abcd\r"),
				},
				&intValue{
					val: 9223372036854775807,
				},
			},
			valid: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			buf := bufio.NewReader(bytes.NewReader(tc.input))
			val, err := deserializeArray(buf)
			if !tc.valid {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.exptected, val.val)
		})
	}
}
