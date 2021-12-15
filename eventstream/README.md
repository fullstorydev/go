# eventstream

An `EventStream` is a fast, efficient way to publish and subscribe to events within a Go process.

The primary difference between `EventStream` and Go's native `channel` is that with `EventStream`,
all subscribers see _every_ event.  (With Go native channels, only one concurrent reader can get
any particular event).

Another difference from Go channels is that multiple concurrent publishers must be externally
synchronized.  (With Go channels, multiple concurrent publishers are internally synchronized.)
We chose to avoid internal synchronization on `Publish` so that the common case of a single
publisher would be maximally efficient.

Event values passed through `EventStream` are opaque to `EventStream`. You must ensure these
values are either effectively immutable or else correctly synchronized, since multiple
subscribers will receive references to the same value.

## How it works

`EventStream` is an immutable, append-only linked list of nodes.  The publishing side keeps
a reference only to the end of the list-- a single tail node that is not yet ready.  During
a publish operation, a new tail is created and linked from the current tail, and then the
current tail is "made ready" by closing its channel, which signals all subscribers so they
can read the next published value.

Subscribers keep a reference to the next node in the linked list that they need to consume.
As subscribers traverse the linked list, the head of the list becomes unreferenced and
available for garbage collection (GC).

In practice, `EventStream` uses an internal buffer of nodes to avoid frequent allocations,
so GC of older events may be delayed until a sufficient number of new events have passed through.
If your events pin a lot of memory, you might want to use a small buffer size so that
nodes can be collected more frequently.

## Use cases

Use this wherever you might have used a Go channel, but you need to multiple subscribers to each
receive all events.

- Publish a shared stream of events to all connected clients (for an example, see [BWAMP](https://bwamp.me))
- Synchronize a shared data model across services using gRPC streams.

## Examples

See [chatterbox](../examples/chatterbox) for a simple chat client implemented using gRPC streams with `EventStream`.
