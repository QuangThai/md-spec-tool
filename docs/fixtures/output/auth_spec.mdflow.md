---
name: "Auth Test Cases"
version: "1.0"
generated_at: "2026-01-28"
_inputs:
  _feature_filter:
    type: text
    description: "Filter by feature name"
---

# Auth Test Cases

This specification contains 5 test cases for authentication flows.

## Register

### AUTH-001: Register with valid email/password

**Priority:** High | **Type:** Positive

**Preconditions:**
- User not registered

**Steps:**
1. Open Register page
2. Enter valid email
3. Enter strong password
4. Submit

**Test Data:**
- email=user1@test.com; password=Str0ng!Pass

**Expected Result:**
- Account created and user logged in

**API/Endpoint:** `/auth/register`

---

### AUTH-002: Register with existing email

**Priority:** High | **Type:** Negative

**Preconditions:**
- Email already registered

**Steps:**
1. Open Register page
2. Enter existing email
3. Enter valid password
4. Submit

**Test Data:**
- email=existing@test.com; password=Str0ng!Pass

**Expected Result:**
- Error shown: email already exists

**API/Endpoint:** `/auth/register`

---

### AUTH-003: Password strength validation

**Priority:** Medium | **Type:** Negative

**Preconditions:**
- User not registered

**Steps:**
1. Open Register page
2. Enter weak password
3. Submit

**Test Data:**
- email=user2@test.com; password=12345

**Expected Result:**
- Inline validation for password policy

**API/Endpoint:** `/auth/register`

---

## Login

### AUTH-004: Login with valid credentials

**Priority:** High | **Type:** Positive

**Preconditions:**
- User exists and active

**Steps:**
1. Open Login page
2. Enter valid email/password
3. Submit

**Test Data:**
- email=user1@test.com; password=Str0ng!Pass

**Expected Result:**
- User logged in; session created

**API/Endpoint:** `/auth/login`

---

### AUTH-005: Login with invalid password

**Priority:** High | **Type:** Negative

**Preconditions:**
- User exists and active

**Steps:**
1. Open Login page
2. Enter valid email
3. Enter wrong password
4. Submit

**Test Data:**
- email=user1@test.com; password=WrongPass

**Expected Result:**
- Error shown: invalid credentials

**API/Endpoint:** `/auth/login`

---
