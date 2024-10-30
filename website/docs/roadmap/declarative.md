---
# SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
---

# Declarative Configuration

Many use-cases could profit from a declarative description of the desired network topology and configuration.

## Example

```yaml
nodes:
  sw1:
    type: Switch

  h1:
    type: Host

  h2:
    type: Host
    run:
    - command: ping
      args: [ 1.1.1.1 ]
      cgroup:
        MaxTasks: 12


links:
- left:
    node: h1
    interface:
      address: 1.1.1.1/24

  right:
    to: sw1

- left:
    node: h2
    interface:
      address: 1.1.1.2/24

  right:
    to: sw1
```