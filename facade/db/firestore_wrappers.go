package db

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go/src/core/facade"
)

// RunTransaction a wrapper to run a Firestore transaction
var RunTransaction = func(client *firestore.Client, ctx context.Context, f func(ctx2 context.Context, transaction *firestore.Transaction) error) error {
	return client.RunTransaction(ctx, f)
}

//var txDelete = func(tx *firestore.Transaction, dr *firestore.DocumentRef, opts... firestore.Precondition) error {
//	return tx.Delete(dr, opts...)
//}

// TxGetAll a wrapper to get all records
var TxGetAll = func(tx *firestore.Transaction, drs []*firestore.DocumentRef, records []Record) (docSnapshots []facade.DocumentSnapshot, err error) {
	if drs == nil {
		panic("drs == nil")
	}
	if records != nil {
		if len(drs) != len(records) {
			panic(fmt.Sprintf("len(drs) != len(records): %v, %v", len(drs), len(records)))
		}
		for i, d := range records {
			if d == nil {
				err = fmt.Errorf("TxGetAll: records[%v] == nil", i)
				return
			}
		}
	}
	if len(drs) == 0 {
		return
	}
	ds, err := tx.GetAll(drs)
	docSnapshots = make([]facade.DocumentSnapshot, len(ds))
	// Do not check error here, first copy `ds` to docSnapshots
	for i, v := range ds {
		docSnapshots[i] = v
		if v.Exists() {
			if err = v.DataTo(records[i]); err != nil {
				return
			}
		}
	}
	if err != nil {
		return nil, err
	}
	return
}

// TxGet is a wrapper to call tx.Get()
var TxGet = func(tx *firestore.Transaction, dr *firestore.DocumentRef, record Record, opts ...firestore.SetOption) (docSnapshot *firestore.DocumentSnapshot, err error) {
	docSnapshot, err = tx.Get(dr)
	if err != nil {
		return
	}
	if record != nil {
		err = docSnapshot.DataTo(record)
	}
	return
}

// TxDelete is a wrapper to call tx.Delete()
var TxDelete = func(ctx context.Context, tx dal.ReadwriteTransaction, key *dal.Key, opts ...dal.Precondition) error {
	return tx.Delete(ctx, key)
}

// TxSet is a wrapper to call tx.Set()
var TxSet = func(ctx context.Context, tx dal.ReadwriteTransaction, record dal.Record) error {
	return tx.Set(ctx, record)
}

// TxCreate is a wrapper to call tx.Create()
var TxCreate = func(ctx context.Context, tx dal.ReadwriteTransaction, record dal.Record) error {
	return tx.Insert(ctx, record)
}

// TxUpdate is a wrapper to call tx.Update()
var TxUpdate = func(ctx context.Context, tx dal.ReadwriteTransaction, key *dal.Key, updates []dal.Update, opts ...dal.Precondition) error {
	return tx.Update(ctx, key, updates, opts...)
}

// Get is a wrapper to call Get()
var Get = func(ctx context.Context, ref *firestore.DocumentRef) (*firestore.DocumentSnapshot, error) {
	return ref.Get(ctx)
}

// NewRootCollectionRef creates root collection ref
var NewRootCollectionRef = func(client *firestore.Client, path string) *firestore.CollectionRef {
	return client.Collection(path)
}

// NewSubCollectionRef creates sub collection ref
var NewSubCollectionRef = func(doc *firestore.DocumentRef, path string) *firestore.CollectionRef {
	return doc.Collection(path)
}

// NewDocRef creates doc ref
var NewDocRef = func(collection *firestore.CollectionRef, id string) *firestore.DocumentRef {
	return collection.Doc(id)
}
