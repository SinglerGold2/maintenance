# Changelog

What an operator sees change when they upgrade this module, newest release
first. `version:` in `module.yaml` is the newest heading below.

## 0.14.0 - 2026-11-19

Prettier, more capable, and it speaks German and Czech.

### Added

- A manual page beside the module: what it counts, how a printer picks its set,
  the nozzle lives, the tool hours, what each colour means, and what to check
  when a figure looks wrong. It carries English and German, independent of the
  interface language, and shows seven screenshots with captions: the printer
  tab, the schedules, a schedule open with its nozzle materials, the card
  gallery, its setting and two cards on a dashboard. Section headings are bold,
  and the three states are spelled out in their own colour.
- A language setting on the extension's Settings page, next to `alert_percent`:
  `system` follows the browser, `en`, `de` and `cs` pin one. It applies to the
  page, the printer tab, the cards and the manual, and takes effect when the
  module restarts. A country code such as `cz` or a name such as `Deutsch` is
  accepted and turned into the tag the build ships.
- The module speaks German and Czech. Every text sits in `locales/<tag>.json`,
  the manual in `manual/<tag>.json`, and the interface follows the language the
  browser asks for. The manual opens in the system language when it has a file
  for it. A language that is not there falls back to English, and the README
  says how to add one.
- One dashboard card, Maintenance, with a setting: all printers, or one picked
  from the fleet. Put one card per printer on the dashboard and name each in its
  title, or leave the setting at all printers for the fleet view. The card keeps
  the printer name, every task with its share, the schedule and the printed
  hours.
- The Maintenance due card names the printer. A printer without a name shows its
  model.

### Changed

- The due card shows every task with its share of the interval and a coloured
  dot, red first, then yellow, then green. Before, a printer's name was not drawn
  at all and only the marked tasks appeared. Both cards and the printer tab use
  the app's own state tints, so the colours follow the theme and mean the same
  as on the Printers screen. On a printer's Maintenance tab the word reads green
  for ok, yellow at the threshold and red for due now; before, ok was grey and
  looked like a note rather than a state. The middle state uses the app's warning
  yellow, not its attention orange, which sat too close to the red.
- The manual states that 14d 7h and 343 h are the same figure, and that tool
  hours are counted to the second: on an INDX print a nozzle may work for forty
  seconds, which a count in whole minutes would hide.
- The colour section of the manual describes the card as it is now: where a task
  sits in its group, and that an all-green printer keeps its group because a
  green line is the proof that the interval holds.
- The Tasks button on a set is green and outlined. As a ghost button it read as a
  label between Disable and Duplicate, and it is the one that opens the plan.
- The marketplace summary names the schedules, the nozzle life by material and
  the tool hours read from the G-code.
- The bundle declares 0.1.0 as the oldest UI library, the one it is built
  against, instead of 0.0.1.
- The README opens with the artwork, and its closing section names every source
  of that artwork, its licence and its author.

### Fixed

- The dots and the coloured figures on the card were invisible: the classes were
  assembled from a variable, which the stylesheet compiler cannot see, so the
  rules never reached the bundle. They are written out now.
- The German manual is written with its umlauts. The first draft spelled them as
  ae, oe and ue, which reads like a telegram.

## 0.10.2 - 2026-11-17

### Fixed

- The module asks for `read_storage`, so the server lets it read a print's file.
  0.10.1 asked for the API but not for the permission the API checks.

## 0.10.1 - 2026-11-17

### Fixed

- The migration that adds the tool tables is named with a time that exists.
  0.10.0 named it 24:00:00, and the database refused it, so the module did not
  start.

## 0.10.8 - 2026-11-17

### Fixed

- The reset button is on the tool line itself. 0.10.7 built the dialog but the
  line carried no button.

## 0.10.7 - 2026-11-17

### Added

- Each tool can be reset from the printer's Maintenance tab, after a nozzle
  change. A dialog asks first; the printer's own counter is not touched.
- The tab says the per-tool hours are an indication, not a measurement.

## 0.10.6 - 2026-11-17

### Fixed

- A printer with one tool reports the printer's own counter as that tool's
  hours. It used to report only the hours of the prints whose file this module
  could still read, which read far too low.

## 0.10.5 - 2026-11-17

### Fixed

- A print whose file is gone from the storage is not opened again on every
  load; it counts as read.
- A print is marked read only once its file was read, so a failure no longer
  loses its hours.

## 0.10.4 - 2026-11-17

### Fixed

- The prints the earlier 0.10 releases failed on are read again.

## 0.10.3 - 2026-11-17

### Fixed

- A print is read from the server's storage by the file's own name when the
  printer's path (`usb/…`) names no file there.

## 0.10.2 - 2026-11-17

### Fixed

- The module asks for `read_storage`, the permission the storage API checks.

## 0.10.1 - 2026-11-17

### Fixed

- The tool tables are added by a migration with a name the database accepts.

## 0.10.0 - 2026-11-17

### Added

- The printer's Maintenance tab says how many hours each tool ran. The module
  reads the G-code of a finished print once, both `.gcode` and `.bgcode`
  (deflate, heatshrink 11/4 and 12/4, MeatPack with and without spaces), and
  splits the print's length by the filament each tool pushed.
- The module asks for `core-storage.stat` and `core-storage.fetch` to read a
  print's file from the server's storage.

## 0.9.2 - 2026-11-17

### Fixed

- The release workflow is valid YAML again: the named signing secret sits beside
  `with`, not inside it.

### Changed

- README: what to do after installing, step by step, and how the hours are
  counted.

## 0.9.1 - 2026-11-17

### Fixed

- The release workflow hands the signing key to the marketplace's publish
  workflow by name. It inherited its secrets before, and an inherited secret
  does not cross an organisation boundary.

## 0.9.0 - 2026-11-17

### Changed

- The Maintenance page is one schedule for the whole fleet: four lines, no
  printer, no counters. Changing an interval there carries it to every built-in
  task of that name on every printer.
- `GET /template` and `POST /template` carry that schedule. A printer that joins
  the server takes its tasks from the template.

## 0.8.2 - 2026-11-17

### Changed

- The Maintenance page is about intervals alone: no counter, no printer hours
  field, no material field. The printer's own numbers live on its Maintenance
  tab.

## 0.8.1 - 2026-11-17

### Fixed

- The manifest declares the printer tab and the dashboard card the bundle
  registers. Without them the host refused the module as undeclared.

## 0.8.0 - 2026-11-17

### Added

- Every printer page carries a Maintenance tab of its own, with that printer's
  counter and tasks alone.

## 0.7.0 - 2026-11-17

### Added

- A dashboard card, Maintenance due, lists every task that is due now or due
  soon, with its printer.

## 0.6.1 - 2026-11-17

### Added

- Done asks first: a dialog names the task and says the counter starts again at
  the printer's current hours.

## 0.6.0 - 2026-11-17

### Added

- The module publishes `maintenance.task` on the event bus when a task reaches
  due soon or due, once per change.

## 0.5.0 - 2026-11-17

### Changed

- A task counts in the printer's own hours, not in what this server has seen.
  A machine whose display reads 343 hours shows 343 hours of elapsed time on
  every built-in task, so a 200-hour task reads as due at once, as it should.
- Done and Start over anchor the next interval at the printer's counter.
- The built-in tasks start at the printer's zero, because nobody recorded a
  service on this server before.

## 0.4.3 - 2026-11-17

### Fixed

- The distance stored by 0.4.2 counted the hours the server had already seen
  twice. It is rebuilt from the count the operator typed in and the hours the
  server has tracked, once.

## 0.4.2 - 2026-11-17

### Fixed

- The printer's own counter moves on with every print. What the operator types
  in sets the distance between the machine's display and the hours the server
  has seen, and that distance stays. A printer 283 hours ahead stays 283 hours
  ahead after the next print.

## 0.4.1 - 2026-11-17

### Changed

- Every task keeps its interval field, so any task can be retuned. The free
  text field for the nozzle material is gone; the material a task carries is
  shown beside its interval, and a proper picker replaces the text field later.

## 0.4.0 - 2026-11-17

### Added

- A task the module ships carries `builtin: true`. Its interval may be changed,
  but `DELETE /tasks` refuses it with `BUILTIN_TASK`, and the page shows no
  Remove button for it. Only the tasks an operator writes can be removed.

## 0.3.4 - 2026-11-17

### Fixed

- A printer whose tasks were removed by 0.3.3 gets its default schedule written
  again on the next load, once.

## 0.3.3 - 2026-11-17

### Fixed

- The default schedule is written with the printer's id. Earlier releases wrote
  it without, so the tasks belonged to no printer and every load wrote them
  again. The rows that belonged to no printer are removed.
- The field for a new task shows a real example, `e.g. fasten all screws`, and
  starts at 1,000 hours.

## 0.3.2 - 2026-11-17

### Fixed

- A printer's default schedule is written once. Two page loads that arrived
  together used to write it twice, so a printer could carry the same four tasks
  several times over.
- The duplicates the earlier releases wrote are removed when this release opens
  the database.

## 0.3.1 - 2026-11-17

### Changed

- The printer's own counter takes what the machine writes on its display, such
  as `14d 7h`, and turns it into hours.
- The interval and material fields stay only on a nozzle task, where they mean
  something. Other tasks keep their fixed interval.

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
