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
  import german from './manual-de.json';
  import { heldHost } from './host';

  const host = heldHost();
  const t = host.t;

  // The manual is prose, so it carries its own two languages instead of
  // following the interface language of the whole app.
  let language = $state<'en' | 'de'>('en');

  type GermanKey = keyof typeof german;

  function m(key: string): string {
    if (language === 'de' && key in german) {
      return german[key as GermanKey];
    }

    return t('maintenance.manual.' + key);
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
    <Button
      variant={language === 'en' ? 'secondary' : 'ghost'}
      size="sm"
      onclick={() => (language = 'en')}
    >
      {t('maintenance.manual.langEn')}
    </Button>
    <Button
      variant={language === 'de' ? 'secondary' : 'ghost'}
      size="sm"
      onclick={() => (language = 'de')}
    >
      {t('maintenance.manual.langDe')}
    </Button>
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
