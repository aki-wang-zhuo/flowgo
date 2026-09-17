# Global Variables

Declare flow-scoped variables available as `global.xx`. No input or output ports.

## Purpose

- Centralize API bases, flags, limits, etc.
- Templates `${global.name}`; expressions / JS `global.name`

## Configuration

| Field | Notes |
| --- | --- |
| Variable list | Dynamic rows: name / type / value |
| type | string / number / boolean / json |

## Limits

At most **one** such node per flow; cannot be the entry node; no edges.
