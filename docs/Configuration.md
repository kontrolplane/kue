# Configuration

Kue reads AWS configuration the same way the AWS SDK does. Most users can rely on existing profiles in `~/.aws/credentials`.

## Flags / Environment Variables

| Option | Env Var | Default | Description |
|--------|---------|---------|-------------|
| `--region` | `AWS_REGION` | value in profile | AWS region to use |
| `--profile` | `AWS_PROFILE` | `default` | Named profile to load |

**TODO**: Document additional flags or config file once implemented.
