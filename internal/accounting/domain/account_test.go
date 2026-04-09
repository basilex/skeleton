package domain

import (
	"testing"

	moneypkg "github.com/basilex/skeleton/pkg/money"
)

func TestAccount_DebitCredit(t *testing.T) {
	account, err := NewAccount("1010", "Cash", AccountTypeAsset, CurrencyUAH, nil)
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	// Initial balance should be 0
	if account.GetBalance().GetAmount() != 0 {
		t.Errorf("initial balance = %v, want 0", account.GetBalance().GetAmount())
	}

	// Debit asset account (increase)
	money, _ := moneypkg.New(100000, "UAH") // 1000.00 UAH in cents
	err = account.Debit(money)
	if err != nil {
		t.Errorf("Debit() error = %v", err)
	}

	if account.GetBalance().GetAmount() != 100000 {
		t.Errorf("balance after debit = %v, want 100000", account.GetBalance().GetAmount())
	}

	// Credit asset account (decrease)
	err = account.Credit(money)
	if err != nil {
		t.Errorf("Credit() error = %v", err)
	}

	if account.GetBalance().GetAmount() != 0 {
		t.Errorf("balance after credit = %v, want 0", account.GetBalance().GetAmount())
	}
}

func TestAccount_DifferentAccountTypes(t *testing.T) {
	tests := []struct {
		name         string
		accountType  AccountType
		debitEffect  string
		creditEffect string
	}{
		{
			name:         "asset account",
			accountType:  AccountTypeAsset,
			debitEffect:  "increase",
			creditEffect: "decrease",
		},
		{
			name:         "liability account",
			accountType:  AccountTypeLiability,
			debitEffect:  "decrease",
			creditEffect: "increase",
		},
		{
			name:         "equity account",
			accountType:  AccountTypeEquity,
			debitEffect:  "decrease",
			creditEffect: "increase",
		},
		{
			name:         "revenue account",
			accountType:  AccountTypeRevenue,
			debitEffect:  "decrease",
			creditEffect: "increase",
		},
		{
			name:         "expense account",
			accountType:  AccountTypeExpense,
			debitEffect:  "increase",
			creditEffect: "decrease",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, _ := NewAccount("1000", "Test Account", tt.accountType, CurrencyUAH, nil)

			// For liability/equity/revenue accounts, we need to credit first to have a positive balance
			// before we can debit (which decreases these account types)
			if tt.creditEffect == "increase" {
				// Credit first to increase balance for liability/equity/revenue
				creditAmount, _ := moneypkg.New(10000, "UAH")
				err := account.Credit(creditAmount)
				if err != nil {
					t.Errorf("Credit() error = %v", err)
				}

				balanceAfterCredit := account.GetBalance().GetAmount()
				if balanceAfterCredit != 10000 {
					t.Errorf("credit should increase %s account, balance = %v, want 10000",
						tt.accountType, balanceAfterCredit)
				}

				// Then debit to decrease
				debitAmount, _ := moneypkg.New(5000, "UAH")
				err = account.Debit(debitAmount)
				if err != nil {
					t.Errorf("Debit() error = %v", err)
				}

				balanceAfterDebit := account.GetBalance().GetAmount()
				if balanceAfterDebit != 5000 {
					t.Errorf("debit should decrease %s account, balance = %v, want 5000",
						tt.accountType, balanceAfterDebit)
				}
			} else {
				// For asset/expense accounts, debit increases, credit decreases
				debitAmount, _ := moneypkg.New(10000, "UAH")
				err := account.Debit(debitAmount)
				if err != nil {
					t.Errorf("Debit() error = %v", err)
				}

				balanceAfterDebit := account.GetBalance().GetAmount()

				creditAmount, _ := moneypkg.New(10000, "UAH")
				err = account.Credit(creditAmount)
				if err != nil {
					t.Errorf("Credit() error = %v", err)
				}

				balanceAfterCredit := account.GetBalance().GetAmount()

				// Verify expected behavior based on account type
				if tt.debitEffect == "increase" {
					if balanceAfterDebit != 10000 {
						t.Errorf("debit should increase %s account, balance = %v, want 10000",
							tt.accountType, balanceAfterDebit)
					}
				}

				// After credit, should be back to 0
				if balanceAfterCredit != 0 {
					t.Errorf("balance after credit = %v, want 0", balanceAfterCredit)
				}
			}
		})
	}
}

func TestAccount_ActivateDeactivate(t *testing.T) {
	account, _ := NewAccount("1010", "Cash", AccountTypeAsset, CurrencyUAH, nil)

	// Start active
	if !account.IsActive() {
		t.Error("account should be active by default")
	}

	// Deactivate
	account.Deactivate()
	if account.IsActive() {
		t.Error("account should be inactive after deactivate")
	}

	// Cannot debit/credit inactive account
	money, _ := moneypkg.New(10000, "UAH") // 100.00 UAH in cents
	err := account.Debit(money)
	if err != ErrAccountInactive {
		t.Errorf("expected ErrAccountInactive, got %v", err)
	}

	// Activate
	account.Activate()
	if !account.IsActive() {
		t.Error("account should be active after activate")
	}

	// Now can debit/credit
	err = account.Debit(money)
	if err != nil {
		t.Errorf("Debit() error = %v", err)
	}
}

func TestMoney_Operations(t *testing.T) {
	money1, _ := moneypkg.New(10000, "UAH") // 100.00 UAH
	money2, _ := moneypkg.New(5000, "UAH")  // 50.00 UAH

	// Test Add
	result, err := money1.Add(money2)
	if err != nil {
		t.Errorf("Add() error = %v", err)
	}
	if result.GetAmount() != 15000 {
		t.Errorf("Add() result = %v, want 15000", result.GetAmount())
	}

	// Test Subtract
	result, err = money1.Subtract(money2)
	if err != nil {
		t.Errorf("Subtract() error = %v", err)
	}
	if result.GetAmount() != 5000 {
		t.Errorf("Subtract() result = %v, want 5000", result.GetAmount())
	}

	// Cannot create money with negative amount
	_, err = moneypkg.New(-100, "UAH")
	if err == nil {
		t.Error("expected error when creating money with negative amount")
	}
}

func TestAccountHierarchy(t *testing.T) {
	// Create parent account
	parentAccount, _ := NewAccount("1000", "Assets", AccountTypeAsset, CurrencyUAH, nil)

	// Create child account
	childAccount, err := NewAccount("1010", "Cash", AccountTypeAsset, CurrencyUAH, &parentAccount.id)
	if err != nil {
		t.Errorf("NewAccount() error = %v", err)
	}

	if childAccount.GetParentID() == nil {
		t.Error("expected parent ID to be set")
	}

	if *childAccount.GetParentID() != parentAccount.GetID() {
		t.Errorf("parent ID = %v, want %v", *childAccount.GetParentID(), parentAccount.GetID())
	}
}
