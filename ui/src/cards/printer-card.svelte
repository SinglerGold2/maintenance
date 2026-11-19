<script lang="ts">
  import type { CardProps } from '@prusa/connect-ui/slots';
  import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
  } from '@prusa/connect-ui/primitives';

  import { heldHost } from '../host';
  import { maintenanceOn, type Overview } from '../maintenance';

  let { title }: CardProps = $props();

  const host = heldHost();
  const t = host.t;
  const api = maintenanceOn(host);

  let overview = $state<Overview | undefined>();
  let selected = $state<string | undefined>();

  $effect(() => {
    api
      .overview(host.signal)
      .then((answer) => {
        overview = answer;
        selected ??= answer.devices[0]?.id;
      })
      .catch(() => undefined);
  });

  const devices = $derived(overview?.devices ?? []);

  const device = $derived(
    devices.find((candidate) => candidate.id === selected) ?? devices[0],
  );

  const options = $derived(
    devices.map((candidate) => ({
      value: candidate.id,
      label: candidate.name ?? candidate.model ?? candidate.id,
    })),
  );

  const tasks = $derived(
    [...(device?.tasks ?? [])].sort((a, b) => b.percent - a.percent),
  );
</script>

<div class="flex flex-col gap-2">
  <div class="flex items-center justify-between gap-2">
    <h3 class="text-base font-semibold">
      {title.trim() === '' ? t('maintenance.cards.printer.heading') : title}
    </h3>

    {#if devices.length > 1}
      <Select
        value={selected ?? ''}
        onValueChange={(value: string) => (selected = value)}
        items={options}
      >
        <SelectTrigger size="sm" class="max-w-40">
          <SelectValue placeholder={t('maintenance.cards.printer.choose')} />
        </SelectTrigger>
        <SelectContent>
          {#each options as option (option.value)}
            <SelectItem value={option.value}>{option.label}</SelectItem>
          {/each}
        </SelectContent>
      </Select>
    {/if}
  </div>

  {#if !overview}
    <p class="text-neutral-600 text-sm">
      {t('maintenance.cards.printer.loading')}
    </p>
  {:else if !device}
    <p class="text-neutral-600 text-sm">
      {t('maintenance.cards.printer.none')}
    </p>
  {:else}
    <p class="text-neutral-500 text-xs">
      {device.set_name
        ? `${t('maintenance.cards.printer.set')}: ${device.set_name}`
        : ''}
      {device.printer_hours !== undefined
        ? ` · ${t('maintenance.cards.printer.hours')}: ${String(Math.round(device.printer_hours))} h`
        : ''}
    </p>

    <ul class="flex max-h-56 flex-col gap-1 overflow-y-auto text-xs">
      {#each tasks as task (task.id)}
        <li class="flex items-center justify-between gap-2">
          <span class="flex min-w-0 items-center gap-2">
            <span
              class="size-1.5 shrink-0 rounded-full"
              class:bg-state-fault-edge={task.status === 'due'}
              class:bg-state-attention-edge={task.status === 'soon'}
              class:bg-state-done-edge={task.status === 'ok'}
            ></span>
            <span class="truncate">{task.name}</span>
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
  {/if}
</div>
