# Changelog

What an operator sees change when they upgrade this module, newest release
first. `version:` in `module.yaml` is the newest heading below.

## 0.3.0 - 2026-11-17

### Added

- A printer's card tells the two counters apart: what the server tracked and
  what the machine's own display shows. The operator types the machine's number
  in, with a hint saying to read it off the printer.
- `POST /devices/hours` stores the count the operator read off the printer.

## 0.2.0 - 2026-11-17

### Added

- Every task names what it reads at a glance: the interval, the hours or days
  since the last check, and the point at which it reads as due soon, in the
  currency of the task.
- A printer's card leads with the print hours so far and the number of prints.
- A task may name the nozzle material, so a nozzle task carries the wear class
  the operator works with. The CORE One default is a hardened steel nozzle at
  5,000 hours; brass, ruby and diamond are meant to be typed in and tuned.
- `POST /tasks` takes `material` and `last_done_hours`, so an operator can set a
  task's counter to any value, zero included.

### Changed

- The default nozzle task counts 5,000 hours instead of 250, because 250 was a
  made-up number and 5,000 is the class a hardened steel nozzle is in.

## 0.1.1 - 2026-11-17

### Fixed

- The page no longer stops working on its first draw: the draft of a new task is
  prepared when the fleet is read, not while the page renders.

## 0.1.0 - 2026-11-17

### Added

- The page **Maintenance** lists every printer with its print hours, its print
  count and its maintenance tasks.
- `GET /overview` answers the fleet with the hours, the prints and the tasks,
  each with the share of its interval used and one of `ok`, `soon` or `due`.
- `POST /tasks` creates a task on a printer or changes one, `POST /tasks/done`
  sets its counter to the printer's current print hours, and `DELETE /tasks`
  removes it.
- The setting `alert_percent` (default `90`) sets the share of an interval at
  which a task reads as due soon.
- A printer with no tasks yet gets the schedule Prusa publishes for its model:
  the CORE One starts with 200-hour axes and belts, 600-hour connectors, a
  90-day rail lubrication and a 250-hour nozzle check.
