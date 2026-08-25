package application

import "errors"

var ErrIdempotencyKey = errors.New("idempotency key required")
