---
# SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
---

# Namespaces

What are Linux namespaces? They..

-   Partition kernel resources from **a process-perspective.**
-   Power most of Linux containerization tools.
-   Appear in many Linux subsystems.
-   Are used by many sandboxing solutions in browsers.


## Available Linux namespaces

-   **mnt:** Mounts
-   **pid:** Process IDs
-   **net:** Networking
-   **ipc:** Interprocess Communication
-   **uts:** Unix Timesharing (hostname)
-   **user:** User identification & privileges
-   **cgroup:** Process Control Groups
-   **time:** System time


## But what exactly is a namespace?

-   Can be identified by a file descriptor
-   A namespace can have multiple processes assigned to it
-   It lives as long as there is still at least one remaining process
-   Child processes inherit parent namespaces


## How do I create a namespace?

-   [unshare(2)](https://man7.org/linux/man-pages/man2/unshare.2.html)
-   [clone(2)](https://man7.org/linux/man-pages/man2/clone.2.html)


## unshare(2)

```go
func main() {
  err := syscall.Unshare(syscall.CLONE_NEWNS);
}
```


## How do I create a namespace?

```go
static int child(void *arg) {
  struct utsname uts;

  sethostname(arg, "ernie")

  uname(&uts)
  printf("nodename in child:  %s\n", uts.nodename);

  return 0;
}

int main() {
  struct utsname uts;

  // Allocate stack for child
  char *stack = malloc(STACK_SIZE);
  if (stack == NULL)
    return -1;

  // Start new kernel task in new UTS namespace
  pid_t child_pid = clone(child, stack + STACK_SIZE, CLONE_NEWUTS | SIGCHLD, NULL);

  // Output hostname
  uname(&uts)
  printf("nodename in parent:  %s\n", uts.nodename);
}
```


## How can I share a namespace with other processes?

-   By forking with
    [clone(2)](https://man7.org/linux/man-pages/man2/clone.2.html)
-   By passing file descriptor and
    [setns(2)](https://man7.org/linux/man-pages/man2/setns.2.html)

## Joining namespace of another process: `/proc`

You can join the namespace of another process by using [/proc/\{pid\}/ns/\*](https://man7.org/linux/man-pages/man5/proc.5.html)

```go
fd := syscall.Open("/proc/1234/ns/uts", syscall.O_RDONLY);
err := unix.Setns(fd, syscall.CLONE_NEWUTS);
```

**Note:** Can only set a single namespace per `netns(2)` invocation.

## Joining namespace of another process: `pidfd_open`

You can join the namespace of another process by using [pidfd\_open(2)](https://man7.org/linux/man-pages/man2/pidfd_open.2.html)

```go
pid_t pid = 1234;
int fd = pidfd_open(pid, 0);

setns(fd, CLONE_NEWUTS | CLONE_NEWNET);
 ```

**Note:** Can only set a multiple namespaces per `netns(2)` invocation.

## Persisting namespaces

```go
err := syscall.Mount("/proc/self/ns/uts",
                     "/home/acs/my_uts_namespace", "", syscall.MS_BIND, nil);
```

And in another process...

```go
fd := syscall.Open("/home/acs/my_uts_namespace", syscall.O_RDONLY);
err := unix.Setns(fd, syscall.CLONE_NEWUTS);
```
