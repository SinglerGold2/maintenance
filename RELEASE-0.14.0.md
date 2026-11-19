Maintenance 0.14.0

Prettier, more capable, and it speaks German and Czech.

This is the largest release since the module appeared. The counter is the same
honest counter it always was; what changed is how it reads, what it explains,
and who it speaks to.

## A manual page

The module now explains itself. A second page beside the schedules walks through
what is counted, how a printer picks its plan, what the nozzle lives mean, where
the tool hours come from, what each colour says, where the language setting
lives, and what to check when a figure looks wrong.

It carries eight screenshots with captions, and the captions are in the same
language as the text:

- the Maintenance tab of a printer, with its counter and its tasks
- the schedules page
- a schedule open, with its nozzle materials
- the Schedule field on a printer
- the card gallery
- the card's setting
- two cards on one dashboard
- the language field in the server's settings

German and Czech are shipped; the button switches between EN, DE and CZ without
touching the rest of the app.

## One card, configured

There were two cards, and they overlapped. There is now one, with a setting: all
printers, or one picked from the fleet. A dashboard can therefore carry one card
per printer, each titled by its owner, or a single card for the whole farm.

The card names its printer, shows every task with its share of the interval, and
keeps the schedule and the printed hours beside it. Every task, not only the
marked ones: a green line is the proof that this interval still holds.

## Colours that mean something

The card, the printer tab and the schedule lines wear the app's own state tints.
A green dot here means what a green dot means on the Printers screen, in light
and dark alike.

The middle state is the warning yellow. It used to be the attention orange, which
sat so close to the fault red that the two were hard to tell apart on a small
dot.

## Speaks your language

A language setting on the extension's Settings page, next to `alert_percent`:
`system` follows the browser, or pin `en`, `de` or `cs`. A country code such as
`cz` or a name such as `Deutsch` is accepted and turned into the tag the build
ships.

Every text is a file. The interface reads `locales/<tag>.json`, the manual reads
`manual/<tag>.json`, so adding a language is two copied files and a pull request,
and the README says exactly that. The manual asks for help in all three
languages.

The Czech is a trial balloon: translated by a machine, offered so the mechanism
can be seen. Whether it holds up is for Czech readers to decide, and a pull
request that improves it is more welcome than one that adds a language.

## Fixed

- The card drew no dots at all. The colour classes were assembled from a variable,
  which the stylesheet compiler cannot see, so the rules never reached the
  bundle. They are written out now.
- The German manual spelled its umlauts as ae, oe and ue. It reads like a telegram;
  eighteen texts now carry the letters they should.
- The colour section of the manual described the card as it was before every task
  was shown, and said a green task is not marked.

## Upgrading from 0.13.0

**The dashboard cards were renamed.** `due` and `printer` became one card,
`maintenance`. A dashboard that still holds the old ones shows gaps where they
were: remove those and add the new card, then pick its printer in the card's
settings.

**A language change restarts the module.** The value is read at start-up, so the
new language arrives with the next run. Pick a quiet moment: an import that is
halfway through, or a print whose G-code is being read, is cut off by that
restart.

Nothing else needs doing. Tasks, sets, counters, Done marks and tool hours are
untouched by the upgrade.

## Install

From the marketplace: Settings, Extensions, Maintenance, Install.

By hand, from the bundle in this release:

```bash
docker cp maintenance-0.14.0.tar.gz prusa-connect-local:/tmp/
docker exec -w /app prusa-connect-local /app/prusa_connect_local \
  module add github.com/singlergold2/maintenance@0.14.0 \
  -from /tmp/maintenance-0.14.0.tar.gz -approve -database sqlite
docker restart prusa-connect-local
```

## Compatibility

```
Core            >=0.1.0-beta.1
UI library      >=0.1.0 <1.0.0
Contract        1
Platforms       linux_amd64, linux_arm64, darwin_arm64, windows_amd64
Permissions     observe, read_storage, a database of its own
Events          maintenance.task, when a task reaches due soon or due
```

## Artwork

The module's banner, icon and marketplace logo combine three icons from SVG
Repo. The wrench requires attribution, so it is named here, in `CREDITS.md` and
on the images themselves.

- 3D printer: CC0, SVG Repo
- Wrench: CC Attribution, Lugine Hiary
- Clock: Public Domain, Sensa

## Thanks

To everyone who installed 0.13.0 and told us what was missing. The manual, the
single card and the colours come from those notes. If your language is not here
yet, the files are waiting.
