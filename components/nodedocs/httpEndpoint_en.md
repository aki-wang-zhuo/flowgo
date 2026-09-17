# HTTP Endpoint

Ingress node that feeds external HTTP requests into a flow.

## Purpose

- Listens on the server HTTP endpoint port
- Each route maps to one outgoing edge (a route cannot be shared)
- Typical for webhooks, public APIs, and form callbacks

## Configuration

| Field | Notes |
| --- | --- |
| Routes | Method + path; at most one edge per path |
| TLS | Enable when required by deployment |

## Wiring

Drag from the right output; pick the route when connecting. The path label sits on the edge midpoint.
