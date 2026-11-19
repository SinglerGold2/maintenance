<script lang="ts">
  import {
    Button,
    ConfirmDialog,
    Input,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
  } from '@prusa/connect-ui/primitives';

  import { heldHost } from './host';
  import { t } from './i18n.svelte';
  import {
    maintenanceOn,
    type Overview,
    type Set,
    type TaskView,
  } from './maintenance';

  let { deviceId }: { deviceId: string } = $props();

  const host = heldHost();
  const api = maintenanceOn(host);

  let overview = $state<Overview | undefined>();
  let busy = $state(false);
  let printerDraft = $state('');
  let pendingDone = $state<TaskView | undefined>();
  let pendingToolReset = $state<number | undefined>();
  let sets = $state<Set[]>([]);
  let failed = $state('');

  const device = $derived(
    overview?.devices.find((entry) => entry.id === deviceId),
  );

  async function load(): Promise<void> {
    try {
      overview = await api.overview(host.signal);
      failed = '';
    } catch (error) {
      overview = undefined;
      failed = String(error);
      return;
    }

    // The schedule list is a convenience: when it cannot be read, the printer's
    // own page still has to show its hours and tasks.
    try {
      sets = (await api.sets(host.signal)).sets;
    } catch {
      sets = [];
    }
  }

  function assign(setID: string): void {
    if (setID === '') return;

    void act(() => api.assignSet({ device_id: deviceId, set_id: setID }));
  }

  const setChoices = $derived(
    sets
      .filter((entry) => entry.enabled === true)
      .map((entry) => ({ value: entry.id, label: entry.name })),
  );

  async function act(work: () => Promise<unknown>): Promise<void> {
    busy = true;

    try {
      await work();
      await load();
    } finally {
      busy = false;
    }
  }

  function num(value: number | undefined): number {
    return value ?? 0;
  }

  function kindOf(task: TaskView): 'hours' | 'days' {
    return task.kind === 'days' ? 'days' : 'hours';
  }

  function hours(value: number): string {
    const total = Math.round(value * 3600);
    const h = Math.floor(total / 3600);
    const m = Math.floor((total % 3600) / 60);
    const s = total % 60;
    const parts: string[] = [];

    if (h > 0) parts.push(String(h) + 'h');
    if (m > 0) parts.push(String(m) + 'm');
    if (s > 0 || parts.length === 0) parts.push(String(s) + 's');

    return parts.join(' ');
  }

  function intervalText(task: TaskView): string {
    if (kindOf(task) === 'hours') return hours(num(task.interval_hours));

    return String(num(task.interval_days)) + ' d';
  }

  function statusText(task: TaskView): string {
    if (task.status === 'due') return t('maintenance.page.statusDue');
    if (task.status === 'soon') return t('maintenance.page.statusSoon');

    return t('maintenance.page.statusOk');
  }

  function sinceText(task: TaskView): string {
    if (kindOf(task) === 'hours') return hours(task.elapsed);

    return String(Math.round(task.elapsed / 24)) + ' d';
  }

  function parseHours(text: string): number {
    const pattern = /(\d+(?:[.,]\d+)?)\s*([wdh])/g;
    let total = 0;
    let matched = false;
    let match = pattern.exec(text.toLowerCase());

    while (match !== null) {
      const value = Number(match[1]?.replace(',', '.') ?? '0');
      const unit = match[2];

      if (unit === 'w') total += value * 7 * 24;
      if (unit === 'd') total += value * 24;
      if (unit === 'h') total += value;

      matched = true;
      match = pattern.exec(text.toLowerCase());
    }

    if (matched) return total;

    const plain = Number(text.trim().replace(',', '.'));

    return Number.isFinite(plain) ? plain : Number.NaN;
  }

  function savePrinterHours(): void {
    const typed = parseHours(printerDraft);
    if (!Number.isFinite(typed) || typed < 0) return;

    void act(async () => {
      await api.savePrinterHours(deviceId, typed);
      printerDraft = '';
    });
  }

  $effect(() => {
    void load();
  });
</script>

{#if device === undefined}
  <p class="text-neutral-600 text-sm">{t('maintenance.tab.waiting')}</p>
  {#if failed !== ''}
    <p class="text-xs">{failed}</p>
  {/if}
{:else}
  <div class="flex flex-col gap-4">
    <div class="text-neutral-600 flex items-center gap-2 text-sm">
      <span>{t('maintenance.tab.set')}</span>
      <Select
        value={device.set_id ?? ''}
        items={setChoices}
        disabled={busy}
        onValueChange={assign}
      >
        <SelectTrigger size="sm" class="w-56">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {#each setChoices as choice (choice.value)}
            <SelectItem value={choice.value} label={choice.label}>
              {choice.label}
            </SelectItem>
          {/each}
        </SelectContent>
      </Select>
    </div>
    <p class="text-neutral-500 -mt-2 text-sm">{t('maintenance.tab.setHint')}</p>

    <div>
      <p class="text-neutral-600 text-sm">
        {t('maintenance.page.printed')}:
        <strong>{hours(device.hours)}</strong>
        ·
        {t('maintenance.page.prints')}:
        {device.prints}
      </p>
      {#if device.printer_hours !== undefined}
        <p class="text-neutral-600 text-sm">
          {t('maintenance.page.printerSays')}:
          <strong>{hours(device.printer_hours)}</strong>
          ·
          {t('maintenance.page.serverTracked')}:
          {hours(device.hours)}
        </p>
      {/if}
      <div class="mt-2 flex flex-wrap items-center gap-2">
        <Input
          value={printerDraft}
          placeholder={t('maintenance.page.printerHoursPlaceholder')}
          aria-label={t('maintenance.page.printerHours')}
          class="w-40"
          oninput={(event: Event) => {
            const target = event.currentTarget as HTMLInputElement | null;
            if (target !== null) printerDraft = target.value;
          }}
        />
        <Button
          variant="secondary"
          size="sm"
          disabled={busy}
          onclick={savePrinterHours}
        >
          {t('maintenance.page.save')}
        </Button>
        <span class="text-neutral-500 text-xs"
          >{t('maintenance.page.printerHoursHint')}</span
        >
      </div>
    </div>

    {#if device.tools && device.tools.length > 0}
      <div>
        <h3 class="font-medium">{t('maintenance.tab.tools')}</h3>
        <p class="text-neutral-600 text-sm">{t('maintenance.tab.toolsHint')}</p>
        <ul class="mt-1 flex flex-wrap gap-3 text-sm">
          {#each device.tools as tool (tool.tool)}
            <li
              class="border-neutral-200 rounded flex items-center gap-2 border px-2 py-1"
            >
              <span>T{tool.tool}: <strong>{hours(tool.hours)}</strong></span>
              <Button
                variant="secondary"
                size="sm"
                disabled={busy}
                onclick={() => (pendingToolReset = tool.tool)}
              >
                {t('maintenance.page.reset')}
              </Button>
            </li>
          {/each}
        </ul>
      </div>
    {/if}

    <div class="flex flex-col gap-2">
      {#each device.tasks as task (task.id)}
        <div
          class="border-neutral-200 flex flex-wrap items-center justify-between gap-2 border-t pt-2"
        >
          <div>
            <div class="font-medium">{task.name}</div>
            <div class="text-neutral-600 text-sm">
              {t('maintenance.page.every')}
              {intervalText(task)}
              · {t('maintenance.page.sinceLast')}
              {sinceText(task)}
            </div>
          </div>
          <div class="flex items-center gap-2">
            <span
              class="text-sm font-semibold"
              class:text-state-fault-text={task.status === 'due'}
              class:text-warning-text={task.status === 'soon'}
              class:text-state-done-text={task.status === 'ok'}
            >
              {statusText(task)}
            </span>
            <Button
              variant="secondary"
              size="sm"
              disabled={busy}
              onclick={() => (pendingDone = task)}
            >
              {t('maintenance.page.done')}
            </Button>
          </div>
        </div>
      {/each}
    </div>
  </div>

  <ConfirmDialog
    open={pendingToolReset !== undefined}
    onOpenChange={(open) => {
      if (!open) pendingToolReset = undefined;
    }}
    title={t('maintenance.tab.confirmResetTitle')}
    description={t('maintenance.tab.confirmResetBody')}
    confirmLabel={t('maintenance.tab.confirmResetYes')}
    variant="confirm"
    onConfirm={() => {
      const tool = pendingToolReset;
      pendingToolReset = undefined;
      if (tool !== undefined) void act(() => api.resetTool(deviceId, tool));
    }}
  />

  <ConfirmDialog
    open={pendingDone !== undefined}
    onOpenChange={(open) => {
      if (!open) pendingDone = undefined;
    }}
    title={t('maintenance.page.confirmDoneTitle')}
    description={t('maintenance.page.confirmDoneBody')}
    confirmLabel={t('maintenance.page.confirmDoneYes')}
    variant="confirm"
    onConfirm={() => {
      const task = pendingDone;
      pendingDone = undefined;
      if (task !== undefined) void act(() => api.markDone(task.id));
    }}
  />
{/if}
