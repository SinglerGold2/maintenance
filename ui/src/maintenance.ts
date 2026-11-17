// The service layer: this module's own routes, reached through the client the
// host handed over. Nothing here names a URL or a token.
import type { ModuleHost } from '@prusa/connect-ui/slots';
import { bind, type operations } from 'maintenance-client';

/** What GET /overview answers, as openapi.json declares it. */
export type Overview =
  operations['read-overview']['responses'][200]['content']['application/json'];

/** One task with its counters. */
export type TaskView = Overview['devices'][number]['tasks'][number];

/** The route set, bound onto the host's client. */
export const maintenanceOn = (host: ModuleHost) => {
  const routes = bind(host.client);

  return {
    overview: (signal?: AbortSignal): Promise<Overview> =>
      routes.readOverview({}, signal),
    saveTask: (
      body: operations['save-task']['requestBody']['content']['application/json'],
      signal?: AbortSignal,
    ) => routes.saveTask({ body }, signal),
    markDone: (id: string, signal?: AbortSignal) =>
      routes.markDone({ body: { id } }, signal),
    deleteTask: (id: string, signal?: AbortSignal) =>
      routes.deleteTask({ query: { id } }, signal),
    savePrinterHours: (deviceID: string, hours: number, signal?: AbortSignal) =>
      routes.savePrinterHours({ body: { device_id: deviceID, hours } }, signal),
    template: (signal?: AbortSignal) => routes.readTemplate({}, signal),
    saveTemplate: (
      body: operations['save-template']['requestBody']['content']['application/json'],
      signal?: AbortSignal,
    ) => routes.saveTemplate({ body }, signal),
  };
};
