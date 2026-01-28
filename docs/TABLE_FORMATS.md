# Supported Table Formats

This document describes the table formats supported by the MDFlow Converter.

## Format 1: Row-per-Requirement (Default)

The most common format where each row represents a single requirement or test case.

```
| Feature | Description | Expected |
|---------|-------------|----------|
| Login   | Enter creds | Success  |
| Logout  | Click btn   | Redirect |
```

**Detection:** First row contains recognizable header terms, subsequent rows are data.

## Format 2: Key-Value Pairs

Single requirement spread across multiple rows as key-value pairs.

```
| Field       | Value                    |
|-------------|--------------------------|
| Feature     | User Authentication      |
| Description | Login flow               |
| Expected    | User is authenticated    |
```

**Detection:** Two columns where first column contains field names.

## Format 3: Sectioned Tables

Multiple logical groups separated by section headers.

```
| Authentication Module           |
|---------------------------------|
| Feature | Description | Expected |
| Login   | ...         | ...      |
|                                 |
| User Management Module          |
|---------------------------------|
| Feature | Description | Expected |
| Create  | ...         | ...      |
```

**Detection:** Merged cells or single-column rows acting as section headers.

## Header Synonyms

The converter uses fuzzy matching to recognize common column names:

| Canonical Field | Accepted Synonyms |
|-----------------|-------------------|
| `feature` | feature, req, requirement, story, user story, task, title, name, test case, tc |
| `instructions` | instructions, description, steps, test steps, action, scenario |
| `inputs` | inputs, input, test data, data, parameters |
| `expected` | expected, expected output, expected result, acceptance, acceptance criteria |
| `precondition` | precondition, preconditions, pre-condition, pre, given |
| `priority` | priority, prio, p, severity |
| `notes` | notes, note, comments, remarks |
| `endpoint` | endpoint, api, url, route |
| `type` | type, category, test type |
| `status` | status, state |

## Paste Format Support

### TSV (Tab-Separated Values)
Default format when copying from Google Sheets or Excel.

### CSV (Comma-Separated Values)
Auto-detected when tabs are not found.

### Handling Multi-line Cells
Cells containing newlines are preserved and converted to markdown lists in the output.
