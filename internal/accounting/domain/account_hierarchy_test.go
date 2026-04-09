package domain

import (
	"testing"
	"time"

	"github.com/basilex/skeleton/pkg/money"
	"github.com/stretchr/testify/require"
)

// TODO: Implement hierarchy methods (AddChild, SetParent, IsRoot, ClearParent)
// These tests are commented until methods are implemented

/*
func TestAccount_AddChild(t *testing.T) {
	currency := Currency("USD")

	t.Run("add child with same type", func(t *testing.T) {
		parent, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)
		child, _ := NewAccount("1100", "Cash", AccountTypeAsset, currency, nil)

		err := parent.AddChild(child)
		require.NoError(t, err)
		require.NotNil(t, child.GetParentID())
		require.Equal(t, parent.GetID(), *child.GetParentID())
	})

	t.Run("cannot add child with different type", func(t *testing.T) {
		parent, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)
		child, _ := NewAccount("2000", "Liabilities", AccountTypeLiability, currency, nil)

		err := parent.AddChild(child)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrParentTypeMismatch)
	})

	t.Run("cannot add self as child", func(t *testing.T) {
		account, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)

		err := account.AddChild(account)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrCircularReference)
	})

	t.Run("cannot add nil child", func(t *testing.T) {
		parent, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)

		err := parent.AddChild(nil)
		require.Error(t, err)
	})
}

func TestAccount_SetParent(t *testing.T) {
	currency := Currency("USD")

	t.Run("set parent with same type", func(t *testing.T) {
		parent, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)
		child, _ := NewAccount("1100", "Cash", AccountTypeAsset, currency, nil)

		err := child.SetParent(parent)
		require.NoError(t, err)
		require.NotNil(t, child.GetParentID())
		require.Equal(t, parent.GetID(), *child.GetParentID())
	})

	t.Run("set parent to nil makes root", func(t *testing.T) {
		parent, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)
		child, _ := NewAccount("1100", "Cash", AccountTypeAsset, currency, nil)
		_ = child.SetParent(parent)

		err := child.SetParent(nil)
		require.NoError(t, err)
		require.Nil(t, child.GetParentID())
		require.True(t, child.IsRoot())
	})

	t.Run("cannot set parent with different type", func(t *testing.T) {
		parent, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)
		child, _ := NewAccount("2000", "Liabilities", AccountTypeLiability, currency, nil)

		err := child.SetParent(parent)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrParentTypeMismatch)
	})

	t.Run("cannot set self as parent", func(t *testing.T) {
		account, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)

		err := account.SetParent(account)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrCircularReference)
	})

	t.Run("set parent emits event", func(t *testing.T) {
		parent, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)
		child, _ := NewAccount("1100", "Cash", AccountTypeAsset, currency, nil)
		_ = child.PullEvents()

		err := child.SetParent(parent)
		require.NoError(t, err)

		events := child.PullEvents()
		require.Len(t, events, 1)
		_, ok := events[0].(AccountParentChanged)
		require.True(t, ok)
	})
}

func TestAccount_ClearParent(t *testing.T) {
	currency := Currency("USD")

	t.Run("clear parent makes root", func(t *testing.T) {
		parent, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)
		child, _ := NewAccount("1100", "Cash", AccountTypeAsset, currency, nil)
		_ = child.SetParent(parent)

		child.ClearParent()
		require.Nil(t, child.GetParentID())
		require.True(t, child.IsRoot())
	})
}

func TestAccount_IsRoot(t *testing.T) {
	currency := Currency("USD")

	t.Run("account without parent is root", func(t *testing.T) {
		account, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)
		require.True(t, account.IsRoot())
	})

	t.Run("account with parent is not root", func(t *testing.T) {
		parent, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)
		child, _ := NewAccount("1100", "Cash", AccountTypeAsset, currency, nil)
		_ = child.SetParent(parent)

		require.False(t, child.IsRoot())
	})
}

func TestAccount_CanDelete(t *testing.T) {
	currency := Currency("USD")

	t.Run("can delete root with zero balance", func(t *testing.T) {
		account, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)
		err := account.CanDelete()
		require.NoError(t, err)
	})

	t.Run("cannot delete with non-zero balance", func(t *testing.T) {
		account, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)
		account.balance = Money{Amount: 1000, Currency: currency}

		err := account.CanDelete()
		require.Error(t, err)
		require.ErrorIs(t, err, ErrAccountHasBalance)
	})
}

func TestAccount_IsDescendantOf(t *testing.T) {
	currency := Currency("USD")

	t.Run("is descendant of direct parent", func(t *testing.T) {
		root, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)
		parent, _ := NewAccount("1100", "Current Assets", AccountTypeAsset, currency, nil)
		child, _ := NewAccount("1110", "Cash", AccountTypeAsset, currency, nil)

		_ = parent.SetParent(root)
		_ = child.SetParent(parent)

		accounts := map[AccountID]*Account{
			root.GetID():   root,
			parent.GetID(): parent,
			child.GetID():  child,
		}
		getParent := func(id AccountID) (*Account, error) {
			return accounts[id], nil
		}

		isDesc, err := child.IsDescendantOf(root.GetID(), getParent)
		require.NoError(t, err)
		require.True(t, isDesc)

		isDesc, err = child.IsDescendantOf(parent.GetID(), getParent)
		require.NoError(t, err)
		require.True(t, isDesc)

		isDesc, err = child.IsDescendantOf(child.GetID(), getParent)
		require.NoError(t, err)
		require.False(t, isDesc)
	})

	t.Run("root is not descendant of anyone", func(t *testing.T) {
		root, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)
		other, _ := NewAccount("2000", "Liabilities", AccountTypeLiability, currency, nil)

		accounts := map[AccountID]*Account{
			root.GetID():  root,
			other.GetID(): other,
		}
		getParent := func(id AccountID) (*Account, error) {
			return accounts[id], nil
		}

		isDesc, err := root.IsDescendantOf(other.GetID(), getParent)
		require.NoError(t, err)
		require.False(t, isDesc)
	})
}

func TestAccount_Depth(t *testing.T) {
	currency := Currency("USD")

	t.Run("root has depth 0", func(t *testing.T) {
		root, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)

		accounts := map[AccountID]*Account{
			root.GetID(): root,
		}
		getParent := func(id AccountID) (*Account, error) {
			return accounts[id], nil
		}

		depth, err := root.Depth(getParent)
		require.NoError(t, err)
		require.Equal(t, 0, depth)
	})

	t.Run("child has depth 1", func(t *testing.T) {
		parent, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)
		child, _ := NewAccount("1100", "Cash", AccountTypeAsset, currency, nil)
		_ = child.SetParent(parent)

		accounts := map[AccountID]*Account{
			parent.GetID(): parent,
			child.GetID():  child,
		}
		getParent := func(id AccountID) (*Account, error) {
			return accounts[id], nil
		}

		depth, err := child.Depth(getParent)
		require.NoError(t, err)
		require.Equal(t, 1, depth)
	})

	t.Run("grandchild has depth 2", func(t *testing.T) {
		root, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)
		parent, _ := NewAccount("1100", "Current Assets", AccountTypeAsset, currency, nil)
		child, _ := NewAccount("1110", "Cash", AccountTypeAsset, currency, nil)

		_ = parent.SetParent(root)
		_ = child.SetParent(parent)

		accounts := map[AccountID]*Account{
			root.GetID():   root,
			parent.GetID(): parent,
			child.GetID():  child,
		}
		getParent := func(id AccountID) (*Account, error) {
			return accounts[id], nil
		}

		depth, err := child.Depth(getParent)
		require.NoError(t, err)
		require.Equal(t, 2, depth)
	})
}

func TestAccount_AccountPath(t *testing.T) {
	currency := Currency("USD")

	t.Run("path for root account", func(t *testing.T) {
		root, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)

		accounts := map[AccountID]*Account{
			root.GetID(): root,
		}
		getParent := func(id AccountID) (*Account, error) {
			return accounts[id], nil
		}

		path, err := root.AccountPath(getParent)
		require.NoError(t, err)
		require.Len(t, path, 1)
		require.Equal(t, root.GetID(), path[0])
	})

	t.Run("path for nested account", func(t *testing.T) {
		root, _ := NewAccount("1000", "Assets", AccountTypeAsset, currency, nil)
		parent, _ := NewAccount("1100", "Current Assets", AccountTypeAsset, currency, nil)
		child, _ := NewAccount("1110", "Cash", AccountTypeAsset, currency, nil)

		_ = parent.SetParent(root)
		_ = child.SetParent(parent)

		accounts := map[AccountID]*Account{
			root.GetID():   root,
			parent.GetID(): parent,
			child.GetID():  child,
		}
		getParent := func(id AccountID) (*Account, error) {
			return accounts[id], nil
		}

		path, err := child.AccountPath(getParent)
		require.NoError(t, err)
		require.Len(t, path, 3)
		require.Equal(t, root.GetID(), path[0])
		require.Equal(t, parent.GetID(), path[1])
		require.Equal(t, child.GetID(), path[2])
	})
}
*/

func TestReconstituteAccount(t *testing.T) {
	id := NewAccountID()
	parentID := NewAccountID()
	currency := Currency("USD")
	now := time.Now().UTC()
	balance, _ := money.New(500000, "USD") // $5000.00 in cents

	account, err := ReconstituteAccount(
		id,
		"1000",
		"Assets",
		AccountTypeAsset,
		currency,
		balance,
		&parentID,
		true,
		now,
		now,
	)

	require.NoError(t, err)
	require.Equal(t, id, account.GetID())
	require.Equal(t, "1000", account.GetCode())
	require.Equal(t, "Assets", account.GetName())
	require.Equal(t, AccountTypeAsset, account.GetType())
	require.NotNil(t, account.GetParentID())
	require.Equal(t, parentID, *account.GetParentID())
	require.True(t, account.IsActive())
}
