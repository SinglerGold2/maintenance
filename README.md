# Maintenance for Prusa Connect

A Prusa Connect extension that counts the print hours of each printer and
reminds you when a maintenance task is due.

- **The process** (`main.go`): a program Prusa Connect starts for the extension.
  It keeps the tasks in its own SQLite database, reads the print history through
  the Connect SDK, and serves the routes in `openapi.json`.
- **The browser bundle** (`ui/`): a Svelte page called **Maintenance** in the
  Connect web app, reading the routes through a typed client generated from
  `openapi.json`.

## What it does

A task is a check with an interval, either in print hours or in days.

| Field | Meaning |
| --- | --- |
| name | What to do, such as "Check the belts" |
| kind | `hours` counts print time, `days` counts calendar time |
| interval | How often, in hours or days |
| last done | The print hours and the date when it was last done |

The page lists every printer with its print hours and its tasks, each with the
share of its interval used and one of three states: `ok`, `soon` (at the
`alert_percent` share, 90 % by default) and `due`.

**Done** sets the task's counter to the printer's current print hours, so the
next interval starts where the machine stands now.

## Defaults

A printer that has no tasks yet gets the schedule Prusa publishes for its model.
For the CORE One, from the Prusa Knowledge Base article on regular printer
maintenance:

| Task | Interval |
| --- | --- |
| Check axes, belts, fans and extruder | every 200 print hours |
| Inspect the electrical connectors | every 600 print hours |
| Lubricate the linear rail | every 90 days |
| Check the nozzle, replace when worn | every 250 print hours |

A model this release does not know gets one generic task of 200 print hours.

## Settings

| Key | Type | Default | Meaning |
| --- | --- | --- | --- |
| `alert_percent` | int | `90` | The share of an interval at which a task reads as due soon |

Write it in the instance's config block:

```yaml
resources:
  sink:
    maintenance:
      config:
        alert_percent: 85
```

## Installing it by hand

```bash
make release
tar czf /tmp/maintenance.tar.gz -C release $(ls -A release)
docker cp /tmp/maintenance.tar.gz prusa-connect-local:/tmp/maintenance.tar.gz
docker exec -w /app prusa-connect-local /app/prusa_connect_local \
  module add github.com/singlergold2/maintenance@0.1.0 \
  -from /tmp/maintenance.tar.gz -approve -database sqlite
docker restart prusa-connect-local
```

## Working on it

```bash
make core      # clone the Prusa Connect checkout the SDK and the npm packages come from
make check     # Go vet and tests, the client is current, the bundle lints, type-checks and builds
make release   # the release directory, in release/
```

The module path is `github.com/singlergold2/maintenance`. The host and the
marketplace derive the extension id from it,
`com.github.aux.singlergold2.maintenance`; you never write the id yourself.

## Licence

MIT. See `LICENSE`.
