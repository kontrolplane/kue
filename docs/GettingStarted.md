# Getting Started with Kue

Welcome to **Kue**, a Terminal User Interface (TUI) for Amazon SQS.

This quick start guide will help you install Kue, connect to your AWS account, and open your first queue view.

## Prerequisites

- A working Go environment (version TODO)
- AWS credentials with access to SQS (e.g., via `~/.aws/credentials` or environment variables)

## Installation (quick)

```
go install github.com/kontrolplane/kue@latest
```

Alternatively, see the detailed [Installation](./Installation.md) page for other options.

## First Run

```bash
kue
```

Kue will automatically try to use your default AWS profile. If you need to specify another region or profile, see [Configuration](./Configuration.md).

## Next Steps

- Learn the [keybindings](./Usage.md#keybindings)
- Explore queues and messages
- Contribute improvements! See [Contributing](./Contributing.md)
