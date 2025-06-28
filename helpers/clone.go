package helpers

import (
	"bytes"
	"encoding/gob"
)

type Pointer[T any] interface {
	*T
}

func Clone[T any](r *T) *T {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)

	err := enc.Encode(r)
	if err != nil {
		panic(err)
	}

	var result T
	err = dec.Decode(&result)
	if err != nil {
		panic(err)
	}

	return &result
}

type Cloneable[T any] interface {
	Clone() *T
}
