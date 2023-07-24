package db

import "testing"

func TestDbWrappersAreSet(t *testing.T) {
	t.Run("RunTransaction", func(t *testing.T) {
		if RunTransaction == nil {
			t.Fatal("RunTransaction is null")
		}
	})
	t.Run("TxGetAll", func(t *testing.T) {
		if TxGetAll == nil {
			t.Fatal("TxGetAll is null")
		}
	})
	t.Run("TxSet", func(t *testing.T) {
		if TxGetAll == nil {
			t.Fatal("TxSet is null")
		}
	})
	t.Run("TxCreate", func(t *testing.T) {
		if TxGetAll == nil {
			t.Fatal("TxCreate is null")
		}
	})
	t.Run("TxUpdate", func(t *testing.T) {
		if TxGetAll == nil {
			t.Fatal("TxUpdate is null")
		}
	})
	t.Run("Get", func(t *testing.T) {
		if TxGetAll == nil {
			t.Fatal("Get is null")
		}
	})
}
