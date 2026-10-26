package slugs

import (
	"context"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/firestorepolicy"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNamespacePlacementAndIsolation(t *testing.T) {
	scopes := []struct {
		namespace Namespace
		path      string
	}{
		{Namespace{ExtensionID: "schoolus", Kind: "space"}, "ext/schoolus/slugNamespaces/space/claims/hall"},
		{Namespace{ExtensionID: "communitycentrum", Kind: "space"}, "ext/communitycentrum/slugNamespaces/space/claims/hall"},
		{Namespace{ExtensionID: "bookius", Kind: "bookingType", SpaceID: "space-1"}, "spaces/space-1/ext/bookius/slugNamespaces/bookingType/claims/hall"},
		{Namespace{ExtensionID: "bookius", Kind: "bookingType", SpaceID: "space2"}, "spaces/space2/ext/bookius/slugNamespaces/bookingType/claims/hall"},
		{Namespace{ExtensionID: "bookius", Kind: "bookingType"}, "ext/bookius/slugNamespaces/bookingType/claims/hall"},
	}
	db, err := firestorepolicy.ProtectRootCollections(sneatcoretesting.NewMemoryDB())
	require.NoError(t, err)
	ctx := context.Background()
	for _, scope := range scopes {
		require.Equal(t, scope.path, claimKey(scope.namespace, "hall").String())
		err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			_, e := Claim(ctx, tx, scope.namespace, "hall", scope.path, "space")
			return e
		})
		require.NoError(t, err)
	}
	for _, scope := range scopes {
		info, e := Resolve(ctx, db, scope.namespace, "hall")
		require.NoError(t, e)
		require.Equal(t, scope.path, info.TargetID)
		rows, e := Enumerate(ctx, db, scope.namespace)
		require.NoError(t, e)
		require.Len(t, rows, 1)
	}
	err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, e := Claim(ctx, tx, Namespace{ExtensionID: "unregistered", Kind: "space"}, "hall", "target", "space")
		return e
	})
	require.Error(t, err)
}
