<script lang="ts">
  import {
    Button,
    Card,
    CardContent,
    PageHeader,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
  } from '@prusa/connect-ui/primitives';

  import cardGallery from './assets/cards-gallery.webp';
  import cardSettings from './assets/card-settings.webp';
  import dashboard from './assets/dashboard.webp';
  import printerSchedule from './assets/printer-schedule.webp';
  import printerTab from './assets/printer-tab.webp';
  import scheduleTasks from './assets/schedule-tasks.webp';
  import schedules from './assets/schedules.webp';
  import english from './manual/en.json';
  import german from './manual/de.json';
  import czech from './manual/cs.json';
  import { applyLanguage, currentTag, t } from './i18n.svelte';

  // One file per language under manual/, so a translator copies en.json and
  // works in their own. The page follows the extension's language setting; the
  // buttons switch it for this visit.
  const books = { en: english, de: german, cs: czech } as const;
  type Tag = keyof typeof books;
  const tags = Object.keys(books) as Tag[];

  // The page follows the extension's language setting. The buttons switch it
  // for this visit without writing anything.
  const language = $derived<Tag>(currentTag());

  // The file follows the language tag, which is what the browser reports; the
  // button shows the country, which is what a reader recognises.
  const labels: Record<Tag, string> = { en: 'EN', de: 'DE', cs: 'CZ' };

  function m(key: string): string {
    const book = books[language] as Record<string, string>;

    return book[key] ?? (english as Record<string, string>)[key] ?? key;
  }

  // The English name stays the one the module and the slicer use; the German
  // manual hangs a short translation beside it, because a German reader who
  // sees only "Hardened steel" has to guess which nozzle that is.
  const materials: [string, string, string][] = [
    ['materialBrass', '500', 'Messing'],
    ['materialStainlessSteel', '700', 'Edelstahl'],
    ['materialPlatedCopper', '800', 'beschicht. Kupfer'],
    ['materialHardenedSteel', '1500', 'gehärt. Stahl'],
    ['materialBimetal', '1700', 'Bimetall'],
    ['materialSiliconCarbide', '2500', 'Siliziumkarbid'],
    ['materialTungstenCarbide', '3000', 'Wolframkarbid'],
    ['materialRubyDiamond', '5000', 'Rubin/Diamant'],
  ];

  function materialLabel(english: string, german: string): string {
    return language === 'de' ? `${english} (${german})` : english;
  }
</script>

<div class="flex flex-col gap-4">
  <PageHeader title={m('title')} />

  <div class="flex items-center gap-2">
    {#each tags as tag (tag)}
      <Button
        variant={language === tag ? 'secondary' : 'ghost'}
        size="sm"
        onclick={() => applyLanguage(tag)}>{labels[tag]}</Button
      >
      >
    {/each}
  </div>

  <Card>
    <CardContent>
      <div class="flex flex-col gap-3 text-sm">
        <h3 class="font-semibold">{m('whatTitle')}</h3>
        <p>{m('what')}</p>

        <h3 class="mt-2 font-semibold">{m('countersTitle')}</h3>
        <p>{m('counters')}</p>
        <ol class="ml-4 list-decimal">
          <li>{m('countersStep1')}</li>
          <li>{m('countersStep2')}</li>
          <li>{m('countersStep3')}</li>
        </ol>

        <figure class="mt-2">
          <img
            src={printerTab}
            alt={m('captionTab')}
            class="w-full rounded-sm border border-border"
            loading="lazy"
          />
          <figcaption class="text-neutral-500 mt-1 text-xs">
            {m('captionTab')}
          </figcaption>
        </figure>

        <h3 class="mt-2 font-semibold">{m('setsTitle')}</h3>
        <p>{m('sets')}</p>
        <ul class="ml-4 list-disc">
          <li>{m('setsShipped')}</li>
          <li>{m('setsCopy')}</li>
          <li>{m('setsAssign')}</li>
          <li>{m('setsOff')}</li>
        </ul>

        <figure class="mt-2">
          <img
            src={schedules}
            alt={m('captionSchedules')}
            class="w-full rounded-sm border border-border"
            loading="lazy"
          />
          <figcaption class="text-neutral-500 mt-1 text-xs">
            {m('captionSchedules')}
          </figcaption>
        </figure>

        <h3 class="mt-2 font-semibold">{m('intervalsTitle')}</h3>
        <p>{m('intervals')}</p>

        <h3 class="mt-2 font-semibold">{m('nozzlesTitle')}</h3>
        <p>{m('nozzles')}</p>
        <Table class="mt-1 max-w-md">
          <TableHeader>
            <TableRow>
              <TableHead>{m('nozzleMaterial')}</TableHead>
              <TableHead>{m('nozzleLife')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {#each materials as material (material[0])}
              <TableRow>
                <TableCell
                  >{materialLabel(
                    t('maintenance.page.' + material[0]),
                    material[2],
                  )}</TableCell
                >
                <TableCell>{material[1]} h</TableCell>
              </TableRow>
            {/each}
          </TableBody>
        </Table>
        <p class="text-neutral-500">{t('maintenance.page.abrasiveHint')}</p>

        <figure class="mt-2">
          <img
            src={scheduleTasks}
            alt={m('captionTasks')}
            class="w-full rounded-sm border border-border"
            loading="lazy"
          />
          <figcaption class="text-neutral-500 mt-1 text-xs">
            {m('captionTasks')}
          </figcaption>
        </figure>

        <figure class="mt-2">
          <img
            src={printerSchedule}
            alt={m('captionSchedulePicker')}
            class="w-full rounded-sm border border-border"
            loading="lazy"
          />
          <figcaption class="text-neutral-500 mt-1 text-xs">
            {m('captionSchedulePicker')}
          </figcaption>
        </figure>

        <h3 class="mt-2 font-semibold">{m('toolsTitle')}</h3>
        <p>{m('tools')}</p>

        <h3 class="mt-2 font-semibold">{m('statusTitle')}</h3>
        <p>{m('status')}</p>
        <ul class="ml-4 flex list-none flex-col gap-1">
          <li>
            <span class="font-semibold text-state-done-text"
              >{m('statusOkWord')}</span
            >: {m('statusOkBody')}
          </li>
          <li>
            <span class="font-semibold text-warning-text"
              >{m('statusSoonWord')}</span
            >: {m('statusSoonBody')}
          </li>
          <li>
            <span class="font-semibold text-state-fault-text"
              >{m('statusDueWord')}</span
            >: {m('statusDueBody')}
          </li>
        </ul>

        <h3 class="mt-2 font-semibold">{m('cardTitle')}</h3>
        <p>{m('card')}</p>

        <figure class="mt-2">
          <img
            src={cardGallery}
            alt={m('captionGallery')}
            class="w-full rounded-sm border border-border"
            loading="lazy"
          />
          <figcaption class="text-neutral-500 mt-1 text-xs">
            {m('captionGallery')}
          </figcaption>
        </figure>

        <figure class="mt-2">
          <img
            src={cardSettings}
            alt={m('captionSettings')}
            class="w-full rounded-sm border border-border"
            loading="lazy"
          />
          <figcaption class="text-neutral-500 mt-1 text-xs">
            {m('captionSettings')}
          </figcaption>
        </figure>

        <figure class="mt-2">
          <img
            src={dashboard}
            alt={m('captionDashboard')}
            class="w-full rounded-sm border border-border"
            loading="lazy"
          />
          <figcaption class="text-neutral-500 mt-1 text-xs">
            {m('captionDashboard')}
          </figcaption>
        </figure>

        <h3 class="mt-2 font-semibold">{m('languageTitle')}</h3>
        <p>{m('language')}</p>
        <p class="text-neutral-500">{m('languageRestart')}</p>
        <p class="text-neutral-500">{m('languageContribute')}</p>
        <p>
          <a
            class="underline"
            href="https://github.com/SinglerGold2/maintenance"
            target="_blank"
            rel="noreferrer">{m('languageRepo')}</a
          >
        </p>

        <h3 class="mt-2 font-semibold">{m('whereTitle')}</h3>
        <p>{m('where')}</p>

        <h3 class="mt-2 font-semibold">{m('checkTitle')}</h3>
        <ul class="ml-4 list-disc">
          <li>{m('check1')}</li>
          <li>{m('check2')}</li>
          <li>{m('check3')}</li>
        </ul>
      </div>
    </CardContent>
  </Card>
</div>
