package domain

import (
	"encoding/json"
	"fmt"
)

type BankAccount struct {
	BankName      string `json:"bank_name"`
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	SWIFTCode     string `json:"swift_code"`
	IBAN          string `json:"iban"`
	Currency      string `json:"currency"`
}

func NewBankAccount(bankName, accountName, accountNumber, currency string) (BankAccount, error) {
	ba := BankAccount{
		BankName:      bankName,
		AccountName:   accountName,
		AccountNumber: accountNumber,
		Currency:      currency,
	}

	if err := ba.Validate(); err != nil {
		return BankAccount{}, err
	}

	return ba, nil
}

func (ba BankAccount) Validate() error {
	if ba.BankName == "" {
		return fmt.Errorf("bank name is required")
	}
	if ba.AccountName == "" {
		return fmt.Errorf("account name is required")
	}
	if ba.AccountNumber == "" {
		return fmt.Errorf("account number is required")
	}
	if ba.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	if len(ba.Currency) != 3 {
		return fmt.Errorf("currency must be 3 characters (ISO 4217)")
	}
	return nil
}

func (ba BankAccount) ToJSON() (json.RawMessage, error) {
	data, err := json.Marshal(ba)
	if err != nil {
		return nil, fmt.Errorf("marshal bank account: %w", err)
	}
	return json.RawMessage(data), nil
}

func BankAccountFromJSON(data json.RawMessage) (BankAccount, error) {
	var ba BankAccount
	if len(data) == 0 {
		return BankAccount{}, nil
	}
	if err := json.Unmarshal(data, &ba); err != nil {
		return BankAccount{}, fmt.Errorf("unmarshal bank account: %w", err)
	}
	return ba, nil
}
