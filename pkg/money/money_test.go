package money_test

import (
	"testing"

	"github.com/basilex/skeleton/pkg/money"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("valid money creation", func(t *testing.T) {
		m, err := money.New(1234, "USD")
		require.NoError(t, err)
		require.Equal(t, int64(1234), m.GetAmount())
		require.Equal(t, "USD", m.GetCurrency())
	})

	t.Run("negative amount returns error", func(t *testing.T) {
		_, err := money.New(-100, "USD")
		require.Error(t, err)
		require.Equal(t, money.ErrNegativeAmount, err)
	})

	t.Run("empty currency returns error", func(t *testing.T) {
		_, err := money.New(100, "")
		require.Error(t, err)
	})

	t.Run("invalid currency length returns error", func(t *testing.T) {
		_, err := money.New(100, "US")
		require.Error(t, err)
		require.Equal(t, money.ErrInvalidCurrency, err)
	})

	t.Run("lowercase currency is converted to uppercase", func(t *testing.T) {
		m, err := money.New(100, "usd")
		require.NoError(t, err)
		require.Equal(t, "USD", m.GetCurrency())
	})
}

func TestNewFromFloat(t *testing.T) {
	t.Run("convert float to money", func(t *testing.T) {
		m, err := money.NewFromFloat(12.34, "USD")
		require.NoError(t, err)
		require.Equal(t, int64(1234), m.GetAmount())
	})

	t.Run("negative amount returns error", func(t *testing.T) {
		_, err := money.NewFromFloat(-12.34, "USD")
		require.Error(t, err)
		require.Equal(t, money.ErrNegativeAmount, err)
	})

	t.Run("rounding works correctly", func(t *testing.T) {
		m, err := money.NewFromFloat(12.345, "USD")
		require.NoError(t, err)
		require.Equal(t, int64(1235), m.GetAmount())
	})
}

func TestAdd(t *testing.T) {
	t.Run("add same currencies", func(t *testing.T) {
		m1, _ := money.New(1000, "USD")
		m2, _ := money.New(500, "USD")

		result, err := m1.Add(m2)
		require.NoError(t, err)
		require.Equal(t, int64(1500), result.GetAmount())
	})

	t.Run("add different currencies returns error", func(t *testing.T) {
		m1, _ := money.New(1000, "USD")
		m2, _ := money.New(500, "EUR")

		_, err := m1.Add(m2)
		require.Error(t, err)
		require.Equal(t, money.ErrDifferentCurrencies, err)
	})
}

func TestSubtract(t *testing.T) {
	t.Run("subtract same currencies", func(t *testing.T) {
		m1, _ := money.New(1000, "USD")
		m2, _ := money.New(300, "USD")

		result, err := m1.Subtract(m2)
		require.NoError(t, err)
		require.Equal(t, int64(700), result.GetAmount())
	})

	t.Run("subtract larger amount returns error", func(t *testing.T) {
		m1, _ := money.New(300, "USD")
		m2, _ := money.New(500, "USD")

		_, err := m1.Subtract(m2)
		require.Error(t, err)
		require.Equal(t, money.ErrNegativeAmount, err)
	})

	t.Run("subtract different currencies returns error", func(t *testing.T) {
		m1, _ := money.New(1000, "USD")
		m2, _ := money.New(300, "EUR")

		_, err := m1.Subtract(m2)
		require.Error(t, err)
		require.Equal(t, money.ErrDifferentCurrencies, err)
	})
}

func TestMultiply(t *testing.T) {
	t.Run("multiply by positive factor", func(t *testing.T) {
		m, _ := money.New(1000, "USD")

		result, err := m.Multiply(1.5)
		require.NoError(t, err)
		require.Equal(t, int64(1500), result.GetAmount())
	})

	t.Run("multiply by negative factor returns error", func(t *testing.T) {
		m, _ := money.New(1000, "USD")

		_, err := m.Multiply(-1.5)
		require.Error(t, err)
		require.Equal(t, money.ErrNegativeAmount, err)
	})

	t.Run("multiply by zero", func(t *testing.T) {
		m, _ := money.New(1000, "USD")

		result, err := m.Multiply(0)
		require.NoError(t, err)
		require.Equal(t, int64(0), result.GetAmount())
	})
}

func TestDivide(t *testing.T) {
	t.Run("divide by positive factor", func(t *testing.T) {
		m, _ := money.New(1000, "USD")

		result, err := m.Divide(2)
		require.NoError(t, err)
		require.Equal(t, int64(500), result.GetAmount())
	})

	t.Run("divide by zero returns error", func(t *testing.T) {
		m, _ := money.New(1000, "USD")

		_, err := m.Divide(0)
		require.Error(t, err)
	})

	t.Run("divide by negative factor returns error", func(t *testing.T) {
		m, _ := money.New(1000, "USD")

		_, err := m.Divide(-2)
		require.Error(t, err)
	})
}

func TestToFloat64(t *testing.T) {
	t.Run("convert to float", func(t *testing.T) {
		m, _ := money.New(1234, "USD")
		require.Equal(t, 12.34, m.ToFloat64())
	})

	t.Run("zero amount", func(t *testing.T) {
		m, _ := money.New(0, "USD")
		require.Equal(t, 0.0, m.ToFloat64())
	})
}

func TestString(t *testing.T) {
	t.Run("format money as string", func(t *testing.T) {
		m, _ := money.New(1234, "USD")
		require.Equal(t, "12.34 USD", m.String())
	})
}

func TestEquals(t *testing.T) {
	t.Run("equal money", func(t *testing.T) {
		m1, _ := money.New(1000, "USD")
		m2, _ := money.New(1000, "USD")
		require.True(t, m1.Equals(m2))
	})

	t.Run("different amounts", func(t *testing.T) {
		m1, _ := money.New(1000, "USD")
		m2, _ := money.New(2000, "USD")
		require.False(t, m1.Equals(m2))
	})

	t.Run("different currencies", func(t *testing.T) {
		m1, _ := money.New(1000, "USD")
		m2, _ := money.New(1000, "EUR")
		require.False(t, m1.Equals(m2))
	})
}

func TestIsZero(t *testing.T) {
	t.Run("zero amount", func(t *testing.T) {
		m, _ := money.New(0, "USD")
		require.True(t, m.IsZero())
	})

	t.Run("positive amount", func(t *testing.T) {
		m, _ := money.New(100, "USD")
		require.False(t, m.IsZero())
	})
}

func TestIsPositive(t *testing.T) {
	t.Run("positive amount", func(t *testing.T) {
		m, _ := money.New(100, "USD")
		require.True(t, m.IsPositive())
	})

	t.Run("zero amount", func(t *testing.T) {
		m, _ := money.New(0, "USD")
		require.False(t, m.IsPositive())
	})
}

func TestComparisonOperations(t *testing.T) {
	m1, _ := money.New(1000, "USD")
	m2, _ := money.New(2000, "USD")
	m3, _ := money.New(1000, "USD")

	t.Run("less than", func(t *testing.T) {
		require.True(t, m1.LessThan(m2))
		require.False(t, m2.LessThan(m1))
		require.False(t, m1.LessThan(m3))
	})

	t.Run("greater than", func(t *testing.T) {
		require.True(t, m2.GreaterThan(m1))
		require.False(t, m1.GreaterThan(m2))
		require.False(t, m1.GreaterThan(m3))
	})

	t.Run("less than or equal", func(t *testing.T) {
		require.True(t, m1.LessThanOrEqual(m2))
		require.True(t, m1.LessThanOrEqual(m3))
		require.False(t, m2.LessThanOrEqual(m1))
	})

	t.Run("greater than or equal", func(t *testing.T) {
		require.True(t, m2.GreaterThanOrEqual(m1))
		require.True(t, m1.GreaterThanOrEqual(m3))
		require.False(t, m1.GreaterThanOrEqual(m2))
	})
}

func TestChainedOperations(t *testing.T) {
	t.Run("add multiple times", func(t *testing.T) {
		m1, _ := money.New(1000, "USD")
		m2, _ := money.New(500, "USD")
		m3, _ := money.New(250, "USD")

		result, err := m1.Add(m2)
		require.NoError(t, err)

		result, err = result.Add(m3)
		require.NoError(t, err)

		require.Equal(t, int64(1750), result.GetAmount())
	})

	t.Run("complex calculation", func(t *testing.T) {
		// Test: (100 + 50 - 25) * 2 = 250
		m1, _ := money.New(10000, "USD")
		m2, _ := money.New(5000, "USD")
		m3, _ := money.New(2500, "USD")

		result, err := m1.Add(m2)
		require.NoError(t, err)

		result, err = result.Subtract(m3)
		require.NoError(t, err)

		result, err = result.Multiply(2)
		require.NoError(t, err)

		require.Equal(t, int64(25000), result.GetAmount()) // 250.00 USD
	})
}
