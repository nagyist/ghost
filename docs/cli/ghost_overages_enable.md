---
title: "ghost overages enable"
slug: "ghost_overages_enable"
description: "CLI reference for ghost overages enable"
---

## ghost overages enable

Enable compute overages

### Synopsis

Enable compute overage billing for your Ghost space.

Once enabled, you will be charged for compute beyond the included free
allowance each calendar month (see 'ghost pricing'). By default there is no
monthly cap on overage usage — pass --limit <hours> to set one. When the cap
is reached, standard databases in the space auto-pause until the next month.

A payment method must be on file before overages can be enabled. Run
'ghost payment add' to add one.

This command is also used to update an existing overages-enabled space:
re-run it with a different --limit value (or with no flag, to switch to
no-limit mode).

```
ghost overages enable [flags]
```

### Examples

```
  # Enable overages with a 200-hour monthly cap
  ghost overages enable --limit 200

  # Enable overages with no monthly cap (charges have no upper bound)
  ghost overages enable

  # Skip the no-limit confirmation prompt
  ghost overages enable --confirm
```

### Options

```
      --confirm     Skip the no-limit confirmation prompt
  -h, --help        help for enable
      --limit int   Monthly compute cap in hours (must exceed the included free allowance). Omit for no cap.
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
