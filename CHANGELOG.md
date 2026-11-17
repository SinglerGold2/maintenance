# Changelog

What an operator sees change when they upgrade this module, newest release
first. `version:` in `module.yaml` is the newest heading below.

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
