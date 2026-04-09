/**
 * Money Value Object (TypeScript)
 * 
 * Matches Go implementation in backend/pkg/money/money.go
 * - Stored as int64 cents for precision
 * - No floating-point errors
 * - Type-safe currency operations
 */
export class Money {
  constructor(
    private readonly amount: number,  // int64 cents
    private readonly currency: string
  ) {
    if (!Number.isInteger(amount)) {
      throw new Error('Amount must be an integer (cents)');
    }
    if (!currency || currency.length !== 3) {
      throw new Error('Currency must be a 3-letter code (e.g., USD, EUR)');
    }
  }

  /**
   * Get amount in cents (int64)
   */
  getAmount(): number {
    return this.amount;
  }

  /**
   * Get currency code (e.g., "USD", "EUR")
   */
  getCurrency(): string {
    return this.currency;
  }

  /**
   * Convert to float for display (e.g., 100.50)
   */
  toFloat(): number {
    return this.amount / 100;
  }

  /**
   * Format for display (e.g., "$100.50")
   */
  format(locale = 'en-US'): string {
    return new Intl.NumberFormat(locale, {
      style: 'currency',
      currency: this.currency,
    }).format(this.toFloat());
  }

  /**
   * Create Money from float (e.g., 100.50)
   */
  static fromFloat(value: number, currency: string): Money {
    if (value < 0) {
      throw new Error('Amount cannot be negative');
    }
    return new Money(Math.round(value * 100), currency);
  }

  /**
   * Create Money from cents (int64)
   */
  static fromCents(cents: number, currency: string): Money {
    return new Money(cents, currency);
  }

  /**
   * Create zero Money
   */
  static zero(currency: string): Money {
    return new Money(0, currency);
  }

  /**
   * Add two Money values
   */
  add(other: Money): Money {
    if (this.currency !== other.currency) {
      throw new Error(`Cannot add different currencies: ${this.currency} and ${other.currency}`);
    }
    return new Money(this.amount + other.amount, this.currency);
  }

  /**
   * Subtract Money
   */
  subtract(other: Money): Money {
    if (this.currency !== other.currency) {
      throw new Error(`Cannot subtract different currencies: ${this.currency} and ${other.currency}`);
    }
    return new Money(this.amount - other.amount, this.currency);
  }

  /**
   * Multiply by factor
   */
  multiply(factor: number): Money {
    if (factor < 0) {
      throw new Error('Factor cannot be negative');
    }
    return new Money(Math.round(this.amount * factor), this.currency);
  }

  /**
   * Divide by factor
   */
  divide(factor: number): Money {
    if (factor <= 0) {
      throw new Error('Factor must be positive');
    }
    return new Money(Math.round(this.amount / factor), this.currency);
  }

  /**
   * Check if zero
   */
  isZero(): boolean {
    return this.amount === 0;
  }

  /**
   * Check if negative
   */
  isNegative(): boolean {
    return this.amount < 0;
  }

  /**
   * Check if positive
   */
  isPositive(): boolean {
    return this.amount > 0;
  }

  /**
   * Compare equality
   */
  equals(other: Money): boolean {
    return this.amount === other.amount && this.currency === other.currency;
  }

  /**
   * Compare greater than
   */
  greaterThan(other: Money): boolean {
    if (this.currency !== other.currency) {
      throw new Error('Cannot compare different currencies');
    }
    return this.amount > other.amount;
  }

  /**
   * Compare less than
   */
  lessThan(other: Money): boolean {
    if (this.currency !== other.currency) {
      throw new Error('Cannot compare different currencies');
    }
    return this.amount < other.amount;
  }

  /**
   * Serialize for JSON
   */
  toJSON(): { amount: number; currency: string } {
    return {
      amount: this.amount,
      currency: this.currency,
    };
  }

  /**
   * Parse from JSON
   */
  static fromJSON(data: { amount: number; currency: string }): Money {
    return new Money(data.amount, data.currency);
  }

  /**
   * String representation
   */
  toString(): string {
    return `${this.currency} ${this.toFloat().toFixed(2)}`;
  }
}

/**
 * Currency codes
 */
export type Currency = 'USD' | 'EUR' | 'UAH' | 'GBP' | 'CAD' | 'AUD';

/**
 * Money utilities
 */
export const MoneyUtils = {
  /**
   * Sum multiple Money values (same currency)
   */
  sum(values: Money[]): Money {
    if (values.length === 0) {
      throw new Error('Cannot sum empty array');
    }
    
    const currency = values[0].currency;
    const total = values.reduce((sum, m) => {
      if (m.currency !== currency) {
        throw new Error('All Money values must have the same currency');
      }
      return sum + m.amount;
    }, 0);
    
    return new Money(total, currency);
  },

  /**
   * Parse from API response (float or cents)
   */
  parseFromAPI(value: number | { amount: number; currency: string }, currency?: Currency): Money {
    if (typeof value === 'object') {
      return Money.fromJSON(value);
    }
    
    // Assume value is float from API response
    return Money.fromFloat(value, currency || 'USD');
  },
};