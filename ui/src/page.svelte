<script lang="ts">
  import {
    Button,
    Card,
    CardContent,
    Input,
    Label,
    PageHeader,
    RadioGroup,
    RadioGroupItem,
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

  interface Model {
    value: string;
    label: string;
    abrasiveSafe?: boolean;
  }

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
    { value: 'MK3', label: 'MK3' },
    { value: 'M3', label: 'M3' },
  ];

  // Ordered by how long each material lasts, shortest first. The ones a
  // carbon or glass filled filament eats through are marked.
  const nozzleMaterials: Model[] = [
    { value: 'brass', label: t('maintenance.page.materialBrass') },
    {
      value: 'stainless steel',
      label: t('maintenance.page.materialStainlessSteel'),
    },
    {
      value: 'plated copper',
      label: t('maintenance.page.materialPlatedCopper'),
    },
    {
      value: 'hardened steel',
      label: t('maintenance.page.materialHardenedSteel'),
      abrasiveSafe: true,
    },
    {
      value: 'bimetal',
      label: t('maintenance.page.materialBimetal'),
      abrasiveSafe: true,
    },
    {
      value: 'silicon carbide',
      label: t('maintenance.page.materialSiliconCarbide'),
      abrasiveSafe: true,
    },
    {
      value: 'tungsten carbide',
      label: t('maintenance.page.materialTungstenCarbide'),
      abrasiveSafe: true,
    },
    {
      value: 'ruby/diamond',
      label: t('maintenance.page.materialRubyDiamond'),
      abrasiveSafe: true,
    },
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
      sets = answer.sets;
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
      <div class="flex flex-col gap-2 text-sm">
        <p>{t('maintenance.page.introWhat')}</p>
        <p class="text-neutral-600">{t('maintenance.page.introHow')}</p>
        <p class="text-neutral-600">{t('maintenance.page.introValues')}</p>
        <p class="text-neutral-600">{t('maintenance.page.introNew')}</p>
      </div>

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
            {#if set.kb_url}
              <a
                class="text-neutral-500 ml-2 text-sm underline"
                href={set.kb_url}
                target="_blank"
                rel="noreferrer"
              >
                {t('maintenance.page.kb')}
              </a>
            {/if}
            <span class="text-neutral-500 ml-2 text-sm">
              {set.model === '*' ? t('maintenance.page.anyModel') : set.model}
              {#if set.builtin}· {t('maintenance.page.builtin')}{/if}
              {#if !set.enabled}· {t('maintenance.page.off')}{/if}
            </span>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <Button
              variant="secondary"
              size="sm"
              disabled={busy}
              onclick={() => toggle(set)}
            >
              {set.enabled
                ? t('maintenance.page.disable')
                : t('maintenance.page.enable')}
            </Button>
            <Button
              variant="secondary"
              size="sm"
              disabled={busy}
              onclick={() => duplicate(set)}
            >
              {t('maintenance.page.duplicate')}
            </Button>
            {#if !set.builtin}
              <Button
                variant="destructive"
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
                class="border-neutral-200 flex flex-wrap items-center justify-between gap-2 border-t pt-2"
              >
                <div class="flex min-w-60 flex-col">
                  <span class="text-sm">{task.name}</span>
                  {#if task.kind === 'nozzle'}
                    <RadioGroup
                      value={task.material ?? 'hardened steel'}
                      aria-label={t('maintenance.page.nozzleMaterial')}
                      onValueChange={(next: string) => {
                        if (task.id !== undefined) {
                          void act(() =>
                            api.updateSetTask(set.id, task.id ?? '', {
                              name: task.name,
                              kind: 'nozzle',
                              material: next,
                            }),
                          );
                        }
                      }}
                    >
                      <div class="mt-1 flex flex-wrap gap-3">
                        {#each nozzleMaterials as material (material.value)}
                          <span
                            class="rounded flex items-center gap-1 px-1 {material.abrasiveSafe ===
                            true
                              ? ''
                              : 'bg-red-500/15'}"
                          >
                            <RadioGroupItem
                              value={material.value}
                              id={`${task.id ?? set.id}-${material.value}`}
                            />
                            <Label
                              variant="option"
                              for={`${task.id ?? set.id}-${material.value}`}
                            >
                              {material.label}
                            </Label>
                          </span>
                        {/each}
                      </div>
                    </RadioGroup>
                    <p class="text-neutral-500 mt-1 text-xs">
                      {t('maintenance.page.abrasiveHint')}
                    </p>
                  {/if}
                </div>
                <span class="text-neutral-500 ml-auto text-sm">
                  {task.kind === 'days'
                    ? `${String(task.interval_days ?? 0)} d`
                    : `${String(task.interval_hours ?? 0)} h`}
                  {#if task.material && task.kind !== 'nozzle'}· {task.material}{/if}
                </span>
                {#if !set.builtin && task.kind !== 'nozzle'}
                  <Button
                    variant="secondary"
                    size="sm"
                    disabled={busy}
                    onclick={() =>
                      saveTask(
                        set,
                        task.id ?? '',
                        task.name,
                        task.kind,
                        (task.kind === 'days'
                          ? task.interval_days
                          : task.interval_hours) ?? 0,
                      )}
                  >
                    {t('maintenance.page.save')}
                  </Button>
                {/if}
                {#if !set.builtin}
                  <Button
                    variant="destructive"
                    size="sm"
                    disabled={busy}
                    onclick={() =>
                      task.id !== undefined && dropTask(set, task.id)}
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
