<script lang="ts">
  import { Badge, type BadgeVariant } from '@prusa/connect-ui/primitives';
  import type { CardProps } from '@prusa/connect-ui/slots';

  import { heldHost } from '../host';
  import { maintenanceOn, type Overview, type TaskView } from '../maintenance';

  let { config, title }: CardProps = $props();

  const host = heldHost();
  const t = host.t;
  const api = maintenanceOn(host);

  let overview = $state<Overview | undefined>();

  $effect(() => {
    api
      .overview(host.signal)
      .then((answer) => (overview = answer))
      .catch(() => undefined);
  });

  // The card's setting: a printer id, or empty for the whole fleet. A printer
  // that is gone leaves the card empty rather than silently showing another.
  const wanted = $derived(
    typeof config.device === 'string' ? config.device : '',
  );

  const devices = $derived(overview?.devices ?? []);

  const missing = $derived(
    wanted !== '' && !devices.some((d) => d.id === wanted),
  );

  interface Row {
    device: string;
    tasks: TaskView[];
    worst: string;
    set: string;
    hours: number | undefined;
  }

  const weights: Record<string, number> = { due: 0, soon: 1, ok: 2 };
  const weight = (status: string) => weights[status] ?? 2;

  const badges: Record<string, BadgeVariant> = {
    due: 'stateFault',
    soon: 'stateAttention',
    ok: 'stateDone',
  };

  const badgeOf = (status: string): BadgeVariant =>
    badges[status] ?? 'stateDone';

  const labelOf = (status: string): string => {
    if (status === 'due') return t('maintenance.page.statusDue');
    if (status === 'soon') return t('maintenance.page.statusSoon');
    return t('maintenance.page.statusOk');
  };

  // One printer when the setting names one, the whole fleet otherwise. Every
  // task is shown: a green line is the proof that this interval still holds.
  const rows = $derived<Row[]>(
    devices
      .filter((device) => wanted === '' || device.id === wanted)
      .map((device) => {
        const tasks = [...device.tasks].sort((a, b) => b.percent - a.percent);
        const worst = tasks.reduce(
          (worst, task) =>
            weight(task.status) < weight(worst) ? task.status : worst,
          'ok',
        );

        return {
          device: device.name ?? device.model ?? device.id,
          tasks,
          worst,
          set: device.set_name ?? '',
          hours: device.printer_hours,
        };
      })
      .sort((a, b) => weight(a.worst) - weight(b.worst)),
  );

  const single = $derived(wanted !== '' && rows.length === 1);
</script>

<div class="flex flex-col gap-2">
  <h3 class="text-base font-semibold">
    {title.trim() === '' ? t('maintenance.cards.maintenance.heading') : title}
  </h3>

  {#if !overview}
    <p class="text-neutral-600 text-sm">
      {t('maintenance.cards.maintenance.loading')}
    </p>
  {:else if missing}
    <p class="text-neutral-600 text-sm">
      {t('maintenance.cards.maintenance.missing')}
    </p>
  {:else if rows.length === 0}
    <p class="text-neutral-600 text-sm">
      {t('maintenance.cards.maintenance.none')}
    </p>
  {:else}
    {#if single && rows[0]}
      <p class="text-neutral-500 text-xs">
        {rows[0].set
          ? `${t('maintenance.cards.maintenance.set')}: ${rows[0].set}`
          : ''}
        {rows[0].hours !== undefined
          ? ` · ${t('maintenance.cards.maintenance.hours')}: ${String(Math.round(rows[0].hours))} h`
          : ''}
      </p>
    {/if}

    <ul class="flex max-h-64 flex-col gap-3 overflow-y-auto text-sm">
      {#each rows as row (row.device)}
        <li class="flex flex-col gap-1">
          <div class="flex items-center justify-between gap-2">
            <span class="flex min-w-0 items-center gap-2">
              <span
                class="size-2 shrink-0 rounded-full"
                class:bg-state-fault-edge={row.worst === 'due'}
                class:bg-state-attention-edge={row.worst === 'soon'}
                class:bg-state-done-edge={row.worst === 'ok'}
              ></span>
              <span class="truncate font-medium">{row.device}</span>
            </span>
            <Badge variant={badgeOf(row.worst)} size="sm"
              >{labelOf(row.worst)}</Badge
            >
          </div>
          <ul class="flex flex-col gap-0.5 pl-4 text-xs">
            {#each row.tasks as task (task.id)}
              <li class="flex items-center justify-between gap-2">
                <span class="flex min-w-0 items-center gap-2">
                  <span
                    class="size-1.5 shrink-0 rounded-full"
                    class:bg-state-fault-edge={task.status === 'due'}
                    class:bg-state-attention-edge={task.status === 'soon'}
                    class:bg-state-done-edge={task.status === 'ok'}
                  ></span>
                  <span class="text-neutral-600 truncate">{task.name}</span>
                </span>
                <span
                  class="shrink-0 tabular-nums"
                  class:text-state-fault-text={task.status === 'due'}
                  class:text-state-attention-text={task.status === 'soon'}
                  class:text-state-done-text={task.status === 'ok'}
                  >{Math.round(task.percent)}%</span
                >
              </li>
            {/each}
          </ul>
        </li>
      {/each}
    </ul>
  {/if}
</div>
