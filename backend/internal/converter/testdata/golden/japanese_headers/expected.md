---
name: "Specification"
version: "1.0"
generated_at: "2026-02-28"
type: "specification"
---

## Summary

| Metric | Value |
|--------|-------|
| Total Items | 2 |
| Mapped Columns | 6 |
| Extra Columns | 0 |
| Avg Confidence | 95% |

# Converted Specification

## Specifications

### Uncategorized

#### TC-001: ログイン成功

| Field | Value |
|-------|-------|
| id | TC-001 |
| status | 合格 |

**Description:**
ログイン成功

**Precondition:**
ログインページ表示

**Steps:**
1. ユーザー名入力\n2. パスワード入力\n3. ログインクリック

**Expected Result:**
ダッシュボード表示


#### TC-002: ログイン失敗

| Field | Value |
|-------|-------|
| id | TC-002 |
| status | 未テスト |

**Description:**
ログイン失敗

**Precondition:**
ログインページ表示

**Steps:**
1. ユーザー名入力\n2. 誤パスワード入力

**Expected Result:**
エラーメッセージ表示


