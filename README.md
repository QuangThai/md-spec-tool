# MD-Spec-Tool

Powerful tool to convert Excel/CSV/pasted data to structured Markdown specifications with intelligent header detection and customizable formatting.

## 🚀 Features

- **Multi-format Input**: Parse Excel (.xlsx), CSV, or paste table data directly
- **Smart Header Detection**: Automatically detect and handle table headers
- **Intelligent Data Processing**: Handle merged cells, formatting, and complex table structures
- **Markdown Spec Generation**: Convert to professional Markdown documentation
- **MDFlow Support**: Generate structured .mdflow format for advanced workflows
- **Live Preview**: Real-time conversion preview with error handling
- **Full-Stack Architecture**: Go backend API + Next.js 16 frontend with React 19

## 📋 Tech Stack

### Backend
- **Go 1.20+** with Gin framework
- **Internal Converter**: Smart parsing & Markdown generation
  - Header detection & column mapping
  - Matrix transformation & data validation
  - XLSX, CSV, and paste parser support
  - Template-based rendering (Go templates)
- **HTTP API**: RESTful endpoints with error handling

### Frontend
- **Next.js 16** with React 19 & TypeScript
- **Tailwind CSS 4** for styling (PostCSS integration)
- **Zustand 5** for state management
- **Framer Motion** for smooth animations
- **Lucide React** for icons

## 🏗️ Project Structure

```
md-spec-tool/
├── backend/                           # Go API server
│   ├── cmd/
│   │   ├── server/                   # Main server entry point
│   │   └── usecases/                 # CLI utility commands
│   ├── internal/
│   │   ├── config/                   # Configuration loading
│   │   ├── http/
│   │   │   ├── handlers/             # Request handlers (health, mdflow)
│   │   │   ├── middleware/           # CORS & auth middleware
│   │   │   └── router.go             # Route definitions
│   │   ├── converter/                # Core conversion logic
│   │   │   ├── xlsx_parser.go        # Excel file parsing
│   │   │   ├── paste_parser.go       # Pasted data parsing
│   │   │   ├── header_detect.go      # Smart header detection
│   │   │   ├── column_map.go         # Column mapping
│   │   │   ├── matrix.go             # Data matrix handling
│   │   │   ├── markdown_spec.go      # Markdown generation
│   │   │   ├── renderer.go           # Template rendering
│   │   │   └── model.go              # Data models
│   │   └── ...
│   ├── migrations/                   # Database migrations (if needed)
│   ├── go.mod & go.sum
│   └── Dockerfile
│
├── frontend/                          # Next.js app
│   ├── app/                          # App router structure
│   │   ├── layout.tsx                # Root layout
│   │   ├── page.tsx                  # Home page
│   │   ├── docs/                     # Documentation pages
│   │   └── studio/                   # Studio workspace
│   ├── components/                   # Reusable React components
│   ├── lib/
│   │   ├── mdflowApi.ts              # API client
│   │   ├── mdflowStore.ts            # Zustand state store
│   │   └── utils.ts                  # Utilities
│   ├── styles/                       # Global styles
│   ├── package.json
│   ├── tsconfig.json
│   ├── next.config.js
│   └── Dockerfile
│
├── docs/                             # Documentation
│   ├── IMPLEMENTATION_PLAN.md         # Detailed implementation roadmap
│   ├── TABLE_FORMATS.md               # Supported table formats
│   └── fixtures/                      # Test examples
│
├── use-cases/                        # Usage examples (example-1.md through example-5.md)
├── docker-compose.yml                # Local dev stack
├── Makefile                          # Build & dev commands
└── AGENTS.md                         # Agent configuration
```

## ⚡ Quick Start

### Prerequisites
- **Docker & Docker Compose** (for containerized dev)
- **Node.js 20+** (for frontend)
- **Go 1.20+** (for backend)
- **npm or yarn** (for frontend dependencies)

### Using Docker (Recommended)

```bash
# Build Docker images
make build

# Start all services in background
make up

# View logs
make logs

# Stop services
make down

# Clean up (remove containers & volumes)
make clean
```

**Available at:**
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080

### Local Development (No Docker)

**Terminal 1 - Backend:**
```bash
cd backend
go mod download
make dev-backend
# or: go run ./cmd/server
```

**Terminal 2 - Frontend:**
```bash
cd frontend
npm install
make dev-frontend
# or: npm run dev
```

## 📚 API Documentation

### Core Endpoints

**MDFlow Conversion** (Main functionality)
```bash
POST /mdflow/convert      # Convert table data to Markdown
  Input: { data: string, format: 'excel'|'paste' }
  Output: { markdown: string, mdflow: string }

GET /health               # Health check
```

**Input Formats Supported:**
- Excel files (.xlsx) with smart header detection
- Pasted table data (tab-separated, pipe-separated)
- Column mapping and merge handling

See [TABLE_FORMATS.md](docs/TABLE_FORMATS.md) for detailed format specifications.

## 🧑‍💻 Available Make Commands

```bash
make help          # Show all available commands
make build         # Build Docker images
make up            # Start services
make down          # Stop services
make logs          # View service logs
make clean         # Remove containers & volumes
make test          # Run backend tests (cd backend && go test ./...)
make dev-backend   # Run Go server in dev mode
make dev-frontend  # Run Next.js in dev mode
make dev           # Build & start all services with logs
```

## 🧪 Testing

```bash
# Run backend tests
make test
# or: cd backend && go test ./...

# Frontend tests (configure as needed)
cd frontend && npm test
```

## 📝 Environment Variables

See [`.env.example`](.env.example) for all available options.

**Key variables:**
- `HOST`: Server host (default: 0.0.0.0)
- `PORT`: Server port (default: 8080)
- `APP_ENV`: Environment (dev/prod)
- `NEXT_PUBLIC_API_URL`: Backend API URL for frontend (default: http://localhost:8080)

## 📖 Documentation

- **[IMPLEMENTATION_PLAN.md](docs/IMPLEMENTATION_PLAN.md)** - Complete implementation roadmap & phases
- **[TABLE_FORMATS.md](docs/TABLE_FORMATS.md)** - Supported table formats and specifications
- **[AGENTS.md](AGENTS.md)** - Development agent configuration
- **[use-cases/](use-cases/)** - Example conversions (example-1.md through example-5.md)

## 🛣️ Key Components

### Backend Converter Pipeline
The core conversion logic in `backend/internal/converter/`:
1. **Input Parsing**: Support for XLSX files and pasted data
2. **Header Detection**: Intelligent detection of table headers using scoring algorithm
3. **Column Mapping**: Map detected columns to spec fields
4. **Matrix Processing**: Handle merged cells and complex structures
5. **Markdown Rendering**: Template-based Markdown generation
6. **MDFlow Export**: Structured output format support

### Frontend Features
- Real-time preview of conversions
- Error handling and validation
- Studio workspace for advanced workflows
- Documentation viewer
- State management via Zustand

## 🤝 Contributing

1. Check [IMPLEMENTATION_PLAN.md](docs/IMPLEMENTATION_PLAN.md) for ongoing work
2. Follow code style guidelines in [AGENTS.md](AGENTS.md)
3. Test changes locally before pushing
4. Create feature branches for new functionality
