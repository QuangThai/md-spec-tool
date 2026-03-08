# Documentation Update: Stream Conversion API

[agent: content]

## Scope
Draft for API documentation. **Do not modify `docs/` directly** — create a dev ticket if these changes should be merged into the codebase.

## Stream Conversion Endpoint

**Endpoint**: `POST /api/v1/mdflow/convert/stream`  
**Content-Type**: `application/json`  
**Response**: Server-Sent Events (SSE)

### Request Body (updated)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `paste_text` | string | Yes | — | The pasted text to convert |
| `template` | string | No | `"spec"` | Template name |
| `format` | string | No | `"spec"` | Output format: `"spec"` or `"table"` |
| `include_metadata` | boolean | No | `true` | Whether to include front matter and summary in output |
| `number_rows` | boolean | No | `false` | Whether to add a row number column |

### Example

```json
{
  "paste_text": "Feature\tScenario\tExpected\nLogin\tHappy path\tUser redirected",
  "template": "spec",
  "format": "spec",
  "include_metadata": false,
  "number_rows": true
}
```

### Notes
- `include_metadata` and `number_rows` follow the same semantics as the non-stream paste convert API (`POST /api/v1/mdflow/paste`).
- When omitted, defaults are applied: `include_metadata: true`, `number_rows: false`.

## Suggested Location
Add or update in `docs/api.md` or equivalent API reference. If no such file exists, consider creating one as a dev ticket.
