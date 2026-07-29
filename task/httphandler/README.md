# Tasks

A task is a single unit of asynchronous work - such as transcoding a file or
probing a stream - tracked under a unique uuid. Tasks are created and
started elsewhere in the system; this API lets you list them, inspect an
individual task's progress and result, and cancel one that's still running.

Each task moves through a small set of states, reported as `state` in its
status:

- `not_started` - registered, but not yet running.
- `running` - currently executing.
- `done` - finished successfully.
- `error` - finished, but returned an error.
- `cancelled` - cancellation was requested, and the task has since stopped.

A task starts in `not_started`, moves to `running` once started, and ends
in exactly one of `done`, `error`, or `cancelled`. Cancelling a task only
requests that it stop - it stays in `running` until it actually does, at
which point it settles into `cancelled`. A task's final status remains
available through this API even after it finishes.

## GET /task

Returns a list of tasks tracked by the manager, in the order they were
added. Filter by `state`, and page through results with `offset` and
`limit`.

## GET /task/event

Streams every task event as it happens - a task being added, started,
reporting progress or a result, cancelled, finishing, or removed - as
[server-sent events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events).
Each event's `event:` field is the event name (e.g. `started`, `progress`,
`finished`) and its `data:` field is the task's status at that moment, JSON-
encoded the same way as `GET /task/{uuid}`.

Filter by `uuid` to only stream events for one task, and/or by `event`
(repeatable) to only stream specific event names, e.g.
`?event=started&event=finished`.

The stream stays open until the client disconnects.

## GET /task/{uuid}

Returns the current status of the task with the given uuid.

## DELETE /task/{uuid}

Cancels the task with the given uuid. It's a no-op if the task has already
finished or was never started. Returns the task's status after the
cancellation request.
