// Copyright 2026 Sneat Co.
package facade

import (
	"errors"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/firestorepolicy"
)

var protectRootCollections = firestorepolicy.ProtectRootCollections

// ProtectSneatDB enforces the reviewed collection registry on Firestore handles.
// Non-Firestore deployments retain their own storage layout. Calling this on a
// separately constructed production handle is required before handing it out.
func ProtectSneatDB(db dal.DB) (dal.DB, error) {
	if db == nil {
		return nil, errors.New("sneat database provider returned nil database")
	}
	if db.Adapter().Name() != "firestore" {
		return db, nil
	}
	return protectRootCollections(db)
}
