# Maintenance for Prusa Connect

![Maintenance for Prusa Connect](artwork/card.png)

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

The **Maintenance** page in the sidebar holds the schedules, called sets. Eight
ship with the module, one per model and one for any printer, each starting from
the Prusa Knowledge Base and each linking the article it came from. A printer
follows the set its model brings until you pick another one in the **Schedule**
field on its tab; its own hours and Done marks stay.

A shipped set can be switched off, never edited or deleted: duplicate it, edit
the copy, and assign that. A new set starts empty. Its lines take an interval in
print hours or in days, which is how a quarterly check is asked for no matter
how much the printer runs.

A nozzle line picks its material, and the interval follows: brass 500 h,
stainless steel 700 h, plated copper 800 h, hardened steel 1500 h, bimetal
1700 h, silicon carbide 2500 h, tungsten carbide 3000 h, ruby/diamond 5000 h.
The first three are printed in bold italic because carbon and glass filled
filament ruins them. **Done** asks first and starts the next interval at the
printer's current hours; a shipped task cannot be removed, the ones you add can.

Each printer's tab also says how many hours each tool ran, read from the G-code
of finished prints: `.gcode` and `.bgcode`, deflate, heatshrink and MeatPack
included. The print's length is split by the filament each tool pushed, and the
figure is an indication, not a measurement; the tab says so too. Each tool
carries its own **Start over at 0**, so a nozzle change resets that tool alone
and the printer's counter stays where it is. Times show as `5h 4m 33s`, and a
print of a few seconds shows seconds rather than a rounded zero.

A dashboard card, **Maintenance due**, lists what is due across the fleet, and
the module publishes `maintenance.task` on the event bus when a task reaches due
soon or due.

Setting `alert_percent`, 90 by default, sets when a task reads as due soon.

## Install by hand

The marketplace is the normal way. To install a release yourself:

```bash
# the signed bundle from the GitHub release
docker cp maintenance-0.13.0.tar.gz prusa-connect-local:/tmp/
docker exec -w /app prusa-connect-local /app/prusa_connect_local \
  module add github.com/singlergold2/maintenance@0.13.0 \
  -from /tmp/maintenance-0.13.0.tar.gz -approve -database sqlite
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
.\prusa_connect_windows_amd64.exe module add github.com/singlergold2/maintenance@0.13.0 `
  -from C:\Users\you\Downloads\maintenance-0.13.0.tar.gz -approve -database sqlite
```

Then restart the server. It unpacks the bundle into
`config\config.d\modules\maintenance\bin` itself; nothing is copied by hand.
The `.exe` files in a release are the module's own program, which the server
starts as a child process, not installers.

On macOS the marketplace is the easy way too. By hand, open Terminal in the
folder the server runs from (the one with `config/` in it):

```bash
cd ~/prusa-connect
./prusa_connect_macos_arm64 module add github.com/singlergold2/maintenance@0.13.0 \
  -from ~/Downloads/maintenance-0.13.0.tar.gz -approve -database sqlite
```

Then restart the server; it unpacks the bundle into
`config/config.d/modules/maintenance/bin`. If macOS refuses a binary you
downloaded, clear the quarantine flag once with
`xattr -d com.apple.quarantine <file>`.

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

## Translating

Every text this module draws lives in a JSON file, one per language. The
interface reads `ui/src/locales/<tag>.json`, the manual reads
`ui/src/manual/<tag>.json`. English is the source, and anything missing falls
back to it.

```bash
cp ui/src/locales/en.json ui/src/locales/fr.json
cp ui/src/manual/en.json  ui/src/manual/fr.json
```

Translate the values, leave the keys alone, and send a pull request. The
manifest needs no change: a file that is there is used, a file that is not is
not a fault. Two rules keep it reviewable: the keys stay as they are, and the
English file is never edited into another language.

The manual's language switch offers whatever files exist, and opens in the
system language when one matches.

## Development

```bash
make core      # the Prusa Connect checkout the SDK and the npm packages come from
make check     # Go vet and tests, the client is current, the bundle lints and builds
make release   # the release directory, in release/
```

## Licence

MIT. See `LICENSE`.

## Artwork

`artwork/card.png` and `artwork/icon.png` combine three icons from SVG Repo:

| Icon | Source | Licence | Author |
| --- | --- | --- | --- |
| 3D printer | [3d-printer](https://www.svgrepo.com/svg/19622/3d-printer) | CC0 | SVG Repo |
| Wrench | [wrench-13](https://www.svgrepo.com/svg/492002/wrench-13) | CC Attribution | [Lugine Hiary](https://www.svgrepo.com/author/Lugine%20Hiary/) |
| Clock | [clock](https://www.svgrepo.com/svg/434037/clock) | Public Domain | [Sensa](https://www.svgrepo.com/author/Sensa/) |

The wrench requires attribution, so Lugine Hiary is named here and in
[artwork/CREDITS.md](artwork/CREDITS.md). The banner carries the same credit in
its corner. The project code stays MIT; the artwork keeps the licences listed
above.
