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

  function draftOf(deviceID: string): {
    name: string;
    kind: 'hours' | 'days';
    interval: number | undefined;
  } {
    drafts[deviceID] ??= { name: '', kind: 'hours', interval: 200 };

    return drafts[deviceID];
  }

  async function load(): Promise<void> {
    try {
      overview = await api.overview(host.signal);
      fault = undefined;
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

  function intervalText(task: TaskView): string {
    if (task.kind === 'hours') return String(task.interval_hours) + ' h';

    return String(task.interval_days) + ' d';
  }

  function statusText(task: TaskView): string {
    const percent = String(task.percent) + '%';

    if (task.status === 'due')
      return percent + ' · ' + t('maintenance.page.due');
    if (task.status === 'soon')
      return percent + ' · ' + t('maintenance.page.soon');

    return (
      percent +
      ' · ' +
      String(task.remaining) +
      ' ' +
      t('maintenance.page.remaining')
    );
  }

  function add(deviceID: string): void {
    const draft = draftOf(deviceID);
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
      {@const draft = draftOf(device.id)}
      <Card>
        <CardContent>
          <div class="flex flex-col gap-3">
            <div>
              <h2 class="text-lg font-semibold">{nameOf(device)}</h2>
              <p class="text-neutral-600 text-sm">
                {device.model} · {t('maintenance.page.hours')}: {device.hours} ·
                {t('maintenance.page.prints')}: {device.prints}
              </p>
            </div>

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
                    </div>
                  </div>
                  <div class="flex items-center gap-2">
                    <span
                      class="text-sm"
                      class:font-semibold={task.status !== 'ok'}
                      class:text-red-600={task.status === 'due'}
                      class:text-amber-600={task.status === 'soon'}
                    >
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
                      onclick={() => void act(() => api.deleteTask(task.id))}
                    >
                      {t('maintenance.page.remove')}
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
