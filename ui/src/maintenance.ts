// The service layer: this module's own routes, reached through the client the
// host handed over. Nothing here names a URL or a token.
import type { ModuleHost } from '@prusa/connect-ui/slots';
import { bind, type operations } from 'maintenance-client';

import { applyLanguage } from './i18n.svelte';

/** What GET /overview answers, as openapi.json declares it. */
export type Overview =
  operations['read-overview']['responses'][200]['content']['application/json'];

/** One task with its counters. */
export type TaskView = Overview['devices'][number]['tasks'][number];

/** The route set, bound onto the host's client. */
export const maintenanceOn = (host: ModuleHost) => {
  const routes = bind(host.client);

  return {
    overview: async (signal?: AbortSignal): Promise<Overview> => {
      const answer = await routes.readOverview({}, signal);
      applyLanguage(answer.language);

      return answer;
    },
    saveTask: (
      body: operations['save-task']['requestBody']['content']['application/json'],
      signal?: AbortSignal,
    ) => routes.saveTask({ body }, signal),
    markDone: (id: string, signal?: AbortSignal) =>
      routes.markDone({ body: { id } }, signal),
    deleteTask: (id: string, signal?: AbortSignal) =>
      routes.deleteTask({ query: { id } }, signal),
    resetTool: (deviceID: string, tool: number, signal?: AbortSignal) =>
      routes.resetTool({ body: { device_id: deviceID, tool } }, signal),
    savePrinterHours: (deviceID: string, hours: number, signal?: AbortSignal) =>
      routes.savePrinterHours({ body: { device_id: deviceID, hours } }, signal),
    template: (signal?: AbortSignal) => routes.readTemplate({}, signal),
    saveTemplate: (
      body: operations['save-template']['requestBody']['content']['application/json'],
      signal?: AbortSignal,
    ) => routes.saveTemplate({ body }, signal),
    sets: (signal?: AbortSignal) => routes.listSets({}, signal),
    createSet: (
      body: operations['create-set']['requestBody']['content']['application/json'],
      signal?: AbortSignal,
    ) => routes.createSet({ body }, signal),
    updateSet: (
      id: string,
      body: operations['update-set']['requestBody']['content']['application/json'],
      signal?: AbortSignal,
    ) => routes.updateSet({ path: { id }, body }, signal),
    deleteSet: (id: string, signal?: AbortSignal) =>
      routes.deleteSet({ path: { id } }, signal),
    addSetTask: (
      id: string,
      body: operations['add-set-task']['requestBody']['content']['application/json'],
      signal?: AbortSignal,
    ) => routes.addSetTask({ path: { id }, body }, signal),
    updateSetTask: (
      id: string,
      task: string,
      body: operations['update-set-task']['requestBody']['content']['application/json'],
      signal?: AbortSignal,
    ) => routes.updateSetTask({ path: { id, task }, body }, signal),
    deleteSetTask: (id: string, task: string, signal?: AbortSignal) =>
      routes.deleteSetTask({ path: { id, task } }, signal),
    assignSet: (
      body: operations['assign-set']['requestBody']['content']['application/json'],
      signal?: AbortSignal,
    ) => routes.assignSet({ body }, signal),
  };
};

/** One schedule with its lines. */
export type Set =
  operations['list-sets']['responses'][200]['content']['application/json']['sets'][number];
