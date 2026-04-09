package domain

import (
	"testing"
)

func TestAccount_DebitCredit(t *testing.T) {
	account, err := NewAccount("1010", "Cash", AccountTypeAsset, CurrencyUAH, nil)
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	// Initial balance should be 0
	if account.GetBalance().Amount != 0 {
		t.Errorf("initial balance = %v, want 0", account.GetBalance().Amount)
	}

	// Debit asset account (increase)
	money, _ := NewMoney(1000, CurrencyUAH)
	err = account.Debit(money)
	if err != nil {
		t.Errorf("Debit() error = %v", err)
	}

	if account.GetBalance().Amount != 1000 {
		t.Errorf("balance after debit = %v, want 1000", account.GetBalance().Amount)
	}

	// Credit asset account (decrease)
	err = account.Credit(money)
	if err != nil {
		t.Errorf("Credit() error = %v", err)
	}

	if account.GetBalance().Amount != 0 {
		t.Errorf("balance after credit = %v, want 0", account.GetBalance().Amount)
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

			money, _ := NewMoney(100, CurrencyUAH)

			// Test debit
			err := account.Debit(money)
			if err != nil {
				t.Errorf("Debit() error = %v", err)
			}

			balanceAfterDebit := account.GetBalance().Amount

			// Test credit
			err = account.Credit(money)
			if err != nil {
				t.Errorf("Credit() error = %v", err)
			}

			balanceAfterCredit := account.GetBalance().Amount

			// Verify expected behavior based on account type
			if tt.debitEffect == "increase" {
				if balanceAfterDebit != 100 {
					t.Errorf("debit should increase %s account, balance = %v, want 100",
						tt.accountType, balanceAfterDebit)
				}
			} else {
				if balanceAfterDebit != -100 {
					t.Errorf("debit should decrease %s account, balance = %v, want -100",
						tt.accountType, balanceAfterDebit)
				}
			}

			// After credit, should be back to 0
			if balanceAfterCredit != 0 {
				t.Errorf("balance after credit = %v, want 0", balanceAfterCredit)
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
	money, _ := NewMoney(100, CurrencyUAH)
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
	money1, _ := NewMoney(100, CurrencyUAH)
	money2, _ := NewMoney(50, CurrencyUAH)

	// Test Add
	result, err := money1.Add(money2)
	if err != nil {
		t.Errorf("Add() error = %v", err)
	}
	if result.Amount != 150 {
		t.Errorf("Add() result = %v, want 150", result.Amount)
	}

	// Test Subtract
	result, err = money1.Subtract(money2)
	if err != nil {
		t.Errorf("Subtract() error = %v", err)
	}
	if result.Amount != 50 {
		t.Errorf("Subtract() result = %v, want 50", result.Amount)
	}

	// Cannot subtract more than available
	smallMoney, _ := NewMoney(10, CurrencyUAH)
	largeMoney, _ := NewMoney(100, CurrencyUAH)
	_, err = smallMoney.Subtract(largeMoney)
	if err == nil {
		t.Error("expected error when subtracting more than available")
	}

	// Cannot operate on different currencies
	usdMoney, _ := NewMoney(100, CurrencyUSD)
	_, err = money1.Add(usdMoney)
	if err == nil {
		t.Error("expected error when adding different currencies")
	}

	// Cannot create money with negative amount
	_, err = NewMoney(-100, CurrencyUAH)
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
