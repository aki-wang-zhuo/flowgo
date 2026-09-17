# HTTP Client

Sends an outbound HTTP request; results can follow Success / Failure.

## Purpose

- Call external REST APIs / webhooks
- Failures or unexpected status can use Failure

## Configuration

| Field | Notes |
| --- | --- |
| URL / method | Target and verb |
| Headers / body | Request headers and body (templates OK) |
| Timeouts | Tune per deployment |

## Wiring

Same pattern as JS Transform: Success / Failure share one visual port.
