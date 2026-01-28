# MD-Spec-Tool

Tool to convert Excel to Markdown Specification.

## 🚀 Features

- Upload Excel files containing spec requirements
- Parse and preview table data
- Use Go templates to convert to Markdown
- Store and manage versions of spec documents
- User authentication with JWT
- Full-stack: Go backend + Next.js frontend

## 📋 Tech Stack

### Backend
- **Go 1.21+** with Gin framework
- **PostgreSQL 15** for database
- **excelize/v2** for Excel parsing
- **text/template** for Markdown generation
- **JWT** for authentication

### Frontend
- **Next.js 14** with TypeScript
- **Tailwind CSS** for styling
- **Zustand** for state management
- **React** components

## 🏗️ Project Structure

```
md-spec-tool/
├── backend/                 # Go API server
│   ├── cmd/server/         # Entry point
│   ├── internal/
│   │   ├── config/         # Configuration
│   │   ├── http/           # HTTP handlers & router
│   │   ├── services/       # Business logic
│   │   ├── repositories/   # Data access
│   │   ├── models/         # Data models
│   │   └── converters/     # Markdown conversion
│   ├── migrations/         # SQL migrations
│   ├── go.mod
│   └── Dockerfile
├── frontend/                # Next.js app
│   ├── app/                # App router
│   ├── components/         # React components
│   ├── lib/                # Utilities
│   ├── styles/             # CSS
│   └── Dockerfile
├── docker-compose.yml       # Local dev stack
├── Makefile                 # Commands
└── README.md
```

## ⚡ Quick Start

### Prerequisites
- Docker & Docker Compose
- Node.js 20+
- Go 1.21+

### Using Docker

```bash
# Start all services
make up

# View logs
make logs

# Stop services
make down
```

Services will be available at:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- PostgreSQL: localhost:5432

### Local Development

**Backend:**
```bash
cd backend
go mod download
go run ./cmd/server
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

## 📚 API Documentation

### Health Check
```bash
GET /health
```

### Auth Endpoints
```bash
POST /auth/login          # Email/password login
POST /auth/google         # Google OAuth
```

### Import Endpoints
```bash
POST /import/excel        # Upload Excel file
GET /spec/preview/:id     # Get parsed table data
```

### Conversion Endpoints
```bash
POST /convert/markdown    # Convert table to Markdown
GET /templates            # List templates
GET /templates/:id        # Get template details
```

### Spec Endpoints
```bash
POST /spec                # Save spec document
GET /spec/:id             # Get spec
GET /spec                 # List user specs
GET /spec/:id/versions    # List versions
```

## 🗄️ Database Setup

Migrations run automatically on startup. To manually apply:

```bash
docker-compose exec db psql -U mdspec -d mdspec < backend/migrations/001_create_users.up.sql
```

## 🧪 Testing

```bash
# Backend unit tests
make test

# Frontend tests (when configured)
cd frontend && npm test
```

## 📝 Environment Variables

See `.env.example` for all available options.

Key variables:
- `DB_DSN`: PostgreSQL connection string
- `JWT_SECRET`: Secret key for JWT signing
- `APP_ENV`: Environment (dev/prod)
- `NEXT_PUBLIC_API_URL`: Backend API URL for frontend

## 📖 Documentation

- [Implementation Plan](../md-spec-tool-implementation-plan.md)
- API Documentation (in progress)
- Template Guide (in progress)

## 🛣️ Roadmap

**Phase 1-7**: Full implementation plan in `md-spec-tool-implementation-plan.md`

Current Phase: **Phase 1** - Project Setup & Infrastructure ✓

## 🤝 Contributing

- Follow the phase-based implementation plan
- Create feature branches for each phase
- Test thoroughly before merging

## 📄 License

MIT

---

**Created**: January 2026  
**Version**: 1.0.0
