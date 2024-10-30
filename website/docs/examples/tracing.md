---
# SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
sidebar_position: 6
---

# Tracing

User defined events, Log messages

## A trace event

Gont supports collecting trace events from all processes running in a
distributed system which can carry the following information. Gont
orders trace events by time and saves them to different destinations for
analysis.

```go
type Event struct {
  Timestamp time.Time // Timestamp when the event occurred
  Type      string    // Either: 'log', ́'trace', 'break' & ́ watchpoint'
  Level     uint8     // Log level
  Message   string    // A human readable description
  Source    string    // Logger name
  PID       int 
  Function  string
  File      string
  Line      int
  Args      []any
  Data      any       // User defined data
}
```

### Sink trace events into

-   JSON files
-   Go channels
-   Go callbacks
-   Packet captures

## Create a tracer

```go
import "cunicu.li/gont/v2/trace"
import topt "github.com/cunicu/gont/v2/options/trace"

c := gont.NewCapture(...)
f, _ := os.OpenFile(...)
ch := make(chan trace.Event)

t := gont.NewTracer(
  topt.ToFile(f)
  topt.ToFilename("trace.log"),
  topt.ToChannel(ch),
  topt.ToCapture(c),
  topt.Callback(func(e trace.Event) { ... }))

t.Start()
```

## Attach the tracer

Trace all processes started by nodes of this network

```go
network, _ := gont.NewNetwork("", t)
```

Trace all processes started by a node

```go
host1 := network.NewHost("host1", t)
```

Trace a single process

```go
host1.RunGo("test/main.go", t)
```

## Trace with the `trace` package

```go
import "cunicu.li/gont/v2/pkg/trace"

someData := map[string]string{"Hello": "World"}
count := 42

trace.Start(0)

trace.PrintfWithData(someData, "Count is: %d", count)
trace.Print("Another message")

trace.Stop()
```

Works from:

-  Gont process itself
-  Any process spawned via Gont's
   `Host.{Command,Run,RunGo,Start,StartGo}(...)` functions


## Trace via `slog` structured logging package

```go
import "log/slog"
import "cunicu.li/gont/v2/pkg/trace"

// Create a slog handler which emits trace events
handler := trace.NewTraceHandler(slog.HandlerOptions{})

// Add the tracing option which emits a trace event for each log message
logger := slog.New(handler)
```

Each log message emits a trace event which includes the log message, filename, line number as well function name and more.
Any fields passed to to zap structured logger are included in the `Data` field of the `Event` structure.

## Trace via `go.uber.org/zap` logging package

```go
import "go.uber.org/zap"
import "cunicu.li/gont/v2/pkg/trace"

// Add the tracing option which emits a trace event for each log message
logger := zap.NewDevelopment(trace.Log())

// Add the caller info which gets also included in the trace event
logger = logger.WithOptions(zap.AddCaller())

// Give the logger some name which is added as the Source field to the trace event
logger = logger.Named("my-test-logger")
```

Each log message emits a trace event which includes the log message, filename, line number as well function name and more.
Any fields passed to to zap structured logger are included in the `Data` field of the `Event` structure.
