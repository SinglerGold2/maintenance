<script lang="ts">
  import { describeFault, type FaultText } from '@prusa/connect-ui/i18n';
  import {
    Button,
    Card,
    CardContent,
    FaultAlert,
    Input,
    PageHeader,
    Skeleton,
  } from '@prusa/connect-ui/primitives';

  import { heldHost } from './host';
  import { maintenanceOn } from './maintenance';

  interface TemplateTask {
    id: string;
    name: string;
    kind: string;
    interval_hours?: number;
    interval_days?: number;
    material?: string;
  }

  // A page takes no props: the route is the whole context.
  const host = heldHost();
  const t = host.t;
  const api = maintenanceOn(host);

  let schedule = $state<TemplateTask[] | undefined>();
  let fault = $state<FaultText | undefined>();
  let busy = $state(false);
  let edits = $state<Record<string, number | undefined>>({});

  function kindOf(task: TemplateTask): 'hours' | 'days' {
    return task.kind === 'days' ? 'days' : 'hours';
  }

  function intervalOf(task: TemplateTask): number | undefined {
    return (
      edits[task.id] ??
      (kindOf(task) === 'hours' ? task.interval_hours : task.interval_days)
    );
  }

  function unitOf(task: TemplateTask): string {
    return kindOf(task) === 'hours' ? 'h' : 'd';
  }

  async function load(): Promise<void> {
    try {
      const answer = await api.template(host.signal);
      schedule = answer.tasks;
      fault = undefined;
    } catch (thrown: unknown) {
      if (!host.signal.aborted) fault = describeFault(thrown);
    }
  }

  async function save(task: TemplateTask): Promise<void> {
    const interval = intervalOf(task);
    if (interval === undefined || interval < 1) return;

    busy = true;

    try {
      await api.saveTemplate({
        id: task.id,
        name: task.name,
        kind: kindOf(task),
        interval_hours: kindOf(task) === 'hours' ? interval : 0,
        interval_days: kindOf(task) === 'days' ? interval : 0,
        material: task.material ?? '',
      });
      edits = { ...edits, [task.id]: undefined };
      await load();
    } catch (thrown: unknown) {
      if (!host.signal.aborted) fault = describeFault(thrown);
    } finally {
      busy = false;
    }
  }

  $effect(() => {
    void load();
  });
</script>

<div class="flex flex-col gap-4">
  <PageHeader title={t('maintenance.page.title')} />

  {#if fault}
    <FaultAlert {fault} />
  {:else if !schedule}
    <Card>
      <CardContent>
        <Skeleton />
      </CardContent>
    </Card>
  {:else}
    <p class="text-neutral-600 text-sm">{t('maintenance.page.templateNote')}</p>

    {#each schedule as task (task.id)}
      <Card>
        <CardContent>
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <div class="font-medium">{task.name}</div>
              <div class="text-neutral-600 text-sm">
                {#if task.material}
                  {task.material} ·
                {/if}
                {t('maintenance.page.every')}
                {intervalOf(task)}
                {unitOf(task)}
              </div>
            </div>
            <div class="flex items-center gap-2">
              <Input
                type="number"
                min="1"
                value={intervalOf(task)}
                aria-label={t('maintenance.page.interval')}
                class="w-24"
                oninput={(event: Event) => {
                  const target = event.currentTarget as HTMLInputElement | null;
                  if (target !== null) edits[task.id] = target.valueAsNumber;
                }}
              />
              <Button size="sm" disabled={busy} onclick={() => void save(task)}>
                {t('maintenance.page.save')}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    {/each}
  {/if}
</div>
