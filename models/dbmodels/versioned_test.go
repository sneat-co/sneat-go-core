package dbmodels

import (
	"github.com/sneat-co/sneat-go-core/models/dbmodels/with"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestWithUpdatedAndVersion_IncreaseVersion(t *testing.T) {
	v := &WithUpdatedAndVersion{
		UpdatedFields: with.UpdatedFields{
			UpdatedAt: time.Now(),
			UpdatedBy: "test1",
		},
		WithVersion: WithVersion{
			Version: 1,
		},
	}
	now := time.Now()
	version := v.IncreaseVersion(now, "test2")
	assert.Equal(t, 2, version)
	assert.Equal(t, now, v.UpdatedAt)
	assert.Equal(t, "test2", v.UpdatedBy)
}
