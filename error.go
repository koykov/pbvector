package pbvector

import "errors"

var (
	ErrVarintOverflow       = errors.New("pbvector: varint overflow")
	ErrUnexpectedWireType   = errors.New("pbvector: unexpected wire type")
	ErrTruncated            = errors.New("pbvector: truncated source")
	ErrRecursionLimit       = errors.New("pbvector: recursion limit exceeded")
	ErrIncompatibleWireType = errors.New("pbvector: incompatible wire type")
)
