<script lang="ts">
  import type { CardProps } from '@prusa/connect-ui/slots';

  import { heldHost } from '../host';
  import { maintenanceOn, type Overview } from '../maintenance';

  let { title }: CardProps = $props();

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

  const due = $derived(
    (overview?.devices ?? []).flatMap((device) =>
      device.tasks
        .filter((task) => task.status !== 'ok')
        .map((task) => ({
          device: device.name ?? device.model ?? device.id,
          task,
        })),
    ),
  );
</script>

<div class="flex flex-col gap-2">
  <h3 class="text-base font-semibold">
    {title.trim() === '' ? t('maintenance.cards.due.heading') : title}
  </h3>

  {#if !overview}
    <p class="text-neutral-600 text-sm">{t('maintenance.cards.due.loading')}</p>
  {:else if due.length === 0}
    <p class="text-neutral-600 text-sm">{t('maintenance.cards.due.none')}</p>
  {:else}
    <ul class="flex flex-col gap-1 text-sm">
      {#each due as row (row.device + row.task.id)}
        <li class="flex justify-between gap-2">
          <span>{row.task.name}</span>
          <span
            class:font-semibold={true}
            class:text-red-600={row.task.status === 'due'}
            class:text-amber-600={row.task.status === 'soon'}
          >
            {row.task.status === 'due'
              ? t('maintenance.page.statusDue')
              : t('maintenance.page.statusSoon')}
          </span>
        </li>
      {/each}
    </ul>
  {/if}
</div>
