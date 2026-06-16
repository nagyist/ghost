---
title: "ghost invite accept"
slug: "ghost_invite_accept"
description: "CLI reference for ghost invite accept"
---

## ghost invite accept

Accept an invitation

### Synopsis

Accept an invitation you've received, joining the space.

The space is identified by its ID, as shown by 'ghost invite received'. After
joining, you can switch the CLI's current space to the new space. By default
you'll be prompted; use --switch or --switch=false to decide without a prompt.

```
ghost invite accept <space-id> [flags]
```

### Examples

```
  # Accept an invitation
  ghost invite accept x9y8z7w6v5

  # Accept and immediately switch to the new space
  ghost invite accept x9y8z7w6v5 --switch
```

### Options

```
  -h, --help     help for accept
      --switch   Switch the current space to the joined space
```

### Options inherited from parent commands

```
      --analytics           enable/disable usage analytics (default true)
      --color               enable colored output (default true)
      --config-dir string   config directory (default "~/.config/ghost")
      --version-check       check for updates (default true)
```

### SEE ALSO

* [ghost invite](ghost_invite.md)	 - Invite a user to the current space
