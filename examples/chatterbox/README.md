# chatterbox

A simple chat server that illustrates using [EventStream](https://github.com/fullstorydev/go/tree/master/eventstream) with gRPC.

## Install

### Remote

```bash
go install github.com/fullstorydev/go/examples/chatterbox/cmd/chatterbox@master
```

### Local

```bash
git clone https://github.com/fullstorydev/go.git
cd go/examples/chatterbox
go install ./cmd/chatterbox
```

## Run

```bash
chatterbox server # run the server
chatterbox listen # run a listener
chatterbox client # run a client
chatterbox client # run a client
```

You can run only one server, but as many clients or listeners as you want
in different shells.
