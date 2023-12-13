package parser

import "io"

type Value interface{}

// arrays can be of different types
func deserializeArray(r io.Reader) ([]Value, error)

func deserializeString(r io.Reader) (string, error)
func deserializeInt(i io.Reader) (int64, error)
func deserializeBulkString(r io.Reader) ([]byte, error)
func deserializeError(r io.Reader)
