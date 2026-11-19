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

  const materials: [string, string][] = [
    ['materialBrass', '500'],
    ['materialStainlessSteel', '700'],
    ['materialPlatedCopper', '800'],
    ['materialHardenedSteel', '1500'],
    ['materialBimetal', '1700'],
    ['materialSiliconCarbide', '2500'],
    ['materialTungstenCarbide', '3000'],
    ['materialRubyDiamond', '5000'],
  ];
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
        <h3 class="font-medium">{m('whatTitle')}</h3>
        <p>{m('what')}</p>

        <h3 class="mt-2 font-medium">{m('countersTitle')}</h3>
        <p>{m('counters')}</p>
        <ol class="ml-4 list-decimal">
          <li>{m('countersStep1')}</li>
          <li>{m('countersStep2')}</li>
          <li>{m('countersStep3')}</li>
        </ol>

        <h3 class="mt-2 font-medium">{m('setsTitle')}</h3>
        <p>{m('sets')}</p>
        <ul class="ml-4 list-disc">
          <li>{m('setsShipped')}</li>
          <li>{m('setsCopy')}</li>
          <li>{m('setsAssign')}</li>
          <li>{m('setsOff')}</li>
        </ul>

        <h3 class="mt-2 font-medium">{m('intervalsTitle')}</h3>
        <p>{m('intervals')}</p>

        <h3 class="mt-2 font-medium">{m('nozzlesTitle')}</h3>
        <p>{m('nozzles')}</p>
        <Table class="mt-1 max-w-md">
          <TableHeader>
            <TableRow>
              <TableHead>{t('maintenance.page.nozzleMaterial')}</TableHead>
              <TableHead>{m('nozzleLife')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {#each materials as material (material[0])}
              <TableRow>
                <TableCell>{t('maintenance.page.' + material[0])}</TableCell>
                <TableCell>{material[1]} h</TableCell>
              </TableRow>
            {/each}
          </TableBody>
        </Table>
        <p class="text-neutral-500">{t('maintenance.page.abrasiveHint')}</p>

        <h3 class="mt-2 font-medium">{m('toolsTitle')}</h3>
        <p>{m('tools')}</p>

        <h3 class="mt-2 font-medium">{m('statusTitle')}</h3>
        <p>{m('status')}</p>
        <ul class="ml-4 list-disc">
          <li>{m('statusOk')}</li>
          <li>{m('statusSoon')}</li>
          <li>{m('statusDue')}</li>
        </ul>

        <h3 class="mt-2 font-medium">{m('cardTitle')}</h3>
        <p>{m('card')}</p>

        <h3 class="mt-2 font-medium">{m('whereTitle')}</h3>
        <p>{m('where')}</p>

        <h3 class="mt-2 font-medium">{m('checkTitle')}</h3>
        <ul class="ml-4 list-disc">
          <li>{m('check1')}</li>
          <li>{m('check2')}</li>
          <li>{m('check3')}</li>
        </ul>
      </div>
    </CardContent>
  </Card>
</div>
