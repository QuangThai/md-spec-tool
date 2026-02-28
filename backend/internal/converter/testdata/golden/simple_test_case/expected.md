---
name: "Specification"
version: "1.0"
generated_at: "2026-02-28"
type: "specification"
---

## Summary

| Metric | Value |
|--------|-------|
| Total Items | 3 |
| Mapped Columns | 6 |
| Extra Columns | 0 |
| Avg Confidence | 95% |

# Converted Specification

## Specifications

### Uncategorized

#### TC-001: Login with valid credentials

| Field | Value |
|-------|-------|
| id | TC-001 |
| status | Pass |

**Description:**
Login with valid credentials

**Precondition:**
User is on login page

**Steps:**
1. Enter username\n2. Enter password\n3. Click Login

**Expected Result:**
Dashboard is displayed


#### TC-002: Login with invalid password

| Field | Value |
|-------|-------|
| id | TC-002 |
| status | Not tested |

**Description:**
Login with invalid password

**Precondition:**
User is on login page

**Steps:**
1. Enter username\n2. Enter wrong password\n3. Click Login

**Expected Result:**
Error message shown


#### TC-003: Forgot password flow

| Field | Value |
|-------|-------|
| id | TC-003 |
| status | Pass |

**Description:**
Forgot password flow

**Precondition:**
User is on login page

**Steps:**
1. Click Forgot Password\n2. Enter email\n3. Submit

**Expected Result:**
Reset email is sent


