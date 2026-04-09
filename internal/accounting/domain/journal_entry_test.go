package domain

import (
	"testing"

	moneypkg "github.com/basilex/skeleton/pkg/money"
)

func TestNewJournalEntry(t *testing.T) {
	t.Run("creates_journal_entry_with_valid_data", func(t *testing.T) {
		je := NewJournalEntry("Test Entry", "user-123")

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

	t.Run("publishes_journal_entry_created_event", func(t *testing.T) {
		je := NewJournalEntry("Test Entry", "user-123")
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

	t.Run("adds_debit_line", func(t *testing.T) {
		je := NewJournalEntry("Test Entry", "user-123")
		debit, _ := moneypkg.New(10000, "USD") // $100.00
		credit := moneypkg.Zero("USD")

		err := je.AddLine(accountID, debit, credit, "Debit line")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if je.GetLineCount() != 1 {
			t.Errorf("expected 1 line, got %d", je.GetLineCount())
		}
	})

	t.Run("adds_credit_line", func(t *testing.T) {
		je := NewJournalEntry("Test Entry", "user-123")
		debit := moneypkg.Zero("USD")
		credit, _ := moneypkg.New(10000, "USD") // $100.00

		err := je.AddLine(accountID, debit, credit, "Credit line")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if je.GetLineCount() != 1 {
			t.Errorf("expected 1 line, got %d", je.GetLineCount())
		}
	})

	t.Run("fails_when_both_debit_and_credit_nonzero", func(t *testing.T) {
		je := NewJournalEntry("Test Entry", "user-123")
		debit, _ := moneypkg.New(10000, "USD") // $100.00
		credit, _ := moneypkg.New(5000, "USD")  // $50.00

		err := je.AddLine(accountID, debit, credit, "Invalid line")
		if err == nil {
			t.Error("expected error when both debit and credit are non-zero")
		}
	})

	t.Run("fails_when_both_debit_and_credit_zero", func(t *testing.T) {
		je := NewJournalEntry("Test Entry", "user-123")
		debit := moneypkg.Zero("USD")
		credit := moneypkg.Zero("USD")

		err := je.AddLine(accountID, debit, credit, "Empty line")
		if err == nil {
			t.Error("expected error when both debit and credit are zero")
		}
	})

	t.Run("fails_on_non_draft_entry", func(t *testing.T) {
		je := NewJournalEntry("Test Entry", "user-123")
		debit1, _ := moneypkg.New(10000, "USD")
		credit1 := moneypkg.Zero("USD")
		_ = je.AddLine(accountID, debit1, credit1, "Line 1")
		
		debit2 := moneypkg.Zero("USD")
		credit2, _ := moneypkg.New(10000, "USD")
		_ = je.AddLine(accountID, debit2, credit2, "Line 2")
		_ = je.Post()

		debit3, _ := moneypkg.New(5000, "USD")
		credit3 := moneypkg.Zero("USD")
		err := je.AddLine(accountID, debit3, credit3, "Line 3")
		if err == nil {
			t.Error("expected error when adding line to posted entry")
		}
	})
}

func TestJournalEntry_IsBalanced(t *testing.T) {
	accountID := NewAccountID()

	t.Run("balanced_entry", func(t *testing.T) {
		je := NewJournalEntry("Balanced Entry", "user-123")
		debit, _ := moneypkg.New(10000, "USD")
		credit := moneypkg.Zero("USD")
		_ = je.AddLine(accountID, debit, credit, "Debit")
		
		debit2 := moneypkg.Zero("USD")
		credit2, _ := moneypkg.New(10000, "USD")
		_ = je.AddLine(accountID, debit2, credit2, "Credit")

		if !je.IsBalanced() {
			t.Error("expected entry to be balanced")
		}
	})

	t.Run("unbalanced_entry", func(t *testing.T) {
		je := NewJournalEntry("Unbalanced Entry", "user-123")
		debit, _ := moneypkg.New(10000, "USD")
		credit := moneypkg.Zero("USD")
		_ = je.AddLine(accountID, debit, credit, "Debit")
		
		debit2 := moneypkg.Zero("USD")
		credit2, _ := moneypkg.New(5000, "USD")
		_ = je.AddLine(accountID, debit2, credit2, "Credit")

		if je.IsBalanced() {
			t.Error("expected entry to be unbalanced")
		}
	})

	t.Run("empty_entry_is_balanced", func(t *testing.T) {
		je := NewJournalEntry("Empty Entry", "user-123")
		// Empty entry has no lines, so total is zero

		if !je.IsBalanced() {
			t.Error("expected empty entry to be balanced")
		}
	})
}

func TestJournalEntry_Post(t *testing.T) {
	accountID := NewAccountID()

	t.Run("posts_balanced_entry", func(t *testing.T) {
		je := NewJournalEntry("Test Entry", "user-123")
		debit, _ := moneypkg.New(10000, "USD")
		credit := moneypkg.Zero("USD")
		_ = je.AddLine(accountID, debit, credit, "Debit")
		
		debit2 := moneypkg.Zero("USD")
		credit2, _ := moneypkg.New(10000, "USD")
		_ = je.AddLine(accountID, debit2, credit2, "Credit")

		err := je.Post()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if je.GetStatus() != JournalEntryStatusPosted {
			t.Errorf("expected status posted, got %s", je.GetStatus())
		}
	})

	t.Run("fails_on_unbalanced_entry", func(t *testing.T) {
		je := NewJournalEntry("Test Entry", "user-123")
		debit, _ := moneypkg.New(10000, "USD")
		credit := moneypkg.Zero("USD")
		_ = je.AddLine(accountID, debit, credit, "Debit")
		
		debit2 := moneypkg.Zero("USD")
		credit2, _ := moneypkg.New(5000, "USD")
		_ = je.AddLine(accountID, debit2, credit2, "Credit")

		err := je.Post()
		if err == nil {
			t.Error("expected error when posting unbalanced entry")
		}
	})

	t.Run("fails_with_insufficient_lines", func(t *testing.T) {
		je := NewJournalEntry("Test Entry", "user-123")
		debit, _ := moneypkg.New(10000, "USD")
		credit := moneypkg.Zero("USD")
		_ = je.AddLine(accountID, debit, credit, "Single line")

		err := je.Post()
		if err == nil {
			t.Error("expected error when posting entry with less than 2 lines")
		}
	})

	t.Run("fails_on_already_posted_entry", func(t *testing.T) {
		je := NewJournalEntry("Test Entry", "user-123")
		debit, _ := moneypkg.New(10000, "USD")
		credit := moneypkg.Zero("USD")
		_ = je.AddLine(accountID, debit, credit, "Debit")
		
		debit2 := moneypkg.Zero("USD")
		credit2, _ := moneypkg.New(10000, "USD")
		_ = je.AddLine(accountID, debit2, credit2, "Credit")
		_ = je.Post()

		err := je.Post()
		if err == nil {
			t.Error("expected error when posting already posted entry")
		}
	})

	t.Run("publishes_journal_entry_posted_event", func(t *testing.T) {
		je := NewJournalEntry("Test Entry", "user-123")
		debit, _ := moneypkg.New(10000, "USD")
		credit := moneypkg.Zero("USD")
		_ = je.AddLine(accountID, debit, credit, "Debit")
		
		debit2 := moneypkg.Zero("USD")
		credit2, _ := moneypkg.New(10000, "USD")
		_ = je.AddLine(accountID, debit2, credit2, "Credit")
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

	t.Run("voids_posted_entry", func(t *testing.T) {
		je := NewJournalEntry("Test Entry", "user-123")
		debit, _ := moneypkg.New(10000, "USD")
		credit := moneypkg.Zero("USD")
		_ = je.AddLine(accountID, debit, credit, "Debit")
		
		debit2 := moneypkg.Zero("USD")
		credit2, _ := moneypkg.New(10000, "USD")
		_ = je.AddLine(accountID, debit2, credit2, "Credit")
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
		je := NewJournalEntry("Test Entry", "user-123")

		err := je.Void("Mistake")
		if err == nil {
			t.Error("expected error when voiding draft entry")
		}
	})

	t.Run("publishes_journal_entry_voided_event", func(t *testing.T) {
		je := NewJournalEntry("Test Entry", "user-123")
		debit, _ := moneypkg.New(10000, "USD")
		credit := moneypkg.Zero("USD")
		_ = je.AddLine(accountID, debit, credit, "Debit")
		
		debit2 := moneypkg.Zero("USD")
		credit2, _ := moneypkg.New(10000, "USD")
		_ = je.AddLine(accountID, debit2, credit2, "Credit")
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

func TestJournalEntry_GetTotals(t *testing.T) {
	accountID := NewAccountID()

	je := NewJournalEntry("Test Entry", "user-123")
	debit1, _ := moneypkg.New(10000, "USD") // $100.00
	credit1 := moneypkg.Zero("USD")
	_ = je.AddLine(accountID, debit1, credit1, "Debit 1")
	
	debit2, _ := moneypkg.New(5000, "USD") // $50.00
	credit2 := moneypkg.Zero("USD")
	_ = je.AddLine(accountID, debit2, credit2, "Debit 2")
	
	debit3 := moneypkg.Zero("USD")
	credit3, _ := moneypkg.New(15000, "USD") // $150.00
	_ = je.AddLine(accountID, debit3, credit3, "Credit")

	totalDebits := je.GetTotalDebits()
	totalCredits := je.GetTotalCredits()

	if totalDebits.GetAmount() != 15000 {
		t.Errorf("expected total debits 15000 cents, got %d", totalDebits.GetAmount())
	}
	if totalCredits.GetAmount() != 15000 {
		t.Errorf("expected total credits 15000 cents, got %d", totalCredits.GetAmount())
	}
}
