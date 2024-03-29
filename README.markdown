# Freeze Calendar Resource

This is a Concourse resource that fails (or holds up) a job if the current timestamp is within one of the given freeze windows. This is meant to block deployments during freeze periods.

The resource has two modes of operation:

1. Fuse: fail the execution if there is a freeze window within the next `n` minutes
1. Gate: hold up the execution of a job while there is a freeze window within the next `n` minutes

Emergency deploys are still possible by removing the `get` step from the pipeline.

# Source Configuration

```yaml
- name: my-project-freeze-calendar
  type: freeze-calendar
  source:
    uri: git@github.example.com:my-project/freeze-calendar
    branch: main
    private_key: ((vault/my-key))
    known_hosts: # see https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/githubs-ssh-key-fingerprints
    path: subdir/project-freeze-calendar.yaml
    retry_interval: 1m
```

# `check` Behavior

Fetches the latest freeze calendar and emit its version (e.g. git SHA).

# `get` Behavior

In `fuse` mode:

  1. Fetch the calendar at the version that was discovered by the `check` step
  1. Fail if we are within a freeze window (with a matching scope, if set).

In `gate` mode:

  `loop`:

  - fetch the _latest_ version of the freeze calendar
  - exit `0` we are _not_ within a freeze window (with a matching scope, if set)
  - sleep for `retry_delay`

## Optional Parameters

* `runway` (expected deploy time) will be taken into consideration so that there is enough time left to complete the deployment before the next freeze begins.

  Accepts any string that Go's [`time.ParseDuration`](https://pkg.go.dev/time#ParseDuration) can parse.

* `retry_interval`: How long to wait until fetching the source again and re-checking if we can pass the gate. Defaults to `10s`.

  Accepts any string that Go's [`time.ParseDuration`](https://pkg.go.dev/time#ParseDuration) can parse.

# `put` Behavior

no-op

# Example

Do not deploy if a window of the given `freeze-calendar` has the scope `eu-de` in its list:

```yaml
- get: project-freeze-calendar
  params:
    scope: eu-de
```

# Freeze Calendar Format

One freeze calendar may have `0..n` freeze windows

```yaml
freeze_calendar:
  - name: Holiday Season
    starts_at: 2022-12-01T06:00:00Z # native YAML timestamp
    ends_at: 2022-12-27T06:00:00Z
    scope:
      - eu-de
      - us-east
      - ap-southeast
  - name: Another one
    ...
```

# FAQ

## I have multiple freeze calendars, can you support that?

This resource always operates on a single freeze calendar. You might, however, synthesize a single calendar from multiple sources and use it as this resource's calendar.

## What if I realize the freeze calendar is wrong and the `get` step is already running?

Update the calendar and push the changes:

* If the resource is running in `gate` mode, the get step will update the repo while in front of the gate. It will pick up the new version eventually. If the pipeline is already past the step, you'll have to stop the pipeline manually.
* If the resource is running in `fuse` mode, it will use the version discovered by the check step. Re-running the job with fresh inputs should be sufficient.

# TODO

* Use the new [concourse-resource-go](https://github.com/suhlig/concourse-resource-go) interface
