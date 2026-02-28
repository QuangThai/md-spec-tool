---
name: "Specification"
version: "1.0"
generated_at: "2026-02-28"
type: "specification"
---

## Summary

| Metric | Value |
|--------|-------|
| Total Items | 4 |
| Mapped Columns | 6 |
| Extra Columns | 0 |
| Avg Confidence | 95% |

# Converted Specification

## Specifications

### Uncategorized

#### 

**Description:**
List all users

**API Details:**
- **Endpoint**: `/api/v1/users`
- **Method**: GET
- **Status Code**: 200
- **Parameters:**
page, limit, sort
- **Response:**
{users: [], total: int}


#### 

**Description:**
Create new user

**API Details:**
- **Endpoint**: `/api/v1/users`
- **Method**: POST
- **Status Code**: 201
- **Parameters:**
name, email, role
- **Response:**
{id: string, created: bool}


#### 

**Description:**
Get user by ID

**API Details:**
- **Endpoint**: `/api/v1/users/:id`
- **Method**: GET
- **Status Code**: 200
- **Parameters:**
id (path)
- **Response:**
{user: object}


#### 

**Description:**
Delete user

**API Details:**
- **Endpoint**: `/api/v1/users/:id`
- **Method**: DELETE
- **Status Code**: 204
- **Parameters:**
id (path)
- **Response:**
{deleted: bool}


