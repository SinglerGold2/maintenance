# Maintenance for Prusa Connect

A Prusa Connect extension that counts the print hours of each printer and tells
you when a maintenance task is due.

## What it does

```
Per printer    a Maintenance tab with the hours this server saw, the count the
               printer's own display shows, and one line per task
Per fleet      a Maintenance page with the schedule every printer starts with
On the board   a dashboard card, Maintenance due, listing what is due now
On the bus     maintenance.task whenever a task reaches due soon or due
```

A task counts in the printer's own hours, so a machine whose display reads 343
hours owes you its 200-hour check at once, even if this server only watched the
last 60 of them.

## After installing: what to do

1. **Open a printer and its Maintenance tab.** The first time, the tab lists the
   four tasks the schedule ships with, each at 0 hours.
2. **Type the printer's counter in.** Read the value on the printer's display,
   type it as `343` or `14d 7h`, and press Save. From then on the tab shows the
   printer's counter and, beside it, the hours this server saw. The two move
   together with every print: the printer keeps the lead it had when you typed
   it in.
3. **Adjust the schedule** on the Maintenance page in the sidebar. Change an
   interval there and every printer's task of that name follows. The defaults
   come from the Prusa Knowledge Base:

   | Task | Default |
   | --- | --- |
   | Check axes, belts, fans and extruder | 200 print hours |
   | Inspect the electrical connectors | 600 print hours |
   | Lubricate the linear rail | 90 days |
   | Check the nozzle, replace when worn | 5000 print hours, hardened steel |

4. **Press Done when you did the work.** A dialog asks first; the counter starts
   again at the printer's current hours only after a yes. Start over at 0 does
   the same without asking.
5. **Add your own tasks** at the bottom of the printer's tab: a name, an
   interval, and whether it counts print hours or days. Your own tasks can be
   removed; the shipped ones cannot.
6. **Put the card on the dashboard** from the dashboard editor. It lists every
   task that is due now or due soon, across the fleet.

## Settings

| Key | Type | Default | Meaning |
| --- | --- | --- | --- |
| `alert_percent` | int | `90` | The share of an interval at which a task reads as due soon |

```yaml
resources:
  sink:
    maintenance:
      config:
        alert_percent: 85
```

## Events

When a task reaches due soon or due, the module publishes one
`maintenance.task` event per change with the printer, the task, its share, its
status and its nozzle material. Point a webhook, MQTT, AMQP or Kafka sink at it
to carry the reminder to a phone.

## How the hours are counted

The server only knows the prints it recorded. The printer's own counter is
therefore a value you type in, and the module keeps the distance between that
value and its own count. A task's counter is held in the printer's hours, so
both the display and the tasks agree with the machine's display.

The print history has no per-tool time, so with a multi-tool printer every tool
counts the whole print time. Per-tool attribution from the G-code is the next
step on the list.

## Installing it by hand

Useful before the marketplace lists the release:

```bash
make release
tar czf /tmp/maintenance.tar.gz -C release $(ls -A release)
docker cp /tmp/maintenance.tar.gz prusa-connect-local:/tmp/maintenance.tar.gz
docker exec -w /app prusa-connect-local /app/prusa_connect_local \
  module add github.com/singlergold2/maintenance@0.9.2 \
  -from /tmp/maintenance.tar.gz -approve -database sqlite
docker restart prusa-connect-local
```

## Working on it

```bash
make core      # the Prusa Connect checkout the SDK and the npm packages come from
make check     # Go vet and tests, the client is current, the bundle lints and builds
make release   # the release directory, in release/
```

The module path is `github.com/singlergold2/maintenance`; the host and the
marketplace derive the extension id from it,
`com.github.aux.singlergold2.maintenance`.

## Licence

MIT. See `LICENSE`.
