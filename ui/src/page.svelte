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
  import { maintenanceOn, type Overview, type TaskView } from './maintenance';

  // A page takes no props: the route is the whole context.
  const host = heldHost();
  const t = host.t;
  const api = maintenanceOn(host);

  let overview = $state<Overview | undefined>();
  let fault = $state<FaultText | undefined>();
  let busy = $state(false);

  let drafts = $state<
    Record<
      string,
      { name: string; kind: 'hours' | 'days'; interval: number | undefined }
    >
  >({});
  let edits = $state<
    Record<string, { interval: number | undefined; material: string }>
  >({});

  function draftAt(deviceID: string): {
    name: string;
    kind: 'hours' | 'days';
    interval: number | undefined;
  } {
    return drafts[deviceID] ?? { name: '', kind: 'hours', interval: 200 };
  }

  function editAt(task: TaskView): {
    interval: number | undefined;
    material: string;
  } {
    return (
      edits[task.id] ?? {
        interval:
          task.kind === 'hours' ? task.interval_hours : task.interval_days,
        material: task.material ?? '',
      }
    );
  }

  async function load(): Promise<void> {
    try {
      const answer = await api.overview(host.signal);
      overview = answer;
      fault = undefined;

      for (const device of answer.devices) {
        drafts[device.id] ??= { name: '', kind: 'hours', interval: 200 };
      }
    } catch (thrown: unknown) {
      if (!host.signal.aborted) fault = describeFault(thrown);
    }
  }

  async function act(work: () => Promise<unknown>): Promise<void> {
    busy = true;

    try {
      await work();
      await load();
    } catch (thrown: unknown) {
      if (!host.signal.aborted) fault = describeFault(thrown);
    } finally {
      busy = false;
    }
  }

  function nameOf(device: {
    name?: string;
    model?: string;
    id: string;
  }): string {
    if (device.name !== undefined && device.name !== '') return device.name;
    if (device.model !== undefined && device.model !== '') return device.model;

    return device.id;
  }

  function num(value: number | undefined): number {
    return value ?? 0;
  }

  function kindOf(task: TaskView): 'hours' | 'days' {
    return task.kind === 'days' ? 'days' : 'hours';
  }

  function hours(value: number): string {
    return String(Math.round(value * 10) / 10) + ' h';
  }

  function intervalText(task: TaskView): string {
    if (kindOf(task) === 'hours') return hours(num(task.interval_hours));

    return String(num(task.interval_days)) + ' d';
  }

  function sinceText(task: TaskView): string {
    if (kindOf(task) === 'hours') return hours(task.elapsed);

    return String(Math.round(task.elapsed / 24)) + ' d';
  }

  function thresholdText(task: TaskView, alertPercent: number): string {
    const share = alertPercent / 100;

    if (kindOf(task) === 'hours')
      return hours(num(task.interval_hours) * share);

    return String(Math.round(num(task.interval_days) * share)) + ' d';
  }

  function statusText(task: TaskView): string {
    if (task.status === 'due') return t('maintenance.page.statusDue');
    if (task.status === 'soon') return t('maintenance.page.statusSoon');

    return t('maintenance.page.statusOk');
  }

  function statusColor(task: TaskView): string {
    if (task.status === 'due') return 'text-red-600';
    if (task.status === 'soon') return 'text-amber-600';

    return 'text-neutral-600';
  }

  function resetTask(deviceID: string, task: TaskView): void {
    void act(() =>
      api.saveTask({
        id: task.id,
        device_id: deviceID,
        name: task.name,
        kind: kindOf(task),
        interval_hours: num(task.interval_hours),
        interval_days: num(task.interval_days),
        material: task.material ?? '',
        last_done_hours: 0,
      }),
    );
  }

  function saveTask(deviceID: string, task: TaskView): void {
    const edit = editAt(task);
    const interval = edit.interval ?? 0;

    if (interval < 1) return;

    void act(async () => {
      await api.saveTask({
        id: task.id,
        device_id: deviceID,
        name: task.name,
        kind: kindOf(task),
        interval_hours: kindOf(task) === 'hours' ? interval : 0,
        interval_days: kindOf(task) === 'days' ? interval : 0,
        material: edit.material,
      });
      edits[task.id] = { interval, material: edit.material };
    });
  }

  function add(deviceID: string): void {
    const draft = drafts[deviceID];
    if (draft === undefined) return;

    const interval = draft.interval ?? 0;

    if (draft.name === '' || interval < 1) return;

    const body = {
      device_id: deviceID,
      name: draft.name,
      kind: draft.kind,
      interval_hours: draft.kind === 'hours' ? interval : 0,
      interval_days: draft.kind === 'days' ? interval : 0,
    };

    void act(async () => {
      await api.saveTask(body);
      drafts[deviceID] = { name: '', kind: draft.kind, interval };
    });
  }

  $effect(() => {
    void load();
  });
</script>

<div class="flex flex-col gap-4">
  <PageHeader title={t('maintenance.page.title')} />

  {#if fault}
    <FaultAlert {fault} />
  {:else if !overview}
    <Card>
      <CardContent>
        <Skeleton />
      </CardContent>
    </Card>
  {:else}
    {#if overview.devices.length === 0}
      <Card>
        <CardContent>
          <p class="text-neutral-600">{t('maintenance.page.empty')}</p>
        </CardContent>
      </Card>
    {/if}

    {#each overview.devices as device (device.id)}
      {@const draft = draftAt(device.id)}
      <Card>
        <CardContent>
          <div class="flex flex-col gap-3">
            <div>
              <h2 class="text-lg font-semibold">{nameOf(device)}</h2>
              <p class="text-neutral-600 text-sm">
                {t('maintenance.page.printed')}:
                <strong>{hours(device.hours)}</strong>
                ·
                {t('maintenance.page.prints')}:
                {device.prints}
              </p>
            </div>

            <div class="flex flex-col gap-3">
              {#each device.tasks as task (task.id)}
                {@const edit = editAt(task)}
                <div
                  class="border-neutral-200 flex flex-col gap-2 border-t pt-3"
                >
                  <div class="flex flex-wrap items-start justify-between gap-2">
                    <div>
                      <div class="font-medium">{task.name}</div>
                      <div class="text-neutral-600 text-sm">
                        {#if task.material}
                          {task.material} ·
                        {/if}
                        {t('maintenance.page.every')}
                        {intervalText(task)}
                        · {t('maintenance.page.sinceLast')}
                        {sinceText(task)}
                        · {t('maintenance.page.dueAt')}
                        {thresholdText(task, overview.alert_percent)}
                      </div>
                    </div>
                    <div class="flex flex-wrap items-center gap-2">
                      <span class="text-sm font-semibold {statusColor(task)}">
                        {statusText(task)}
                      </span>
                      <Button
                        variant="secondary"
                        size="sm"
                        disabled={busy}
                        onclick={() => void act(() => api.markDone(task.id))}
                      >
                        {t('maintenance.page.done')}
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        disabled={busy}
                        onclick={() => resetTask(device.id, task)}
                      >
                        {t('maintenance.page.reset')}
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        disabled={busy}
                        onclick={() => void act(() => api.deleteTask(task.id))}
                      >
                        {t('maintenance.page.remove')}
                      </Button>
                    </div>
                  </div>

                  <div class="flex flex-wrap items-center gap-2">
                    <Input
                      type="number"
                      min="1"
                      value={edit.interval}
                      aria-label={t('maintenance.page.interval')}
                      class="w-24"
                      oninput={(event: Event) => {
                        const target =
                          event.currentTarget as HTMLInputElement | null;
                        if (target !== null) {
                          edits[task.id] = {
                            ...edit,
                            interval: target.valueAsNumber,
                          };
                        }
                      }}
                    />
                    <Input
                      value={edit.material}
                      placeholder={t('maintenance.page.materialPlaceholder')}
                      aria-label={t('maintenance.page.material')}
                      class="w-48"
                      oninput={(event: Event) => {
                        const target =
                          event.currentTarget as HTMLInputElement | null;
                        if (target !== null) {
                          edits[task.id] = { ...edit, material: target.value };
                        }
                      }}
                    />
                    <Button
                      size="sm"
                      disabled={busy}
                      onclick={() => saveTask(device.id, task)}
                    >
                      {t('maintenance.page.save')}
                    </Button>
                  </div>
                </div>
              {/each}
            </div>

            <div
              class="border-neutral-200 flex flex-wrap items-end gap-2 border-t pt-3"
            >
              <Input
                value={draft.name}
                placeholder={t('maintenance.page.taskPlaceholder')}
                aria-label={t('maintenance.page.task')}
                class="max-w-xs"
                oninput={(event: Event) => {
                  const target = event.currentTarget as HTMLInputElement | null;
                  if (target !== null)
                    drafts[device.id] = { ...draft, name: target.value };
                }}
              />
              <Input
                type="number"
                min="1"
                value={draft.interval}
                aria-label={t('maintenance.page.interval')}
                class="w-24"
                oninput={(event: Event) => {
                  const target = event.currentTarget as HTMLInputElement | null;
                  if (target !== null) {
                    drafts[device.id] = {
                      ...draft,
                      interval: target.valueAsNumber,
                    };
                  }
                }}
              />
              <Button
                variant={draft.kind === 'hours' ? 'default' : 'secondary'}
                size="sm"
                onclick={() =>
                  (drafts[device.id] = { ...draft, kind: 'hours' })}
              >
                {t('maintenance.page.hoursUnit')}
              </Button>
              <Button
                variant={draft.kind === 'days' ? 'default' : 'secondary'}
                size="sm"
                onclick={() => (drafts[device.id] = { ...draft, kind: 'days' })}
              >
                {t('maintenance.page.daysUnit')}
              </Button>
              <Button size="sm" disabled={busy} onclick={() => add(device.id)}>
                {t('maintenance.page.addButton')}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    {/each}
  {/if}
</div>
