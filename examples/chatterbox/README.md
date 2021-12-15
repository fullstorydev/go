# chatterbox

A simple chat server that illustrates using [EventStream](https://github.com/fullstorydev/go/tree/master/eventstream)
with gRPC.

## Install

### Remote

```bash
go install github.com/fullstorydev/go/examples/chatterbox/...@latest
```

### Local

```bash
git clone https://github.com/fullstorydev/go.git
cd go/examples/chatterbox
go install ./cmd/chatterbox
```

## Run

```bash
chatterbox server  # run the server
chatterbox monitor # run a monitor (in a different terminal)
chatterbox client  # run a client  (in a different terminal)
chatterbox client  # run a client  (in a different terminal)
```

You can run only one server, but as many clients or monitors as you want in different terminals. A client participates
in chat, but a monitor just passively listens.
