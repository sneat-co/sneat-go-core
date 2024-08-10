package facade

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
)

// GetDatabase creates a new DB for a given context
var GetDatabase = func(ctx context.Context) (dal.DB, error) {
	return nil, fmt.Errorf("%w: facade.GetDatabase(context.Context) (dal.DB, error)", ErrNotInitialized)
}
