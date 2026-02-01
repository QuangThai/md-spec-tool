# MD-Spec-Tool MVP Roadmap

> **Vision**: Trở thành "Notion of spreadsheet documentation" - delightful, collaborative, extensible.

## 📊 Current State Analysis

### ✅ Đã có
- Multi-format input (Excel, CSV, paste, Google Sheets)
- Smart header detection
- Template-based rendering
- Diff comparison
- AI suggestions (OpenAI)
- Batch processing
- Live preview
- History tracking
- Dark theme UI với Framer Motion animations

### ❌ Chưa có
- Authentication & user accounts
- Sharing & collaboration
- Template marketplace
- Payment/monetization
- Command palette (⌘K)
- Public specs gallery
- API access for developers

---

## 🎯 MVP Phases

### Phase 1: UX Polish & Power User Features
**Timeline**: 1-2 tuần  
**Goal**: Làm cho tool smooth và delightful hơn

#### 1.1 Command Palette (⌘K)
```
Priority: HIGH
Effort: 3-4 days
```
- [ ] Implement command palette component (cmdk library)
- [ ] Commands: Convert, Export, Copy, Toggle Preview, Switch Template
- [ ] Fuzzy search for templates
- [ ] Recent actions history
- [ ] Keyboard navigation (arrow keys, enter, esc)

#### 1.2 Enhanced Keyboard Shortcuts
```
Priority: HIGH
Effort: 1-2 days
```
| Shortcut | Action |
|----------|--------|
| `⌘K` | Open command palette |
| `⌘Enter` | Run conversion |
| `⌘Shift C` | Copy output |
| `⌘Shift E` | Export to file |
| `⌘P` | Toggle preview |
| `⌘,` | Open settings |
| `⌘/` | Show all shortcuts |

#### 1.3 Loading States Polish
```
Priority: MEDIUM
Effort: 1-2 days
```
- [ ] Replace spinners with skeleton screens
- [ ] Add shimmer effect for loading content
- [ ] Optimistic UI updates
- [ ] Progress indicators for batch operations

#### 1.4 Micro-interactions
```
Priority: MEDIUM
Effort: 2-3 days
```
- [ ] Button press feedback (scale animation)
- [ ] Success checkmark animation (SVG draw)
- [ ] File drop zone pulse effect
- [ ] Toast notifications với slide + fade
- [ ] Copy button → checkmark transition

#### 1.5 Responsive Design Polish
```
Priority: MEDIUM
Effort: 1-2 days
```
- [ ] Mobile-first adjustments
- [ ] Collapsible panels on small screens
- [ ] Touch-friendly controls
- [ ] Swipe gestures for panels

**Phase 1 Deliverables:**
- Command palette với fuzzy search
- Complete keyboard shortcut system
- Polished loading states
- Enhanced animations

---

### Phase 2: Authentication & User Accounts
**Timeline**: 1-2 tuần  
**Goal**: User identity và persistent data

#### 2.1 Authentication Setup
```
Priority: HIGH
Effort: 3-4 days
```
**Option A: Clerk (Recommended for speed)**
- [ ] Install @clerk/nextjs
- [ ] Setup GitHub OAuth
- [ ] Google OAuth
- [ ] Email/password fallback
- [ ] User profile page

**Option B: NextAuth.js (More control)**
- [ ] Configure NextAuth
- [ ] GitHub provider
- [ ] Google provider
- [ ] JWT session strategy

#### 2.2 Database Schema
```
Priority: HIGH
Effort: 2-3 days
```
```sql
-- Users
CREATE TABLE users (
  id UUID PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  name VARCHAR(255),
  avatar_url TEXT,
  subscription_tier VARCHAR(50) DEFAULT 'free',
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Documents (converted specs)
CREATE TABLE documents (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  title VARCHAR(255) NOT NULL,
  input_type VARCHAR(50), -- xlsx, csv, paste, gsheet
  input_data TEXT,
  output_markdown TEXT,
  template_id UUID REFERENCES templates(id),
  meta JSONB,
  is_public BOOLEAN DEFAULT FALSE,
  slug VARCHAR(255) UNIQUE,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Templates
CREATE TABLE templates (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  name VARCHAR(255) NOT NULL,
  description TEXT,
  content TEXT NOT NULL,
  is_public BOOLEAN DEFAULT FALSE,
  is_official BOOLEAN DEFAULT FALSE,
  downloads_count INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Document versions (history)
CREATE TABLE document_versions (
  id UUID PRIMARY KEY,
  document_id UUID REFERENCES documents(id),
  output_markdown TEXT,
  meta JSONB,
  created_at TIMESTAMP DEFAULT NOW()
);
```

#### 2.3 User Dashboard
```
Priority: HIGH
Effort: 2-3 days
```
- [ ] My Documents list với search/filter
- [ ] Recent conversions
- [ ] Saved templates
- [ ] Usage stats (conversions this month)
- [ ] Account settings

#### 2.4 Document Persistence
```
Priority: HIGH
Effort: 2 days
```
- [ ] Auto-save conversions to database
- [ ] Document naming và organization
- [ ] Version history per document
- [ ] Restore previous versions

**Phase 2 Deliverables:**
- GitHub/Google OAuth login
- User dashboard
- Document persistence
- Version history

---

### Phase 3: Sharing & Collaboration
**Timeline**: 2-3 tuần  
**Goal**: Viral growth through sharing

#### 3.1 Public Sharing Links
```
Priority: HIGH
Effort: 3-4 days
```
- [ ] Generate shareable URLs: `md-spec.app/s/abc123`
- [ ] Custom slugs: `md-spec.app/s/my-api-spec`
- [ ] OG image generation for social preview
- [ ] View counter
- [ ] Copy link button

#### 3.2 Permission System
```
Priority: HIGH
Effort: 2-3 days
```
```typescript
type Permission = 'view' | 'comment' | 'edit' | 'admin';

interface Share {
  documentId: string;
  email?: string;        // Specific user
  isPublic?: boolean;    // Anyone with link
  permission: Permission;
  expiresAt?: Date;
}
```
- [ ] Share với specific emails
- [ ] Public/private toggle
- [ ] Permission levels (view, comment, edit)
- [ ] Expiring links option

#### 3.3 Comments System
```
Priority: MEDIUM
Effort: 3-4 days
```
- [ ] Inline comments on output
- [ ] Comment threads
- [ ] @mentions
- [ ] Email notifications
- [ ] Resolve/unresolve comments

#### 3.4 Real-time Collaboration (Optional for MVP)
```
Priority: LOW (Phase 4)
Effort: 1-2 weeks
```
- [ ] Liveblocks integration
- [ ] Cursor presence
- [ ] Live editing sync
- [ ] Conflict resolution

#### 3.5 Public Specs Gallery
```
Priority: MEDIUM
Effort: 2-3 days
```
- [ ] Browse public specs
- [ ] Search và filter
- [ ] Categories/tags
- [ ] Featured specs
- [ ] "Fork this spec" action

**Phase 3 Deliverables:**
- Shareable public links
- Permission system
- Comments
- Public gallery

---

### Phase 4: Template Marketplace
**Timeline**: 2-3 tuần  
**Goal**: Community-driven growth

#### 4.1 Template Publishing
```
Priority: HIGH
Effort: 3-4 days
```
- [ ] Publish template workflow
- [ ] Template metadata (name, description, tags, preview)
- [ ] Version control for templates
- [ ] Update published templates

#### 4.2 Marketplace UI
```
Priority: HIGH
Effort: 3-4 days
```
- [ ] Browse templates grid
- [ ] Categories: API Docs, Changelogs, Database Schema, etc.
- [ ] Search với filters
- [ ] Sorting: Popular, Recent, Trending
- [ ] Template detail page với live preview

#### 4.3 Template Interactions
```
Priority: MEDIUM
Effort: 2-3 days
```
- [ ] Install/use template
- [ ] Star/favorite templates
- [ ] Download counter
- [ ] User ratings & reviews
- [ ] Report inappropriate content

#### 4.4 Creator Profiles
```
Priority: LOW
Effort: 2 days
```
- [ ] Public profile page
- [ ] Templates by creator
- [ ] Follower system (optional)
- [ ] Creator verification badge

**Phase 4 Deliverables:**
- Template marketplace
- Publishing workflow
- Categories và search
- Creator profiles

---

### Phase 5: Monetization
**Timeline**: 1-2 tuần  
**Goal**: Revenue generation

#### 5.1 Pricing Tiers
```
Priority: HIGH
Effort: 1-2 days
```

| Tier | Price | Limits |
|------|-------|--------|
| **Free** | $0 | 5 conversions/month, 1 template, no sharing |
| **Pro** | $15/mo | Unlimited conversions, 10 templates, sharing, API (100 req/mo) |
| **Team** | $49/mo | Everything + 10 seats, SSO, priority support |
| **Enterprise** | Custom | Self-hosted, unlimited, SLA |

#### 5.2 Stripe Integration
```
Priority: HIGH
Effort: 3-4 days
```
- [ ] Stripe Checkout integration
- [ ] Customer portal (manage subscription)
- [ ] Webhook handlers (subscription events)
- [ ] Usage metering for API
- [ ] Invoice generation

#### 5.3 Usage Tracking
```
Priority: HIGH
Effort: 2 days
```
- [ ] Track conversions per user
- [ ] API request counting
- [ ] Storage usage
- [ ] Upgrade prompts when limit reached

#### 5.4 Premium Features Gating
```
Priority: HIGH
Effort: 2 days
```
- [ ] Feature flags per tier
- [ ] Graceful upgrade prompts
- [ ] Trial period (14 days Pro)
- [ ] Grandfathering existing users

**Phase 5 Deliverables:**
- Stripe payments
- Subscription tiers
- Usage tracking
- Premium feature gating

---

### Phase 6: API & Integrations
**Timeline**: 2-3 tuần  
**Goal**: Developer adoption

#### 6.1 Public API
```
Priority: HIGH
Effort: 4-5 days
```
```bash
# Convert paste data
POST /api/v1/convert/paste
Authorization: Bearer <api_key>
Content-Type: application/json

{
  "data": "col1\tcol2\nval1\tval2",
  "template": "default",
  "format": "markdown"
}

# Convert file
POST /api/v1/convert/file
Authorization: Bearer <api_key>
Content-Type: multipart/form-data

# List templates
GET /api/v1/templates

# Get document
GET /api/v1/documents/:id
```

#### 6.2 API Key Management
```
Priority: HIGH
Effort: 2 days
```
- [ ] Generate API keys
- [ ] Key permissions (read, write)
- [ ] Rate limiting per key
- [ ] Usage analytics per key
- [ ] Revoke keys

#### 6.3 Webhooks
```
Priority: MEDIUM
Effort: 2-3 days
```
- [ ] Webhook configuration UI
- [ ] Events: document.created, document.updated, conversion.completed
- [ ] Retry logic
- [ ] Webhook logs

#### 6.4 Integrations
```
Priority: MEDIUM
Effort: 1 week
```
- [ ] **GitHub**: Create issues/PRs from specs
- [ ] **Slack**: Post converted specs to channels
- [ ] **Notion**: Export as Notion pages
- [ ] **Zapier**: No-code automation

**Phase 6 Deliverables:**
- REST API với documentation
- API key management
- Webhooks
- GitHub/Slack integrations

---

## 📅 Timeline Summary

```
Week 1-2:   Phase 1 - UX Polish
Week 3-4:   Phase 2 - Authentication
Week 5-7:   Phase 3 - Sharing
Week 8-10:  Phase 4 - Marketplace
Week 11-12: Phase 5 - Monetization
Week 13-15: Phase 6 - API

Launch: Week 16
```

---

## 🚀 Launch Strategy

### Pre-Launch (Week 14-15)
- [ ] Landing page với waitlist
- [ ] Social media teasers
- [ ] Reach out to beta testers
- [ ] Prepare ProductHunt assets
- [ ] Write launch blog post

### Launch Day (Week 16)
- [ ] ProductHunt submission
- [ ] HackerNews post
- [ ] Twitter/X announcement
- [ ] Dev.to article
- [ ] Reddit posts (r/webdev, r/programming)

### Post-Launch
- [ ] Respond to all comments
- [ ] Fix critical bugs immediately
- [ ] Collect feedback
- [ ] Iterate based on usage data

---

## 📊 Success Metrics

### Phase 1-2 (Foundation)
- Time to first conversion < 30 seconds
- 0 critical bugs
- Page load < 2s

### Phase 3-4 (Growth)
- 100 public specs created
- 50 templates in marketplace
- 20% users share at least 1 doc

### Phase 5-6 (Revenue)
- 5% free-to-paid conversion
- $1000 MRR within 3 months
- 100 API users

---

## 🛠 Tech Stack Additions

| Layer | Current | Add |
|-------|---------|-----|
| Auth | - | Clerk or NextAuth |
| Database | - | PostgreSQL + Prisma |
| Payments | - | Stripe |
| Real-time | - | Liveblocks (optional) |
| Analytics | - | PostHog or Mixpanel |
| Email | - | Resend or SendGrid |
| Storage | - | Cloudflare R2 or S3 |
| Search | - | Algolia or Meilisearch |

---

## 🎨 Design System Updates

### New Components Needed
- [ ] Command Palette (cmdk)
- [ ] User Avatar with dropdown
- [ ] Pricing cards
- [ ] Template cards
- [ ] Share modal
- [ ] Comments thread
- [ ] API key display
- [ ] Usage progress bar

### Animation Tokens
```css
--ease-out-expo: cubic-bezier(0.16, 1, 0.3, 1);
--duration-fast: 150ms;
--duration-normal: 250ms;
--duration-slow: 400ms;
```

---

## 📝 Notes

### Risks & Mitigations
| Risk | Mitigation |
|------|------------|
| Scope creep | Strict phase boundaries |
| Auth complexity | Use Clerk (managed auth) |
| Payment edge cases | Stripe handles most |
| Real-time bugs | Delay to Phase 4+ |

### Quick Wins
1. Command palette (⌘K) - High impact, medium effort
2. Keyboard shortcuts - Low effort, high delight
3. Public sharing links - Enables viral growth
4. GitHub OAuth - Easy signup friction reduction

---

*Last updated: 2026-01-31*
