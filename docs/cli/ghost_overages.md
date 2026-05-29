---
title: "ghost overages"
slug: "ghost_overages"
description: "CLI reference for ghost overages"
---

## ghost overages

Manage compute overages

### Synopsis

Manage compute overage billing for your Ghost space.

By default, each space gets an included free compute allowance each calendar
month; when it is used up, all standard databases in the space are auto-paused
until the next month. Enabling overages lets you pay for compute beyond the
free allowance, optionally capped at a monthly compute-hour limit you choose.

Run 'ghost pricing' to see the included free allowance and the per-hour
overage rate.

### Options

```
  -h, --help   help for overages
```

### Options inherited from parent commands

```
      --analytics           enable/disable usage analytics (default true)
      --color               enable colored output (default true)
      --config-dir string   config directory (default "~/.config/ghost")
      --version-check       check for updates (default true)
```

### SEE ALSO

* [ghost](ghost.md)	 - CLI for managing Postgres databases
* [ghost overages disable](ghost_overages_disable.md)	 - Disable compute overages
* [ghost overages enable](ghost_overages_enable.md)	 - Enable compute overages
