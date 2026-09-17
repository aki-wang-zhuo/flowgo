# HTTP Response

Flow exit that writes the current message back as the HTTP response.

## Purpose

- Pair with **HTTP Endpoint** to finish a request/response
- **No outgoing edges**; incoming only

## Configuration

Status code and body (templates supported). Empty body follows node defaults.

## Wiring

Incoming edges only; do not connect outwards.
