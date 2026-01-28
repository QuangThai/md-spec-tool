# MD-Spec-Tool Implementation Plan

> **Tool to convert Excel to Markdown Specification**
> Version: 1.0 | Created: January 2026

---

## 📋 Table of Contents

1. [Project Overview](#project-overview)
2. [System Architecture](#system-architecture)
3. [Tech Stack](#tech-stack)
4. [API Design](#api-design)
5. [Phase 1: Project Setup & Infrastructure](#phase-1-project-setup--infrastructure)
6. [Phase 2: Backend Core (Auth, Upload, Parsing)](#phase-2-backend-core-auth-upload-parsing)
7. [Phase 3: Template Engine & Conversion](#phase-3-template-engine--conversion)
8. [Phase 4: Document Storage & Versioning](#phase-4-document-storage--versioning)
9. [Phase 5: Frontend Implementation](#phase-5-frontend-implementation)
10. [Phase 6: Integration & Testing](#phase-6-integration--testing)
11. [Phase 7: Deployment & Documentation](#phase-7-deployment--documentation)
12. [Risk Management](#risk-management)
13. [Timeline Summary](#timeline-summary)

---

## Project Overview

### Objectives
Build the **md-spec-tool** to:
- Upload Excel files containing spec requirements
- Parse and preview table data
- Use Go templates to convert to Markdown
- Store and manage versions of spec documents

### User Flow
```
Upload Excel → Select sheet/map columns → Preview table → Convert to Markdown → Edit if needed → Export/Save
```

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         md-spec-tool                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────┐         ┌──────────────────────────────┐  │
│  │    Frontend      │         │         Backend              │  │
│  │    (Next.js)     │ ──────► │         (Go/Gin)             │  │
│  │                  │   REST  │                              │  │
│  │  Port: 3000      │   API   │  Port: 8080                  │  │
│  └──────────────────┘         └─────────────┬────────────────┘  │
│                                             │                   │
│                                             ▼                   │
│                               ┌──────────────────────────────┐  │
│                               │       PostgreSQL             │  │
│                               │       Port: 5432             │  │
│                               └──────────────────────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Project Structure

```
md-spec-tool/
├── frontend/                    # Next.js/React TypeScript
│   ├── app/
│   │   ├── page.tsx            # Landing page
│   │   ├── auth/
│   │   │   └── login/page.tsx
│   │   ├── converter/
│   │   │   └── page.tsx
│   │   ├── templates/
│   │   │   └── page.tsx
│   │   └── documents/
│   │       ├── page.tsx
│   │       └── [id]/page.tsx
│   ├── components/
│   │   ├── excel-upload-zone.tsx
│   │   ├── table-preview.tsx
│   │   ├── json-viewer.tsx
│   │   ├── markdown-preview.tsx
│   │   └── template-selector.tsx
│   ├── lib/
│   │   └── api.ts
│   └── package.json
│
├── backend/                     # Go with Gin framework
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go
│   │   ├── http/
│   │   │   ├── router.go
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go
│   │   │   │   └── cors.go
│   │   │   └── handlers/
│   │   │       ├── auth_handler.go
│   │   │       ├── import_handler.go
│   │   │       ├── convert_handler.go
│   │   │       └── spec_handler.go
│   │   ├── services/
│   │   │   ├── auth_service.go
│   │   │   ├── excel_service.go
│   │   │   ├── template_service.go
│   │   │   └── spec_service.go
│   │   ├── repositories/
│   │   │   ├── user_repository.go
│   │   │   ├── template_repository.go
│   │   │   └── spec_repository.go
│   │   ├── models/
│   │   │   ├── user.go
│   │   │   ├── template.go
│   │   │   ├── table_data.go
│   │   │   └── spec.go
│   │   └── converters/
│   │       └── markdown_converter.go
│   ├── migrations/
│   │   ├── 001_create_users.up.sql
│   │   ├── 002_create_templates.up.sql
│   │   └── 003_create_specs.up.sql
│   └── go.mod
│
├── docker-compose.yml
├── Makefile
└── README.md
```

---

## Tech Stack

### Backend
| Component | Technology | Purpose |
|-----------|------------|---------|
| Language | Go 1.21+ | Performance, concurrency |
| Framework | Gin | HTTP routing, middleware |
| Excel Parsing | excelize/v2 | Parse .xlsx, .xls files |
| Template Engine | text/template | Markdown generation |
| Database | PostgreSQL 15 | Document storage |
| Auth | JWT + bcrypt | Authentication |
| Migration | golang-migrate | DB migrations |

### Frontend
| Component | Technology | Purpose |
|-----------|------------|---------|
| Framework | Next.js 14 | React SSR/SSG |
| Language | TypeScript | Type safety |
| Styling | Tailwind CSS | Utility-first CSS |
| State | Zustand/React Context | State management |
| Markdown | react-markdown | Render preview |
| HTTP Client | fetch/axios | API calls |

### Infrastructure
| Component | Technology | Purpose |
|-----------|------------|---------|
| Containerization | Docker | Deployment consistency |
| Orchestration | docker-compose | Local/prod deployment |
| Reverse Proxy | Nginx/Caddy | TLS, routing |

---

## API Design

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/login` | Email/password login |
| POST | `/auth/google` | Google OAuth login |

### Excel Import
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/import/excel` | Upload Excel file (multipart/form-data) |
| GET | `/spec/preview/:id` | Get parsed table data |

### Conversion
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/convert/markdown` | Convert table data to Markdown |
| GET | `/templates` | List available templates |
| GET | `/templates/:id` | Get template details |

### Spec Documents
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/spec` | Save spec document |
| GET | `/spec/:id` | Get spec document |
| GET | `/spec/:id/versions` | List all versions |
| GET | `/spec` | List user's documents |

---

## Phase 1: Project Setup & Infrastructure

### 🎯 Objectives
Backend, frontend, and DB run via `docker-compose` with health endpoints.

### ⏱️ Effort: 1-3 days

### ✅ Tasks

#### 1.1 Repository & Structure Setup
- [ ] Create repo `md-spec-tool/` with the structure above
- [ ] Initialize Go module:
  ```bash
  cd backend
  go mod init github.com/yourorg/md-spec-tool
  go get github.com/gin-gonic/gin
  go get github.com/jackc/pgx/v5
  go get github.com/xuri/excelize/v2
  go get github.com/joho/godotenv
  go get github.com/golang-jwt/jwt/v5
  go get golang.org/x/crypto/bcrypt
  ```
- [ ] Initialize Next.js app:
   ```bash
   npx create-next-app@latest frontend --typescript --tailwind --app
   ```

#### 1.2 Backend Scaffold
- [ ] Create `cmd/server/main.go` entrypoint
- [ ] Create `internal/http/router.go` with Gin router
- [ ] Create `internal/config/config.go` for env loading
- [ ] Add health endpoint: `GET /health` returns `{"status":"ok"}`
- [ ] Setup CORS middleware

#### 1.3 Database Setup
- [ ] Create `docker-compose.yml`:
  ```yaml
  services:
    db:
      image: postgres:15
      environment:
        POSTGRES_DB: mdspec
        POSTGRES_USER: mdspec
        POSTGRES_PASSWORD: mdspec
      ports:
        - "5432:5432"
      volumes:
        - postgres_data:/var/lib/postgresql/data

    backend:
      build: ./backend
      ports:
        - "8080:8080"
      depends_on:
        - db
      environment:
        - DB_DSN=postgres://mdspec:mdspec@db:5432/mdspec?sslmode=disable

    frontend:
      build: ./frontend
      ports:
        - "3000:3000"
      depends_on:
        - backend

  volumes:
    postgres_data:
  ```
- [ ] Create initial migrations:
  - `001_create_users.up.sql`
  - `002_create_templates.up.sql`
  - `003_create_specs.up.sql`

#### 1.4 Frontend Shell
- [ ] Setup base pages: `/`, `/auth/login`, `/converter`, `/templates`, `/documents`
- [ ] Create layout với navigation bar
- [ ] Setup API client trong `lib/api.ts`

#### 1.5 Dev Tooling
- [ ] Create `Makefile`:
  ```makefile
  run-backend:
    cd backend && go run cmd/server/main.go
  
  run-frontend:
    cd frontend && npm run dev
  
  docker-up:
    docker-compose up -d
  
  migrate:
    migrate -path backend/migrations -database "$$DB_DSN" up
  ```

### 📊 Success Criteria
- [ ] `docker-compose up` runs Postgres + backend + frontend
- [ ] `GET /health` responds from backend
- [ ] All core pages load in browser
- [ ] Frontend can call `/health` and display result

---

## Phase 2: Backend Core (Auth, Upload, Parsing)

### 🎯 Mục tiêu
Implement authentication, Excel file upload, sheet selection, và JSON table representation.

### ⏱️ Effort: 1-3 ngày

### ✅ Tasks

#### 2.1 Database Schema - Users
```sql
-- 001_create_users.up.sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    password_hash VARCHAR(255),
    auth_provider VARCHAR(50) DEFAULT 'local',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### 2.2 Auth Service Implementation
- [ ] Create `internal/services/auth_service.go`:
  ```go
  type AuthService interface {
      AuthenticatePassword(email, password string) (*User, error)
      CreateUserFromGoogle(profile GoogleProfile) (*User, error)
      GenerateToken(userID string) (string, error)
      ParseToken(token string) (string, error)
  }
  ```
- [ ] Implement JWT token generation/validation
- [ ] Implement bcrypt password hashing

#### 2.3 Auth Endpoints
- [ ] `POST /auth/login`:
  - Request: `{ "email": "string", "password": "string" }`
  - Response: `{ "user": {...} }` + JWT cookie
- [ ] `POST /auth/google`:
  - Request: `{ "id_token": "string" }`
  - Response: same as login

#### 2.4 Table Data Model
```go
// internal/models/table_data.go
type TableColumn struct {
    Key    string `json:"key"`
    Header string `json:"header"`
    Index  int    `json:"index"`
}

type TableRow map[string]interface{}

type TableData struct {
    SheetName string        `json:"sheet_name"`
    Columns   []TableColumn `json:"columns"`
    Rows      []TableRow    `json:"rows"`
}
```

#### 2.5 Excel Service
- [ ] Create `internal/services/excel_service.go`:
  ```go
  type ExcelService interface {
      ParseExcel(file multipart.File, sheetName string) (*TableData, error)
      GetSheetNames(file multipart.File) ([]string, error)
  }
  ```
- [ ] Use `excelize/v2` for parsing
- [ ] Handle missing headers, empty rows

#### 2.6 Import Endpoint
- [ ] `POST /import/excel`:
  - Content-Type: `multipart/form-data`
  - Fields: `file` (required), `sheet` (optional)
  - Response:
    ```json
    {
      "import_id": "uuid",
      "table_data": { ... }
    }
    ```
- [ ] Store preview in `spec_previews` table:
  ```sql
  CREATE TABLE spec_previews (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      table_json JSONB NOT NULL,
      created_by UUID REFERENCES users(id),
      created_at TIMESTAMPTZ DEFAULT NOW(),
      expires_at TIMESTAMPTZ DEFAULT NOW() + INTERVAL '24 hours'
  );
  ```

#### 2.7 Preview Endpoint
- [ ] `GET /spec/preview/:id`:
  - Returns stored `TableData` as JSON

### 📊 Success Criteria
- [ ] Can login with seeded user and receive JWT cookie
- [ ] Upload sample Excel returns parsed `table_data`
- [ ] `GET /spec/preview/:id` returns same data

---

## Phase 3: Template Engine & Conversion

### 🎯 Mục tiêu
Cung cấp Markdown conversion từ `table_data` sử dụng Go templates với helper functions.

### ⏱️ Effort: 1-3 ngày

### ✅ Tasks

#### 3.1 Template Database Schema
```sql
-- 002_create_templates.up.sql
CREATE TABLE templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    content TEXT NOT NULL,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### 3.2 Seed Default Templates
```sql
INSERT INTO templates (name, description, content) VALUES
('Specification Table', 'Basic spec table format', 
'# Feature Specification: {{.Meta.FeatureName}}

## Screen: {{.Meta.ScreenName}}

| No | Item Name | Type | Requirement | Description |
|----|-----------|------|-------------|--------------|
{{range .Table.Rows}}| {{.No}} | {{.ItemName}} | {{.ItemType}} | {{.Required}} | {{.Description}} |
{{end}}

## Notes
{{.Meta.Notes}}
'),
('Requirement List', 'List format for requirements',
'{{range .Table.Rows}}
### Requirement Item {{.No}}

- **Name:** {{.ItemName}}
- **Type:** {{.ItemType}}
- **Required:** {{.Required}}
- **Description:** {{.Description}}

{{end}}
');
```

#### 3.3 Markdown Context Struct
```go
// internal/models/markdown_context.go
type MarkdownContext struct {
    Table  TableData              `json:"table"`
    Meta   map[string]interface{} `json:"meta,omitempty"`
    Params map[string]interface{} `json:"params,omitempty"`
}
```

#### 3.4 Template Service with FuncMap
- [ ] Create `internal/services/template_service.go`:
  ```go
  func NewMarkdownTemplate(name, content string) (*template.Template, error) {
      return template.New(name).Funcs(template.FuncMap{
          "escapeMD": escapeMD,
          "upper":    strings.ToUpper,
          "lower":    strings.ToLower,
          "title":    strings.Title,
          "join":     strings.Join,
          "default":  defaultValue,
          "bold":     func(s string) string { return "**" + s + "**" },
          "italic":   func(s string) string { return "*" + s + "*" },
          "code":     func(s string) string { return "`" + s + "`" },
      }).Parse(content)
  }
  ```

#### 3.5 Convert Endpoint
- [ ] `POST /convert/markdown`:
  - Request:
    ```json
    {
      "template_id": "uuid",
      "preview_id": "uuid",       // OR
      "table_data": { ... },
      "meta": { "FeatureName": "Login", "ScreenName": "Auth" }
    }
    ```
  - Response:
    ```json
    {
      "markdown": "# Feature Specification...",
      "template_id": "uuid"
    }
    ```

#### 3.6 Template List Endpoint
- [ ] `GET /templates`: List all templates
- [ ] `GET /templates/:id`: Get template details

### 📊 Success Criteria
- [ ] Given sample `table_data` and seeded template, `/convert/markdown` returns valid Markdown
- [ ] Template helper functions work correctly

---

## Phase 4: Document Storage & Versioning

### 🎯 Mục tiêu
Lưu trữ specs với history, hỗ trợ retrieval và version listing.

### ⏱️ Effort: 2-4 ngày

### ✅ Tasks

#### 4.1 Database Schema
```sql
-- 003_create_specs.up.sql
CREATE TABLE spec_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    template_id UUID REFERENCES templates(id),
    latest_version INT DEFAULT 1,
    owner_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE spec_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    spec_id UUID REFERENCES spec_documents(id) ON DELETE CASCADE,
    version INT NOT NULL,
    table_json JSONB NOT NULL,
    markdown TEXT NOT NULL,
    meta JSONB,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(spec_id, version)
);

CREATE INDEX idx_spec_versions_spec_id ON spec_versions(spec_id);
```

#### 4.2 Spec Service
- [ ] Create `internal/services/spec_service.go`:
  ```go
  type SpecService interface {
      CreateSpec(ownerID, title, templateID string, tableData TableData, markdown string, meta map[string]interface{}) (*Spec, error)
      AddVersion(specID string, tableData TableData, markdown string, meta map[string]interface{}, userID string) (*SpecVersion, error)
      GetSpec(specID string) (*SpecWithLatestVersion, error)
      ListVersions(specID string) ([]SpecVersion, error)
      ListUserSpecs(userID string) ([]Spec, error)
  }
  ```

#### 4.3 Endpoints Implementation
- [ ] `POST /spec`:
  - Request:
    ```json
    {
      "title": "Login Feature Spec",
      "template_id": "uuid",
      "table_data": { ... },
      "markdown": "...",
      "meta": { ... }
    }
    ```
  - Response: `{ "id": "uuid", "version": 1 }`

- [ ] `GET /spec/:id`:
  - Response:
    ```json
    {
      "id": "uuid",
      "title": "Login Feature Spec",
      "template_id": "uuid",
      "latest_version": 3,
      "latest": {
        "version": 3,
        "table_data": { ... },
        "markdown": "...",
        "meta": { ... },
        "created_at": "..."
      }
    }
    ```

- [ ] `GET /spec/:id/versions`:
  - Response:
    ```json
    {
      "spec_id": "uuid",
      "versions": [
        { "version": 1, "created_at": "...", "meta": { ... } },
        { "version": 2, "created_at": "...", "meta": { ... } }
      ]
    }
    ```

- [ ] `GET /spec`: List user's specs

#### 4.4 Optional: Add New Version
- [ ] `POST /spec/:id/versions`: Add new version to existing spec

### 📊 Success Criteria
- [ ] Can create spec via `/spec` and retrieve it
- [ ] Can list all versions of a spec
- [ ] Versions increment correctly

---

## Phase 5: Frontend Implementation

### 🎯 Mục tiêu
Implement UX cho login, upload, preview, conversion, template selection, và documents pages.

### ⏱️ Effort: 3-7 ngày

### ✅ Tasks

#### 5.1 Auth Flow
- [ ] `/auth/login` page:
  - Email/password form → `POST /auth/login`
  - Google login button
- [ ] Auth context/store for user state
- [ ] Protected routes middleware

#### 5.2 Excel Upload Component
```tsx
// components/excel-upload-zone.tsx
- [ ] Drag-and-drop area với visual feedback
- [ ] File picker fallback
- [ ] File type validation (.xlsx, .xls)
- [ ] Size limit validation (10MB)
- [ ] Upload progress indicator
- [ ] Error handling with toast notifications
```

#### 5.3 Table Preview Component
```tsx
// components/table-preview.tsx
- [ ] Render parsed data as HTML table
- [ ] Show sheet info (name, row count, column count)
- [ ] Limit preview to first 10 rows
- [ ] Horizontal scroll for many columns
```

#### 5.4 JSON Viewer Component
```tsx
// components/json-viewer.tsx
- [ ] Formatted JSON display
- [ ] Syntax highlighting
- [ ] Collapsible sections
- [ ] Copy to clipboard
```

#### 5.5 Template Selector Component
```tsx
// components/template-selector.tsx
- [ ] Fetch templates from API
- [ ] Display template name + description
- [ ] Radio/select for template choice
- [ ] Preview template content (optional)
```

#### 5.6 Markdown Preview Component
```tsx
// components/markdown-preview.tsx
- [ ] Tabs: "Preview" | "Source"
- [ ] react-markdown for rendering
- [ ] Editable source textarea
- [ ] Copy/download buttons
```

#### 5.7 Converter Page (`/converter`)
- [ ] Step 1: Excel upload zone
- [ ] Step 2: Table preview (shown after upload)
- [ ] Step 3: Template selector
- [ ] Step 4: "Generate Markdown" button
- [ ] Step 5: Markdown preview
- [ ] Step 6: "Save Document" button

#### 5.8 Templates Page (`/templates`)
- [ ] List all templates
- [ ] Show template details
- [ ] (Future: CRUD templates)

#### 5.9 Documents Page (`/documents`)
- [ ] List user's documents
- [ ] Search/filter
- [ ] Click to view details

#### 5.10 Document Detail Page (`/documents/[id]`)
- [ ] Show document info
- [ ] Version dropdown
- [ ] Display markdown for selected version
- [ ] Download button

### 📊 Success Criteria
- [ ] User can login
- [ ] User can upload Excel and preview table
- [ ] User can select template and generate Markdown
- [ ] User can save document
- [ ] User can view saved documents and versions

---

## Phase 6: Integration & Testing

### 🎯 Objectives
Ensure E2E behavior is reliable with automated tests.

### ⏱️ Effort: 2-4 days

### ✅ Tasks

#### 6.1 Backend Unit Tests
- [ ] Template rendering helpers tests
- [ ] Excel parsing tests (with sample files)
- [ ] Spec service tests
- [ ] Auth service tests

#### 6.2 Backend Integration Tests
- [ ] `/import/excel` E2E with sample file
- [ ] `/convert/markdown` with known template + data
- [ ] `/spec` CRUD operations

#### 6.3 Frontend Tests
- [ ] Component tests with Jest + React Testing Library:
   - [ ] Excel upload zone
   - [ ] Template selector
   - [ ] Markdown preview
- [ ] Page rendering tests

#### 6.4 API Contract Documentation
- [ ] OpenAPI/Swagger spec
- [ ] Postman collection for manual testing

#### 6.5 E2E Sanity Checks
- [ ] Full flow: login → upload → convert → save → view
- [ ] Edge cases:
   - [ ] Excel with missing headers
   - [ ] Empty rows
   - [ ] Wrong sheet name
   - [ ] Large file (near limit)

### 📊 Success Criteria
- [ ] CI pipeline runs tests on push
- [ ] Core endpoints have unit coverage
- [ ] Main user flows work without errors

---

## Phase 7: Deployment & Documentation

### 🎯 Objectives
Containerized deployment with clear documentation.

### ⏱️ Effort: 1-3 days

### ✅ Tasks

#### 7.1 Dockerization
- [ ] Backend Dockerfile (multi-stage build):
  ```dockerfile
  # Build stage
  FROM golang:1.21-alpine AS builder
  WORKDIR /app
  COPY go.* ./
  RUN go mod download
  COPY . .
  RUN CGO_ENABLED=0 go build -o server ./cmd/server

  # Runtime stage
  FROM alpine:3.19
  RUN apk add --no-cache ca-certificates
  COPY --from=builder /app/server /server
  EXPOSE 8080
  CMD ["/server"]
  ```

- [ ] Frontend Dockerfile:
  ```dockerfile
  FROM node:20-alpine AS builder
  WORKDIR /app
  COPY package*.json ./
  RUN npm ci
  COPY . .
  RUN npm run build

  FROM node:20-alpine
  WORKDIR /app
  COPY --from=builder /app/.next ./.next
  COPY --from=builder /app/public ./public
  COPY --from=builder /app/package*.json ./
  RUN npm ci --production
  EXPOSE 3000
  CMD ["npm", "start"]
  ```

#### 7.2 Environment Configuration
- [ ] Document required env vars:
  - `DB_DSN`
  - `JWT_SECRET`
  - `GOOGLE_CLIENT_ID`
  - `GOOGLE_CLIENT_SECRET`
  - `CORS_ORIGINS`
  - `APP_ENV` (dev/prod)

#### 7.3 Production Setup
- [ ] Add Nginx/Caddy for TLS termination
- [ ] Health checks in docker-compose
- [ ] Environment-specific configs

#### 7.4 Documentation
- [ ] `README.md`:
   - Project overview
   - Architecture diagram
   - Setup instructions
   - Example commands

- [ ] `API.md`:
   - All endpoints documented
   - Request/response examples

- [ ] `TEMPLATES.md`:
   - Go template syntax guide
   - Available helper functions
   - Example templates

#### 7.5 Operational Basics
- [ ] Structured logging
- [ ] Standard error response format
- [ ] DB backup script

### 📊 Success Criteria
- [ ] Single command deploys the app
- [ ] New developer can follow README to run locally
- [ ] Logs visible, health checks pass

---

## Risk Management

### ⚠️ Risks & Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Template complexity | Medium | Medium | Keep helpers small, document well |
| Excel variability (merged cells, missing headers) | High | High | Strict validation, clear error messages |
| Auth security (dev stubs in prod) | High | Low | Env flag, fail startup if misconfigured |
| Performance on large files | Medium | Medium | File size limits, row count limits |

### 🛡️ Guardrails
1. **Template debugging**: Log template errors with context
2. **Excel validation**: Reject files without header row
3. **Auth safety**: `GOOGLE_OAUTH_DISABLED_FOR_DEV` flag
4. **Size limits**: Max 10MB file, 10,000 rows

---

## Timeline Summary

| Phase | Description | Effort | Dependencies |
|-------|-------------|--------|--------------|
| **Phase 1** | Project Setup & Infrastructure | 1-3 days | None |
| **Phase 2** | Backend Core (Auth, Upload) | 1-3 days | Phase 1 |
| **Phase 3** | Template Engine & Conversion | 1-3 days | Phase 2 |
| **Phase 4** | Document Storage & Versioning | 2-4 days | Phase 2-3 |
| **Phase 5** | Frontend Implementation | 3-7 days | Phase 2-4 |
| **Phase 6** | Integration & Testing | 2-4 days | Phase 2-5 |
| **Phase 7** | Deployment & Documentation | 1-3 days | All phases |

**Total Estimated Time: 11-27 days** (depending on complexity and parallel work)

---

## Best Practices Applied

### From Librarian Research

1. **Excel Parsing (excelize/v2)**
   - Use `excelize.Options{}` for flexible row reading
   - Always close files with `defer f.Close()`
   - Handle empty sheets gracefully

2. **Go Templates**
   - Use template registry pattern
   - Implement FuncMap with Markdown helpers
   - Cache compiled templates

3. **REST API Design**
   - Handler → Service → Repository pattern
   - Standard error response format
   - JWT with HTTP-only cookies

4. **Frontend Components**
   - Drag-drop with file validation
   - Tabs for source/preview
   - Loading states on all API calls

---

## Future Enhancements

When the tool is working stably:

1. **Template Management**
   - Template versioning
   - Template composition/reuse
   - Approval workflows

2. **Column Mapping UI**
   - UI để map Excel columns → logical fields
   - Save mappings per-template

3. **Collaboration**
   - Team/organization support
   - Granular permissions (RBAC)
   - Audit logging

4. **Performance**
   - Background jobs for large files
   - Caching layer
   - CDN for static assets

---

*Document created: January 2026*  
*Last updated: January 2026*
