# Inject

Entry node that starts a flow on a schedule or by manual run.

## Purpose

- Flow start: **no incoming edges**, outgoing only
- Configure an initial JSON message body
- Useful for debugging, cron-like jobs, and polling flows

## Configuration

| Field | Notes |
| --- | --- |
| Payload / message | Used as the initial msg |
| Debug | Panel **Run** can start from this node |

## Wiring

Connect from the right-side output; no inputs allowed.
