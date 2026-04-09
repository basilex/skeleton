package domain

import (
	"testing"
)

func TestNewJournalEntry(t *testing.T) {
	t.Run("creates_journal_entry_with_valid_data", func(t *testing.T) {
		je, err := NewJournalEntry("Test Entry", "user-123")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if je.GetID() == "" {
			t.Error("expected journal entry ID to be set")
		}
		if je.GetStatus() != JournalEntryStatusDraft {
			t.Errorf("expected status draft, got %s", je.GetStatus())
		}
		if je.GetLineCount() != 0 {
			t.Errorf("expected 0 lines, got %d", je.GetLineCount())
		}
	})

	t.Run("fails_with_empty_description", func(t *testing.T) {
		_, err := NewJournalEntry("", "user-123")
		if err == nil {
			t.Error("expected error for empty description")
		}
	})

	t.Run("fails_with_empty_created_by", func(t *testing.T) {
		_, err := NewJournalEntry("Test Entry", "")
		if err == nil {
			t.Error("expected error for empty created by")
		}
	})

	t.Run("publishes_journal_entry_created_event", func(t *testing.T) {
		je, _ := NewJournalEntry("Test Entry", "user-123")
		events := je.PullEvents()

		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		_, ok := events[0].(JournalEntryCreated)
		if !ok {
			t.Error("expected JournalEntryCreated event")
		}
	})
}

func TestJournalEntry_AddLine(t *testing.T) {
	accountID := NewAccountID()
	usd := CurrencyUSD

	t.Run("adds_debit_line", func(t *testing.T) {
		je, _ := NewJournalEntry("Test Entry", "user-123")
		debit := Money{Amount: 100.00, Currency: usd}
		credit := Money{Amount: 0, Currency: usd}

		err := je.AddLine(accountID, debit, credit, "Debit line")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if je.GetLineCount() != 1 {
			t.Errorf("expected 1 line, got %d", je.GetLineCount())
		}
	})

	t.Run("adds_credit_line", func(t *testing.T) {
		je, _ := NewJournalEntry("Test Entry", "user-123")
		debit := Money{Amount: 0, Currency: usd}
		credit := Money{Amount: 100.00, Currency: usd}

		err := je.AddLine(accountID, debit, credit, "Credit line")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if je.GetLineCount() != 1 {
			t.Errorf("expected 1 line, got %d", je.GetLineCount())
		}
	})

	t.Run("fails_when_both_debit_and_credit_nonzero", func(t *testing.T) {
		je, _ := NewJournalEntry("Test Entry", "user-123")
		debit := Money{Amount: 100.00, Currency: usd}
		credit := Money{Amount: 50.00, Currency: usd}

		err := je.AddLine(accountID, debit, credit, "Invalid line")
		if err == nil {
			t.Error("expected error when both debit and credit are non-zero")
		}
	})

	t.Run("fails_when_both_debit_and_credit_zero", func(t *testing.T) {
		je, _ := NewJournalEntry("Test Entry", "user-123")
		debit := Money{Amount: 0, Currency: usd}
		credit := Money{Amount: 0, Currency: usd}

		err := je.AddLine(accountID, debit, credit, "Empty line")
		if err == nil {
			t.Error("expected error when both debit and credit are zero")
		}
	})

	t.Run("fails_on_non_draft_entry", func(t *testing.T) {
		je, _ := NewJournalEntry("Test Entry", "user-123")
		_ = je.AddLine(accountID, Money{Amount: 100, Currency: usd}, Money{Amount: 0, Currency: usd}, "Line 1")
		_ = je.AddLine(accountID, Money{Amount: 0, Currency: usd}, Money{Amount: 100, Currency: usd}, "Line 2")
		_ = je.Post()

		err := je.AddLine(accountID, Money{Amount: 50, Currency: usd}, Money{Amount: 0, Currency: usd}, "Line 3")
		if err == nil {
			t.Error("expected error when adding line to posted entry")
		}
	})
}

func TestJournalEntry_IsBalanced(t *testing.T) {
	accountID := NewAccountID()
	usd := CurrencyUSD

	t.Run("balanced_entry", func(t *testing.T) {
		je, _ := NewJournalEntry("Balanced Entry", "user-123")
		_ = je.AddLine(accountID, Money{Amount: 100, Currency: usd}, Money{Amount: 0, Currency: usd}, "Debit")
		_ = je.AddLine(accountID, Money{Amount: 0, Currency: usd}, Money{Amount: 100, Currency: usd}, "Credit")

		if !je.IsBalanced() {
			t.Error("expected entry to be balanced")
		}
	})

	t.Run("unbalanced_entry", func(t *testing.T) {
		je, _ := NewJournalEntry("Unbalanced Entry", "user-123")
		_ = je.AddLine(accountID, Money{Amount: 100, Currency: usd}, Money{Amount: 0, Currency: usd}, "Debit")
		_ = je.AddLine(accountID, Money{Amount: 0, Currency: usd}, Money{Amount: 50, Currency: usd}, "Credit")

		if je.IsBalanced() {
			t.Error("expected entry to be unbalanced")
		}
	})

	t.Run("empty_entry_is_balanced", func(t *testing.T) {
		je, _ := NewJournalEntry("Empty Entry", "user-123")
		// Empty entry has no lines, so total is zero

		if !je.IsBalanced() {
			t.Error("expected empty entry to be balanced")
		}
	})
}

func TestJournalEntry_Post(t *testing.T) {
	accountID := NewAccountID()
	usd := CurrencyUSD

	t.Run("posts_balanced_entry", func(t *testing.T) {
		je, _ := NewJournalEntry("Test Entry", "user-123")
		_ = je.AddLine(accountID, Money{Amount: 100, Currency: usd}, Money{Amount: 0, Currency: usd}, "Debit")
		_ = je.AddLine(accountID, Money{Amount: 0, Currency: usd}, Money{Amount: 100, Currency: usd}, "Credit")

		err := je.Post()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if je.GetStatus() != JournalEntryStatusPosted {
			t.Errorf("expected status posted, got %s", je.GetStatus())
		}
	})

	t.Run("fails_on_unbalanced_entry", func(t *testing.T) {
		je, _ := NewJournalEntry("Test Entry", "user-123")
		_ = je.AddLine(accountID, Money{Amount: 100, Currency: usd}, Money{Amount: 0, Currency: usd}, "Debit")
		_ = je.AddLine(accountID, Money{Amount: 0, Currency: usd}, Money{Amount: 50, Currency: usd}, "Credit")

		err := je.Post()
		if err == nil {
			t.Error("expected error when posting unbalanced entry")
		}
	})

	t.Run("fails_with_insufficient_lines", func(t *testing.T) {
		je, _ := NewJournalEntry("Test Entry", "user-123")
		_ = je.AddLine(accountID, Money{Amount: 100, Currency: usd}, Money{Amount: 0, Currency: usd}, "Single line")

		err := je.Post()
		if err == nil {
			t.Error("expected error when posting entry with less than 2 lines")
		}
	})

	t.Run("fails_on_already_posted_entry", func(t *testing.T) {
		je, _ := NewJournalEntry("Test Entry", "user-123")
		_ = je.AddLine(accountID, Money{Amount: 100, Currency: usd}, Money{Amount: 0, Currency: usd}, "Debit")
		_ = je.AddLine(accountID, Money{Amount: 0, Currency: usd}, Money{Amount: 100, Currency: usd}, "Credit")
		_ = je.Post()

		err := je.Post()
		if err == nil {
			t.Error("expected error when posting already posted entry")
		}
	})

	t.Run("publishes_journal_entry_posted_event", func(t *testing.T) {
		je, _ := NewJournalEntry("Test Entry", "user-123")
		_ = je.AddLine(accountID, Money{Amount: 100, Currency: usd}, Money{Amount: 0, Currency: usd}, "Debit")
		_ = je.AddLine(accountID, Money{Amount: 0, Currency: usd}, Money{Amount: 100, Currency: usd}, "Credit")
		_ = je.Post()

		events := je.PullEvents()
		if len(events) != 2 {
			t.Fatalf("expected 2 events (created + posted), got %d", len(events))
		}

		_, ok := events[1].(JournalEntryPosted)
		if !ok {
			t.Error("expected JournalEntryPosted event")
		}
	})
}

func TestJournalEntry_Void(t *testing.T) {
	accountID := NewAccountID()
	usd := CurrencyUSD

	t.Run("voids_posted_entry", func(t *testing.T) {
		je, _ := NewJournalEntry("Test Entry", "user-123")
		_ = je.AddLine(accountID, Money{Amount: 100, Currency: usd}, Money{Amount: 0, Currency: usd}, "Debit")
		_ = je.AddLine(accountID, Money{Amount: 0, Currency: usd}, Money{Amount: 100, Currency: usd}, "Credit")
		_ = je.Post()

		err := je.Void("Mistake")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if je.GetStatus() != JournalEntryStatusVoided {
			t.Errorf("expected status voided, got %s", je.GetStatus())
		}
	})

	t.Run("fails_on_draft_entry", func(t *testing.T) {
		je, _ := NewJournalEntry("Test Entry", "user-123")

		err := je.Void("Mistake")
		if err == nil {
			t.Error("expected error when voiding draft entry")
		}
	})

	t.Run("publishes_journal_entry_voided_event", func(t *testing.T) {
		je, _ := NewJournalEntry("Test Entry", "user-123")
		_ = je.AddLine(accountID, Money{Amount: 100, Currency: usd}, Money{Amount: 0, Currency: usd}, "Debit")
		_ = je.AddLine(accountID, Money{Amount: 0, Currency: usd}, Money{Amount: 100, Currency: usd}, "Credit")
		_ = je.Post()
		_ = je.Void("Mistake")

		events := je.PullEvents()
		if len(events) != 3 {
			t.Fatalf("expected 3 events (created + posted + voided), got %d", len(events))
		}

		_, ok := events[2].(JournalEntryVoided)
		if !ok {
			t.Error("expected JournalEntryVoided event")
		}
	})
}

func TestJournalLine_IsDebit_IsCredit(t *testing.T) {
	usd := CurrencyUSD
	accountID := NewAccountID()

	t.Run("debit_line", func(t *testing.T) {
		line := JournalLine{
			accountID: accountID,
			debit:     Money{Amount: 100, Currency: usd},
			credit:    Money{Amount: 0, Currency: usd},
		}

		if !line.IsDebit() {
			t.Error("expected line to be debit")
		}
		if line.IsCredit() {
			t.Error("expected line not to be credit")
		}
	})

	t.Run("credit_line", func(t *testing.T) {
		line := JournalLine{
			accountID: accountID,
			debit:     Money{Amount: 0, Currency: usd},
			credit:    Money{Amount: 100, Currency: usd},
		}

		if !line.IsCredit() {
			t.Error("expected line to be credit")
		}
		if line.IsDebit() {
			t.Error("expected line not to be debit")
		}
	})
}

func TestJournalEntry_GetTotals(t *testing.T) {
	accountID := NewAccountID()
	usd := CurrencyUSD

	je, _ := NewJournalEntry("Test Entry", "user-123")
	_ = je.AddLine(accountID, Money{Amount: 100, Currency: usd}, Money{Amount: 0, Currency: usd}, "Debit 1")
	_ = je.AddLine(accountID, Money{Amount: 50, Currency: usd}, Money{Amount: 0, Currency: usd}, "Debit 2")
	_ = je.AddLine(accountID, Money{Amount: 0, Currency: usd}, Money{Amount: 150, Currency: usd}, "Credit")

	totalDebits := je.GetTotalDebits()
	totalCredits := je.GetTotalCredits()

	if totalDebits.Amount != 150 {
		t.Errorf("expected total debits 150, got %f", totalDebits.Amount)
	}
	if totalCredits.Amount != 150 {
		t.Errorf("expected total credits 150, got %f", totalCredits.Amount)
	}
}
