<script lang="ts">
  import {
    Button,
    Card,
    CardContent,
    Input,
    PageHeader,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
  } from '@prusa/connect-ui/primitives';

  import { heldHost } from './host';
  import { maintenanceOn, type Set } from './maintenance';

  const host = heldHost();
  const t = host.t;
  const signal = host.signal;
  const api = maintenanceOn(host);

  type Model = { value: string; label: string };

  let sets = $state<Set[]>([]);
  let busy = $state(false);
  let notice = $state('');
  let newName = $state('');
  let newModel = $state('*');
  let expanded = $state<string | undefined>();

  let draftName = $state('');
  let draftKind = $state('hours');
  let draftInterval = $state(200);
  let draftMaterial = $state('');

  const models: Model[] = [
    { value: '*', label: t('maintenance.page.anyModel') },
    { value: 'COREONE', label: 'Core One' },
    { value: 'COREONEL', label: 'Core One L' },
    { value: 'HT90', label: 'HT90' },
    { value: 'XL', label: 'XL' },
    { value: 'MK4S', label: 'MK4S' },
    { value: 'M3', label: 'M3' },
  ];

  const kinds: Model[] = [
    { value: 'hours', label: t('maintenance.page.hours') },
    { value: 'days', label: t('maintenance.page.days') },
  ];

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

  function addTask(set: Set) {
    if (draftName.trim() === '') return;

    void act(() =>
      api.addSetTask(set.id, {
        name: draftName,
        kind: draftKind,
        interval_hours: draftKind === 'hours' ? draftInterval : 0,
        interval_days: draftKind === 'days' ? draftInterval : 0,
        material: draftMaterial,
      }),
    ).then(() => {
      draftName = '';
      draftMaterial = '';
    });
  }

  function saveTask(
    set: Set,
    taskID: string,
    name: string,
    kind: string,
    interval: number,
  ) {
    void act(() =>
      api.updateSetTask(set.id, taskID, {
        name,
        kind,
        interval_hours: kind === 'hours' ? interval : 0,
        interval_days: kind === 'days' ? interval : 0,
        material: '',
      }),
    );
  }

  function dropTask(set: Set, taskID: string) {
    void act(() => api.deleteSetTask(set.id, taskID));
  }
</script>

<div class="flex flex-col gap-4">
  <PageHeader title={t('maintenance.page.title')} />

  <Card>
    <CardContent>
      <p class="text-neutral-600 text-sm">{t('maintenance.page.setsHint')}</p>

      <div class="mt-3 flex flex-wrap items-end gap-2">
        <Input
          value={newName}
          placeholder={t('maintenance.page.newName')}
          aria-label={t('maintenance.page.newName')}
          class="w-60"
          oninput={(event: Event) => {
            const target = event.currentTarget as HTMLInputElement | null;
            if (target !== null) newName = target.value;
          }}
        />
        <Select
          value={newModel}
          items={models}
          onValueChange={(next: string) => {
            newModel = next;
          }}
        >
          <SelectTrigger size="sm" class="w-40">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {#each models as model (model.value)}
              <SelectItem value={model.value} label={model.label}>
                {model.label}
              </SelectItem>
            {/each}
          </SelectContent>
        </Select>
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
                  <Button
                    variant="ghost"
                    size="sm"
                    disabled={busy}
                    onclick={() => dropTask(set, task.id)}
                  >
                    {t('maintenance.page.remove')}
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    disabled={busy}
                    onclick={() =>
                      saveTask(
                        set,
                        task.id,
                        task.name,
                        task.kind,
                        task.kind === 'days'
                          ? task.interval_days
                          : task.interval_hours,
                      )}
                  >
                    {t('maintenance.page.save')}
                  </Button>
                {/if}
              </div>
            {/each}

            {#if !set.builtin}
              <div
                class="border-neutral-200 flex flex-wrap items-end gap-2 border-t pt-2"
              >
                <Input
                  value={draftName}
                  placeholder={t('maintenance.page.taskName')}
                  aria-label={t('maintenance.page.taskName')}
                  class="w-56"
                  oninput={(event: Event) => {
                    const target =
                      event.currentTarget as HTMLInputElement | null;
                    if (target !== null) draftName = target.value;
                  }}
                />
                <Select
                  value={draftKind}
                  items={kinds}
                  onValueChange={(next: string) => {
                    draftKind = next;
                  }}
                >
                  <SelectTrigger size="sm" class="w-28">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {#each kinds as kind (kind.value)}
                      <SelectItem value={kind.value} label={kind.label}>
                        {kind.label}
                      </SelectItem>
                    {/each}
                  </SelectContent>
                </Select>
                <Input
                  value={String(draftInterval)}
                  aria-label={t('maintenance.page.interval')}
                  class="w-24"
                  oninput={(event: Event) => {
                    const target =
                      event.currentTarget as HTMLInputElement | null;
                    if (target !== null) draftInterval = Number(target.value);
                  }}
                />
                <Input
                  value={draftMaterial}
                  placeholder={t('maintenance.page.material')}
                  aria-label={t('maintenance.page.material')}
                  class="w-40"
                  oninput={(event: Event) => {
                    const target =
                      event.currentTarget as HTMLInputElement | null;
                    if (target !== null) draftMaterial = target.value;
                  }}
                />
                <Button
                  variant="secondary"
                  size="sm"
                  disabled={busy}
                  onclick={() => addTask(set)}
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
