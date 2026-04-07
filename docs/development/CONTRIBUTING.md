# Contributing

## Code Style

- Ідіоматичний Go (effective Go, Go proverbs)
- Godoc коментарі для всіх публічних типів/функцій
- Context propagation — ctx перший параметр
- No naked returns
- Errors wrapping: `fmt.Errorf("operationName: %w", err)`
- No `init()` functions
- No global variables (окрім ldflags у main)
- Interface segregation — малі, focused інтерфейси

## Git Workflow

1. Створити гілку від `main`
2. Комітити маленькими, логічними шматками
3. Conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`
4. PR з описом змін

## Before Push

```bash
make tidy
make lint
make test
```

## Adding New Bounded Context

1. Створити структуру `internal/{context}/`
2. Domain → Application → Infrastructure → Ports
3. Написати тести (domain first)
4. Зареєструвати routes у `main.go`
5. Оновити документацію
