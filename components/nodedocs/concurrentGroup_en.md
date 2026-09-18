# Concurrent Group

Run named branches **in parallel** inside a group frame, then join by all-complete or any-complete and leave via Success / Failure.

## Use cases

- Fan-out to multiple downstream steps (HTTP, checks, etc.)
- Merge results into a single message with a `branches` map

## Frame anchors

| Anchor | Description |
| ------ | ----------- |
| Left in | External entry into the group |
| Fork | Fan-out: draw edges, pick a branch name, connect to each branch start |
| Success join | Successful branch paths reconnect here |
| Fail join | Failure paths reconnect here |
| Right out | Exit after join (one visual port; edge labels Success / Failure) |

## Configuration

| Field | Description |
| ----- | ----------- |
| branches | Branch name list |
| completeMode | `all` or `any` |
| cancelOthersOnAny | Cancel remaining branches when any-complete is met |
| timeoutSec | Timeout seconds; `0` = no timeout |

## Downstream message

After join, message body looks like:

```json
{
  "branches": {
    "branch1": { "ok": true, "msg": ... },
    "branch2": { "ok": false, "error": "..." }
  }
}
```
