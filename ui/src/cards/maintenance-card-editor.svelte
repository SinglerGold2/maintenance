<script lang="ts">
  import {
    Label,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
  } from '@prusa/connect-ui/primitives';
  import type { CardEditorProps } from '@prusa/connect-ui/slots';

  import { heldHost } from '../host';
  import { t } from '../i18n.svelte';
  import { maintenanceOn, type Overview } from '../maintenance';

  let { config, onchange }: CardEditorProps = $props();

  const host = heldHost();
  const api = maintenanceOn(host);
  const uid = $props.id();

  let overview = $state<Overview | undefined>();

  $effect(() => {
    api
      .overview(host.signal)
      .then((answer) => (overview = answer))
      .catch(() => undefined);
  });

  // Empty value is the whole fleet. The name is stored beside the id so a card
  // still reads well after the printer is renamed.
  const options = $derived([
    { value: '', label: t('maintenance.card.all') },
    ...(overview?.devices ?? []).map((device) => ({
      value: device.id,
      label: device.name ?? device.model ?? device.id,
    })),
  ]);
</script>

<div class="grid gap-2">
  <Label for="{uid}-printer">{t('maintenance.card.printer')}</Label>
  <Select
    value={typeof config.device === 'string' ? config.device : ''}
    items={options}
    onValueChange={(next: string) => {
      onchange({ ...config, device: next });
    }}
  >
    <SelectTrigger id="{uid}-printer" class="w-full">
      <SelectValue placeholder={t('maintenance.card.all')} />
    </SelectTrigger>
    <SelectContent>
      {#each options as option (option.value)}
        <SelectItem value={option.value} label={option.label}>
          {option.label}
        </SelectItem>
      {/each}
    </SelectContent>
  </Select>
</div>
