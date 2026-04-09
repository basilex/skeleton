# План міграції на Варіант A (Монорепо з підкаталогами)

## 🎯 Цільова структура

```
skeleton/
├── backend/                    # Go API (поточний код)
│   ├── cmd/api/               # Entry point
│   ├── internal/              # Domain logic (9 bounded contexts)
│   ├── pkg/                   # Shared packages
│   ├── migrations/            # Database migrations
│   ├── go.mod
│   └── go.sum
├── frontend/                   # Next.js 14+ app
│   ├── app/                   # App Router
│   ├── components/            # shadcn/ui components
│   │   ├── ui/               # shadcn components
│   │   └── domain/            # Domain-specific components
│   ├── lib/                   # Utilities
│   │   ├── api/              # API client
│   │   └── utils/            # Helpers
│   ├── public/               # Static assets
│   ├── styles/              # Global styles
│   ├── types/                # TypeScript types
│   ├── next.config.js
│   ├── tailwind.config.ts
│   ├── tsconfig.json
│   └── package.json
├── shared/                    # Спільне між backend/frontend
│   ├── types/                # TypeScript types
│   │   ├── api.ts           # API types (generated from Go)
│   │   ├── domain.ts        # Domain types
│   │   └── money.ts         # Money value object
│   └── api-spec/            # OpenAPI spec
│       └── openapi.yaml
├── docs/                      # Документація (вже готова)
├── docker/                    # Docker конфігурація
│   ├── Dockerfile.backend    # Go backend image
│   ├── Dockerfile.frontend   # Next.js image
│   └── nginx/               # Nginx config for production
├── scripts/                   # Utility scripts
│   ├── generate-types.sh    # Generate TS from Go
│   ├── migrate.sh           # DB migrations helper
│   └── deploy.sh            # Deploy script
├── docker-compose.yml        # Development environment
├── docker-compose.prod.yml   # Production environment
├── Makefile                  # Build automation
├── README.md                 # Main readme
└── .gitignore
```

## 📋 Кроки міграції

### Крок 1: Підготовка (бекап + нові директорії)

```bash
# Створити нову структуру
mkdir -p backend frontend shared docker scripts
mkdir -p frontend/{app,components/{ui,domain},lib/{api,utils},public,styles,types}
mkdir -p shared/{types,api-spec}

# Зберегти поточний стан
git add .
git commit -m "chore: prepare for monorepo migration"
git branch backup-before-migration
```

### Крок 2: Переміщення Go коду

```bash
# Перемістити Go файли в backend/
mv cmd backend/
mv internal backend/
mv pkg backend/
mv migrations backend/
mv go.mod backend/
mv go.sum backend/

# Зберегти configs та інше
mv configs backend/
mv keys backend/

# Створити go.work в корені
cat > go.work << 'EOF'
go 1.25

use ./backend
EOF
```

### Крок 3: Ініціалізація Next.js

```bash
cd frontend
npx create-next-app@latest . \
  --typescript \
  --tailwind \
  --eslint \
  --app \
  --src-dir=false \
  --import-alias="@/*" \
  --use-npm

# Встановити shadcn/ui
npx shadcn@latest init
npx shadcn@latest add button card input select dialog
```

### Крок 4: Shared types

Створити базові TypeScript типи для Money та API.

### Крок 5: Docker Compose

Налаштувати єдиний docker-compose.yml для всього проекту.

### Крок 6: Makefile

Створити Makefile в корені для керування всім проектом.

## 🎯 Переваги для бізнесу

### Для клієнта:

```bash
# Клієнт отримує проект:
git clone <repo-url>
cd skeleton
cp .env.example .env

# Єдина команда запуску:
docker-compose up

# Все працює:
# - PostgreSQL: localhost:5432
# - Backend API: http://localhost:8080
# - Frontend: http://localhost:3000
```

### Для розробки:

- ✅ Єдина кодова база
- ✅ Спільні типи (type-safety)
- ✅ Єдиний docker-compose для розробки
- ✅ Простий CI/CD (build все разом)
- ✅ Легко деплоїти клієнту

### Для продажу:

- ✅ Простий продаж: "Запусти docker-compose і працює"
- ✅ Мінімальні вимоги: Docker + Docker Compose
- ✅ Повний стек в одному репозиторії
- ✅ Документація для всього проекту
- ✅ Приклади конфігурацій (.env.example)

## 📦 Що отримує клієнт

1. **Повний моноліт** - backend + frontend + database
2. **Готовий до продакшену** - Docker images, Nginx config
3. **Документація** - README, ARCHITECTURE, SETUP, DATABASE
4. **Міграції** - Всі SQL міграції включені
5. **Makefile** - Прості команди для керування

## 🚀 Наступні кроки після міграції

1. Налаштувати shared types генерацію з Go
2. Налаштувати API client для frontend
3. Створити базові UI компоненти для domain entities
4. Налаштувати CI/CD pipeline
5. Додати E2E тести