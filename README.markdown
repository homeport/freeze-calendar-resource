# Deployment Window Resource

Holds up jobs if current timestamp is within one of the given freeze windows. This is meant to block deployments during blackout periods.

Two modes of operation:

1. Gate: hold up the execution of a job while there is a freeze window within the next `n` minutes
1. Fuse: fail the execution if there is a freeze window within the next `n` minutes

Emergency deploys are still possible by removing the `get` step from the pipeline.

# Source Configuration

```yaml
- name: project-freeze-calendar
  type: freeze-calendar
  source:
    uri: git@github.example.com:my-project/freeze-calendar
    branch: main
    private_key: ((vault/my-key))
    path: subdir/project-freeze-calendar.yaml
```

# `check` Behavior

Fetches the latest freeze calendar and emit its version (e.g. git SHA).

# `get` Behavior

* loop:
  - fetch the _latest_ version of the freeze calendar
  - if gate: sleep if we are within a freeze window, retry after `$INTERVAL`
  - if fuse: fail if we are within a freeze window

# `put` Behavior

no-op

# Example

Do not deploy if a window of the given `freeze-calendar` has the scope `eu-de` in its list:

```yaml
- get: project-freeze-calendar
  params:
    scope: eu-de
```

# TODO

* Allow [private repos](https://pkg.go.dev/github.com/go-git/go-git/v5#example-PlainClone-AccessToken)
* Is it worth cloing into InMemory?
* Add get parameter for `runway` (expected deploy time) in order to not start if there is not enough time left to complete the deployment before the next freeze begins
* Get step writes the fetched freeze calendar to disk (for consumption by following tasks)

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

# Misc

* Multiple freeze calendars are an external concern; this resource always operates on a single freeze calendar (that might be generated from multiple sources by another process).
