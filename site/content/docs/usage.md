---
title: "Usage"
description: "How to use tgdash"
weight: 2
---

## Basic usage

Pipe any Terragrunt command into tgdash:

```bash
terragrunt run --all plan 2>&1 | tgdash
```

```bash
terragrunt run --all apply 2>&1 | tgdash
```

Any command that produces `[unit/path]`-prefixed
output works:

```bash
terragrunt run --all validate 2>&1 | tgdash
```

## What you see

tgdash parses the streaming output and shows:

- **Unit status** - Each Terragrunt unit with its
  current state (waiting, running, done, error)
- **Plan summaries** - Resources to add, change, and
  destroy per unit
- **Apply results** - Final resource counts after apply
- **Errors** - Detected errors highlighted and
  filterable
- **Time estimates** - Based on historical run data

## Views

tgdash has two views you can switch between with
`Tab`:

### Dashboard view

The default view. Shows each unit with full detail
including status, plan summary, output lines, and
timing.

### List view

A compact overview showing all units in a table format.
Good for large repos with many units where you want to
see overall progress at a glance.

## Error filtering

Press `e` to toggle error filtering. When active, only
units with errors are shown, making it easy to find
and diagnose failures in large runs.

## Search

Press `/` to enter search mode. Type to filter units
by path. Press `Escape` to clear the search.
