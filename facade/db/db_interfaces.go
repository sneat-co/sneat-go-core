package db

// Record is a db record interface
type Record interface {
	Kind() string
	Key() RecordKey
	Data() interface{}
	Validate() error
}

// RecordRef hold a reference to a single record within a root or nested recordset.
type RecordRef struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

// RecordKey represents a full path to a given record (1 item in case of root recordset)
type RecordKey = []RecordRef

// NewRecordKey creates a new record key from a sequence of record's references
func NewRecordKey(refs ...RecordRef) RecordKey {
	return refs
}
