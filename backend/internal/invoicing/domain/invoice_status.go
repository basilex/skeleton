package domain

type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusViewed    InvoiceStatus = "viewed"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
)

func (s InvoiceStatus) String() string {
	return string(s)
}

func (s InvoiceStatus) CanTransitionTo(newStatus InvoiceStatus) bool {
	transitions := map[InvoiceStatus][]InvoiceStatus{
		InvoiceStatusDraft:     {InvoiceStatusSent, InvoiceStatusCancelled},
		InvoiceStatusSent:      {InvoiceStatusViewed, InvoiceStatusPaid, InvoiceStatusOverdue, InvoiceStatusCancelled},
		InvoiceStatusViewed:    {InvoiceStatusPaid, InvoiceStatusOverdue, InvoiceStatusCancelled},
		InvoiceStatusOverdue:   {InvoiceStatusPaid, InvoiceStatusCancelled},
		InvoiceStatusPaid:      {},
		InvoiceStatusCancelled: {},
	}

	allowed, exists := transitions[s]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == newStatus {
			return true
		}
	}

	return false
}
