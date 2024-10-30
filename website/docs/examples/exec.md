---
# SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
sidebar_position: 2
---

# Execute code

Inside the network namespaces / hosts


### `exec.Cmd` API

```go
// Get a exec.Cmd-like struct
cmd := host1.Command("ping", "host2")
out, err := cmd.CombinedOutput()

// Directly run a simple process synchronously
cmd, err := host1.Run("ping", "host2")

// Directly start asynchronously
cmd, err := host1.Start("ping", "host2")

time.Sleep(5 * time.Second)

err = cmd.Process.Signal(os.Interrupt)
cmd.Wait()
```

The `gont.Node` type implements an API similar to the one provided by Go's
`exec` package.


### Pass options

```go
import copt "github.com/cunicu/gont/v2/pkg/options/cmd"

outb := &bytes.Buffer{}

cmd := host1.Command("ping", "1.1.1.1",
  copt.DisableASLR(true),
  copt.Dir("/custom/working/dir"),
  copt.EnvVar("DEBUG", "1"),
  copt.Stdin(...), // pass any io.Reader
  copt.Stdout(outb), // pass any io.Writer (can be repeated)
  copt.Stderr(...), // pass any io.Writer (can be repeated)
)

print(outb.String())
```

### Pass non-string arguments

```go
ip := net.ParseIP("1.1.1.1")

cmd := host1.Command("ping", "-c", 10, "-i", 0.1, ip)
```

### Go functions

```go
host1.RunFunc(func() {
  r := http.Get("http://host2:8080")
  io.Copy(os.Stdout, r.Body)
})
```

Call a function inside a network namespace of host `host1` but still
in the same process so you can use channels and access global variables.

:::warning
Spawning Goroutines from within the callback is only indirectly supported:

```go
host1.RunFunc(func() {
  go host1.RunFunc(func() { ... })
})
```
:::

### Go packages

```go
cmd, err := host1.RunGo("test/prog.go", "arg1")
```
