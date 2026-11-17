# Maintenance for Prusa Connect

Counts the print hours of every printer and tells you when a maintenance task is
due. Live in the marketplace: install it under Settings, Extensions.

## What it does

Every printer's page gets a **Maintenance** tab with its own counter: the hours
this server saw, the value the printer's display shows, and one line per task
with its interval, the hours since the last check, and the point at which it
reads as due soon.

Type the printer's own counter in once, as `343` or `14d 7h`, and both counters
move with every print from then on. A CORE One starts with the schedule from the
Prusa Knowledge Base:

| Task | Default |
| --- | --- |
| Check axes, belts, fans and extruder | 200 print hours |
| Inspect the electrical connectors | 600 print hours |
| Lubricate the linear rail | 90 days |
| Check the nozzle, replace when worn | 5000 print hours, hardened steel |

The **Maintenance** page in the sidebar holds one schedule for the fleet, so a
changed interval reaches every printer's task of that name; the hours each
printer collected stay. **Done** asks first and starts the next interval at the
printer's current hours. Tasks that ship with the module cannot be removed, the
ones you add can.

Each printer's tab also says how many hours each tool ran, read from the G-code
of finished prints (`.gcode` and `.bgcode`, deflate, heatshrink and MeatPack).
Treat that figure as an indication, not a measurement. Each tool can be reset
after a nozzle change, without touching the printer's counter.

A dashboard card, **Maintenance due**, lists what is due across the fleet, and
the module publishes `maintenance.task` on the event bus when a task reaches due
soon or due.

Setting `alert_percent`, 90 by default, sets when a task reads as due soon.

## Install by hand

The marketplace is the normal way. To install a release yourself:

```bash
# the signed bundle from the GitHub release
docker cp maintenance-0.10.10.tar.gz prusa-connect-local:/tmp/
docker exec -w /app prusa-connect-local /app/prusa_connect_local \
  module add github.com/singlergold2/maintenance@0.10.10 \
  -from /tmp/maintenance-0.10.10.tar.gz -approve -database sqlite
docker restart prusa-connect-local
```

On a native Linux server the same command works without docker, with
`./prusa_connect_local`.

On Windows the marketplace is the easy way: Settings, Extensions, Maintenance,
Install. To do it by hand, open PowerShell in the folder the server runs from
(the one with `config/` in it) and use the bundle, not the `.exe` from the
release:

```powershell
cd C:\prusa-connect
.\prusa_connect_windows_amd64.exe module add github.com/singlergold2/maintenance@0.10.10 `
  -from C:\Users\you\Downloads\maintenance-0.10.10.tar.gz -approve -database sqlite
```

Then restart the server. It unpacks the bundle into
`config\config.d\modules\maintenance\bin` itself; nothing is copied by hand.
The `.exe` files in a release are the module's own program, which the server
starts as a child process, not installers.

A release carries one program per platform, named after it:

| File | Runs on |
| --- | --- |
| `maintenance_linux_amd64` | Linux, x86-64 |
| `maintenance_linux_arm64` | Linux on ARM 64, such as a Raspberry Pi |
| `maintenance_darwin_arm64` | macOS on Apple silicon |
| `maintenance_windows_amd64.exe` | Windows, x86-64 |

These are the module's own program, which the server starts as a child process;
they are not installers. The signed `.tar.gz` carries the same programs plus
`module.yaml`, the browser bundle and the checksums, and that is what an
installation takes.

The module asks for the fleet, the print history, `read_storage` for the G-code,
and a database of its own. The service runs as the Docker container
`prusa-connect-local` on port 8090.

## Development

```bash
make core      # the Prusa Connect checkout the SDK and the npm packages come from
make check     # Go vet and tests, the client is current, the bundle lints and builds
make release   # the release directory, in release/
```

## Licence

MIT. See `LICENSE`.
