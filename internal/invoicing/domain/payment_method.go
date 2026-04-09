package domain

type PaymentMethod string

const (
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodCard         PaymentMethod = "card"
	PaymentMethodCash         PaymentMethod = "cash"
	PaymentMethodCheck        PaymentMethod = "check"
	PaymentMethodCrypto       PaymentMethod = "crypto"
)

func (m PaymentMethod) String() string {
	return string(m)
}
