Maintenance 0.14.0

Prettier, more capable, and it speaks German and Czech.

**A manual page.** What the module counts, how a printer picks its schedule, the
nozzle lives by material, the tool hours, what each colour means, what to check
when a figure looks wrong, and where the language setting lives. Eight
screenshots with captions, in English, German and Czech, and one of them shows
the language field itself.

**One card, configured.** The dashboard card takes a printer or leaves it at the
whole fleet, so a dashboard can carry one card per printer or one for the fleet.
Every task is shown with its share of the interval, red first, then yellow, then
green.

**Colours that mean something.** The card, the printer tab and the schedules use
the app's own state tints, so a green dot here says what a green dot says on the
Printers screen. The middle state is the warning yellow, not the attention
orange that sat too close to the red.

**Speaks your language.** A language setting on the extension's Settings page:
`system` follows the browser, `en`, `de` and `cs` pin one. Every text lives in
`locales/<tag>.json` and every manual page in `manual/<tag>.json`, so a
translator copies one file. The Czech is a trial balloon, translated by a
machine, and the manual says so and asks for help on GitHub.

**Fixed** the card drawing no dots at all, because the classes were assembled
from a variable and never reached the stylesheet, and the German manual spelling
its umlauts as ae, oe and ue.

A language change restarts the module, so pick a quiet moment.
