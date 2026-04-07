# ADR-013: Language Policy

## Status

Accepted

## Context

Розробка проекту ведеться в Україні, але продукт орієнтований на міжнародне використання. Це створює виклики:

- Команди можуть бути інтернаціональними
- Open-source проекти вимагають англомовної документації
- Код-рев'ю та онбординг нових розробників ускладнюються українською
- Інтеграція з міжнародними сервісами та бібліотеками
- Публікація статей, презентацій на конференціях

Використання змішаних мов (українська в ADR, англійська в API/коді) призводить до:
- Непослідовності в документації
- Складності пошуку інформації
- Бар'єрів для не-українськомовних розробників

## Decision

**Всі матеріали проекту мають бути англійською мовою:**

### 1. Документація (*.md файли)

Всі файли документації, включаючи:
- ADR (Architecture Decision Records)
- README, CONTRIBUTING, CHANGELOG
- Development guides
- API documentation

**Винятки:**
- Лицензія (може бути двомовною, але основна - англійська)
- Юридичні документи (якщо вимагається законодавством)

### 2. Код (Go source files)

Все в коді має бути англійською:
- Comments (inline, doc comments)
- Variable names
- Function names
- Type names
- Package names
- Error messages
- Log messages

**Приклади:**

```go
// Good:✅
// CreateUser creates a new user with the given email and password.
// Returns an error if the email is already registered.
func CreateUser(email, password string) (*User, error) {
    if err := validateEmail(email); err != nil {
        return nil, fmt.Errorf("invalid email format: %w", err)
    }
    // ... implementation
}

// Bad: ❌
// CreateUser створює нового користувача з заданим email та паролем.
// Повертає помилку якщо email вже зареєстрований.
func CreateUser(email, password string) (*User, error) {
    // ... implementation
}
```

### 3. API (HTTP endpoints)

Всі API-endpoints:
- URL paths
- Request/Response JSON fields
- Error messages
- Query parameters

```json
// ✅ Good
{
  "error": {
    "code": "validation_error",
    "message": "Email address is required"
  }
}

// ❌ Bad
{
  "error": {
    "code": "помилка_валідації",
    "message": "Email адреса обов'язкова"
  }
}
```

### 4. Коміт меседжі

Git commit messages:
- Subject line: English
- Body: English
- Footer: English

```
✅ Good:
feat(notifications): add notification worker with retry logic

Implement background worker that processes pending notifications
with exponential backoff retry strategy.

- Add worker that polls for pending notifications
- Implement retry logic with configurable delays
- Add dead letter queue for failed notifications

❌ Bad:
feat(notifications): додати worker для обробки повідомлень

Реалізувати фоновий worker який обробляє повідомлення...
```

### 5. Назви та ідентифікатори

Всі ідентифікатори в системі:
- Database table names
- Column names
- Index names
- Migration files
- Environment variables
- Configuration keys

```sql
-- ✅ Good
CREATE TABLE notification_preferences (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    channels TEXT NOT NULL
);

-- ❌ Bad
CREATE TABLE налаштування_повідомлень (
    id TEXT PRIMARY KEY,
    користувач_id TEXT NOT NULL,
    канали TEXT NOT NULL
);
```

## Implementation

### Міграція існуючого коду

1. **Створення ADR-013** (цей документ) - затвердити політику
2. **Переклад документації** - всі .md файли перекласти на англійську
3. **Переклад коментарів** - всі коментарі в коді перекласти
4. **Оновлення API messages** - перекласти error messages та validation messages

### Процес

```bash
# Step 1: Find all documentation files
find . -name "*.md" -not -path "./vendor/*"

# Step 2: Find all Go source files with comments
find . -name "*.go" -not -path "./vendor/*"

# Step 3: Review and translate each file
# Use IDE or automated tools to assist

# Step 4: Run tests to ensure nothing broken
make test
```

### Code Review Checklist

При рев'ю коду перевіряти:
- [ ] Documentation is in English
- [ ] Comments are in English
- [ ] Variable/function names are in English
- [ ] Error messages are in English
- [ ] Log messages are in English
- [ ] No mixed languages in same file

## Consequences

### Positive
- ✅ Доступність для міжнародної аудиторії
- ✅ Легкість онбордингу нових розробників
- ✅ Відповідність open-source стандартам
- ✅ Єдність проекту (одна мова всюди)
- ✅ Краща інтегration з інструментами (IDE, generators)

### Negative
- ❌ Додаткові зусилля на переклад існуючого коду
- ❌ Можливі помилки перекладу (особливо термінів)
- ❌ Втрата контексту для україномовних розробників

### Neutral
- Може знадобитися глосарій термінів дляuni-сс
- Розробники можуть вести особисті нотатки українською

## Tools and Automation

### Automated Checks

Можна додати pre-commit hooks або CI checks:

```yaml
# .github/workflows/language-check.yml
name: Language Check
on: [pull_request]
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Check for non-English characters in comments
        run: |
          # Add custom script to detect Ukrainian/Russian characters
          # in comments and documentation
          ./scripts/check-language.sh
 ```

### IDEPlugins

Рекомендовані плагіни:
- **Code Spell Checker** - перевірка орфографії
- **Grammarly** - для коментарів та документації

## Examples of Translation

### ADR Title

```markdown
# Before (❌):
# ADR-004: Модель RBAC

# After (✅):
# ADR-004: RBAC Model
```

### Code Comments

```go
// Before (❌):
// Перевіряє чи має користувач дозвіл
func (u *User) HasPermission(perm Permission) bool {
    // ... implementation
}

// After (✅):
// HasPermission checks if the user has the given permission
func (u *User) HasPermission(perm Permission)bool {
    // ... implementation
}
```

### Error Messages

```go
// Before(❌):
return fmt.Errorf("користувача не знайдено")

// After (✅):
return fmt.Errorf("user not found")
```

## References

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)
- [Google Developer Documentation Style Guide](https://developers.google.com/style)
- [Microsoft Writing Style Guide](https://docs.microsoft.com/en-us/style-guide/)

## Migration Log

- 2026-04-07: ADR created
- [ ] Documentation translated
- [ ] Code comments translated
- [ ] API messages translated
- [ ] Migration complete