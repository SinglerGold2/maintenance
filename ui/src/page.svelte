<script lang="ts">
  import {
    Button,
    Card,
    CardContent,
    PageHeader,
  } from '@prusa/connect-ui/primitives';

  import { heldHost } from './host';
  import { maintenanceOn, type Set } from './maintenance';

  const host = heldHost();
  const t = host.t;
  const signal = host.signal;
  const api = maintenanceOn(host);

  let sets = $state<Set[]>([]);
  let busy = $state(false);
  let notice = $state('');
  let newName = $state('');
  let newModel = $state('*');
  let expanded = $state<string | undefined>();

  $effect(() => {
    void load();
  });

  async function load() {
    try {
      const answer = await api.sets(signal);
      sets = answer.sets ?? [];
    } catch (error) {
      notice = String(error);
    }
  }

  async function act(work: () => Promise<unknown>) {
    busy = true;
    notice = '';

    try {
      await work();
      await load();
    } catch (error) {
      notice = String(error);
    } finally {
      busy = false;
    }
  }

  function toggle(set: Set) {
    void act(() =>
      api.updateSet(set.id, {
        name: set.name,
        model: set.model,
        enabled: !set.enabled,
      }),
    );
  }

  function duplicate(set: Set) {
    void act(() =>
      api.createSet({
        name: set.name + ' (copy)',
        model: set.model,
        copy_of: set.id,
      }),
    );
  }

  function remove(set: Set) {
    void act(() => api.deleteSet(set.id));
  }

  function create() {
    const name = newName.trim();
    if (name === '') return;

    void act(() => api.createSet({ name, model: newModel })).then(() => {
      newName = '';
    });
  }

  function addTask(
    set: Set,
    name: string,
    kind: string,
    interval: number,
    material: string,
  ) {
    if (name.trim() === '') return;

    void act(() =>
      api.addSetTask(set.id, {
        name,
        kind,
        interval_hours: kind === 'hours' ? interval : 0,
        interval_days: kind === 'days' ? interval : 0,
        material,
      }),
    );
  }

  function changeTask(
    set: Set,
    taskID: string,
    name: string,
    kind: string,
    interval: number,
    material: string,
  ) {
    void act(() =>
      api.updateSetTask(set.id, taskID, {
        name,
        kind,
        interval_hours: kind === 'hours' ? interval : 0,
        interval_days: kind === 'days' ? interval : 0,
        material,
      }),
    );
  }

  function dropTask(set: Set, taskID: string) {
    void act(() => api.deleteSetTask(set.id, taskID));
  }

  let draftName = $state('');
  let draftKind = $state('hours');
  let draftInterval = $state(200);
  let draftMaterial = $state('');

  function resetDraft() {
    draftName = '';
    draftKind = 'hours';
    draftInterval = 200;
    draftMaterial = '';
  }
</script>

<div class="flex flex-col gap-4">
  <PageHeader title={t('maintenance.page.title')} />

  <Card>
    <CardContent>
      <p class="text-neutral-600 text-sm">{t('maintenance.page.setsHint')}</p>

      <div class="mt-3 flex flex-wrap items-end gap-2">
        <input
          class="border-neutral-300 rounded w-60 border px-2 py-1 text-sm"
          placeholder={t('maintenance.page.newName')}
          bind:value={newName}
        />
        <select
          class="border-neutral-300 rounded border px-2 py-1 text-sm"
          bind:value={newModel}
        >
          <option value="*">{t('maintenance.page.anyModel')}</option>
          <option value="COREONE">Core One</option>
          <option value="COREONEL">Core One L</option>
          <option value="HT90">HT90</option>
          <option value="XL">XL</option>
          <option value="MK4S">MK4S</option>
          <option value="M3">M3</option>
        </select>
        <Button size="sm" disabled={busy} onclick={create}>
          {t('maintenance.page.newSet')}
        </Button>
      </div>

      {#if notice !== ''}
        <p class="mt-2 text-sm">{notice}</p>
      {/if}
    </CardContent>
  </Card>

  {#each sets as set (set.id)}
    <Card>
      <CardContent>
        <div class="flex flex-wrap items-center justify-between gap-2">
          <div>
            <span class="font-medium">{set.name}</span>
            <span class="text-neutral-500 ml-2 text-sm">
              {set.model === '*' ? t('maintenance.page.anyModel') : set.model}
              {#if set.builtin}· {t('maintenance.page.builtin')}{/if}
              {#if !set.enabled}· {t('maintenance.page.off')}{/if}
            </span>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              disabled={busy}
              onclick={() => toggle(set)}
            >
              {set.enabled
                ? t('maintenance.page.disable')
                : t('maintenance.page.enable')}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              disabled={busy}
              onclick={() => duplicate(set)}
            >
              {t('maintenance.page.duplicate')}
            </Button>
            {#if !set.builtin}
              <Button
                variant="ghost"
                size="sm"
                disabled={busy}
                onclick={() => remove(set)}
              >
                {t('maintenance.page.remove')}
              </Button>
            {/if}
            <Button
              variant="ghost"
              size="sm"
              onclick={() =>
                (expanded = expanded === set.id ? undefined : set.id)}
            >
              {expanded === set.id
                ? t('maintenance.page.hideTasks')
                : t('maintenance.page.showTasks')}
            </Button>
          </div>
        </div>

        {#if expanded === set.id}
          <div class="mt-3 flex flex-col gap-2">
            {#each set.tasks as task (task.id)}
              <div
                class="border-neutral-200 flex flex-wrap items-center gap-2 border-t pt-2"
              >
                <span class="text-sm">{task.name}</span>
                <span class="text-neutral-500 text-sm">
                  {task.kind === 'days'
                    ? `${task.interval_days} d`
                    : `${task.interval_hours} h`}
                  {#if task.material}· {task.material}{/if}
                </span>
                {#if !set.builtin}
                  <select
                    class="border-neutral-300 rounded border px-2 py-1 text-sm"
                    bind:value={task.kind}
                  >
                    <option value="hours">{t('maintenance.page.hours')}</option>
                    <option value="days">{t('maintenance.page.days')}</option>
                  </select>
                  {#if task.kind === 'days'}
                    <input
                      class="border-neutral-300 rounded w-24 border px-2 py-1 text-sm"
                      type="number"
                      bind:value={task.interval_days}
                    />
                  {:else}
                    <input
                      class="border-neutral-300 rounded w-24 border px-2 py-1 text-sm"
                      type="number"
                      bind:value={task.interval_hours}
                    />
                  {/if}
                  <Button
                    variant="ghost"
                    size="sm"
                    disabled={busy}
                    onclick={() =>
                      changeTask(
                        set,
                        task.id,
                        task.name,
                        task.kind,
                        task.kind === 'days'
                          ? task.interval_days
                          : task.interval_hours,
                        task.material ?? '',
                      )}
                  >
                    {t('maintenance.page.save')}
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    disabled={busy}
                    onclick={() => dropTask(set, task.id)}
                  >
                    {t('maintenance.page.remove')}
                  </Button>
                {/if}
              </div>
            {/each}

            {#if !set.builtin}
              <div
                class="border-neutral-200 flex flex-wrap items-end gap-2 border-t pt-2"
              >
                <input
                  class="border-neutral-300 rounded w-56 border px-2 py-1 text-sm"
                  placeholder={t('maintenance.page.taskName')}
                  bind:value={draftName}
                />
                <select
                  class="border-neutral-300 rounded border px-2 py-1 text-sm"
                  bind:value={draftKind}
                >
                  <option value="hours">{t('maintenance.page.hours')}</option>
                  <option value="days">{t('maintenance.page.days')}</option>
                </select>
                <input
                  class="border-neutral-300 rounded w-24 border px-2 py-1 text-sm"
                  type="number"
                  bind:value={draftInterval}
                />
                <input
                  class="border-neutral-300 rounded w-40 border px-2 py-1 text-sm"
                  placeholder={t('maintenance.page.material')}
                  bind:value={draftMaterial}
                />
                <Button
                  variant="secondary"
                  size="sm"
                  disabled={busy}
                  onclick={() => {
                    addTask(
                      set,
                      draftName,
                      draftKind,
                      draftInterval,
                      draftMaterial,
                    );
                    resetDraft();
                  }}
                >
                  {t('maintenance.page.addTask')}
                </Button>
              </div>
            {/if}
          </div>
        {/if}
      </CardContent>
    </Card>
  {/each}
</div>
