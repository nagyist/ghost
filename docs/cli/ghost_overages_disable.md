---
title: "ghost overages disable"
slug: "ghost_overages_disable"
description: "CLI reference for ghost overages disable"
---

## ghost overages disable

Disable compute overages

### Synopsis

Disable compute overage billing for your Ghost space.

After disabling, the compute limit resets to the included free allowance.
If your current month-to-date usage is already above that, your non-dedicated
databases will be paused.

```
ghost overages disable [flags]
```

### Examples

```
  # Disable overages (prompts for confirmation)
  ghost overages disable

  # Disable without confirmation prompt
  ghost overages disable --confirm
```

### Options

```
      --confirm   Skip confirmation prompt
  -h, --help      help for disable
```

### Options inherited from parent commands

```
      --analytics           enable/disable usage analytics (default true)
      --color               enable colored output (default true)
      --config-dir string   config directory (default "~/.config/ghost")
      --version-check       check for updates (default true)
```

### SEE ALSO

* [ghost overages](ghost_overages.md)	 - Manage compute overages
