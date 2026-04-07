# Contributing

## Code Style

- Idiomatic Go (effective Go, Go proverbs)
- Godoc comments for all public types/functions
- Context propagation — ctx as first parameter
- No naked returns
- Errors wrapping: `fmt.Errorf("operationName: %w", err)`
- No `init()` functions
- No global variables (except ldflags in main)
- Interface segregation — small, focused interfaces

## Git Workflow

1. Create a branch from `main`
2. Commit in small, logical chunks
3. Conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`
4. PR with description of changes

## Before Push

```bash
make tidy
make lint
make test
```

## Adding New Bounded Context

1. Create structure `internal/{context}/`
2. Domain → Application → Infrastructure → Ports
3. Write tests (domain first)
4. Register routes in `main.go`
5. Update documentation