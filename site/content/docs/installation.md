---
title: "Installation"
description: "How to install tgdash"
weight: 1
---

## From source

If you have Go installed:

```bash
go install github.com/garbas/tgdash@latest
```

## With Nix

Run directly without installing:

```bash
nix run github:garbas/tgdash
```

Or add to your flake inputs:

```nix
{
  inputs.tgdash.url = "github:garbas/tgdash";
}
```

## From release binaries

Download pre-built binaries from the
[GitHub releases](https://github.com/garbas/tgdash/releases)
page. Binaries are available for Linux and macOS
(amd64 + arm64).

## Verify installation

```bash
tgdash --help
```
