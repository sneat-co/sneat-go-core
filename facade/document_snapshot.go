package facade

// DocumentSnapshot fake
type DocumentSnapshot interface {
	Exists() bool
	DataTo(p interface{}) error
}

type documentSnapshot struct {
	exists bool
	dataTo func(dst interface{}) error
}

func (ds *documentSnapshot) Exists() bool {
	return ds.exists
}

func (ds *documentSnapshot) DataTo(p interface{}) error {
	return ds.dataTo(p)
}

// NewDocumentSnapshot creates fake
func NewDocumentSnapshot(exists bool, dataTo func(dst interface{}) error) DocumentSnapshot {
	return &documentSnapshot{exists: exists, dataTo: dataTo}
}
