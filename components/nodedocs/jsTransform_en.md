# JS Transform

Transforms the message with JavaScript (goja). One visual port; Success / Failure edges.

## Purpose

- Mutate msg / metadata / msgType
- Script errors can follow the Failure edge

## Configuration

| Field | Notes |
| --- | --- |
| jsScript | Transform **function body** (no function wrapper) |
| debugValue | Test msg for panel **Run** only |

### Script contract

~~~js
return { msg, metadata, msgType, dataType };
~~~

Available: msg, metadata, msgType, dataType, global, vars.

## Wiring

First outgoing edge defaults to Success, second to Failure; with one edge you can toggle the result on the edge.
