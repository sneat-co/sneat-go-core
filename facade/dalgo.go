package facade

import (
	"context"
	"github.com/dal-go/dalgo/dal"
)

// GetDatabase creates a new DB for a given context
var GetDatabase func(ctx context.Context) dal.Database
