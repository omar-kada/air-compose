import { EventType, type Settings } from '@/api/api';
import z from 'zod';

export const formSchema = z.object({
  repo: z.string().min(1, { message: 'SETTINGS.FORM.repo_REQUIRED' }),
  branch: z.string().optional(),
  token: z.string().optional(),
  username: z.string().optional(),
  cron: z.string().optional(),
  notificationURL: z.string().optional(),
  notificationTypes: z.array(z.enum(EventType)),
  retriesOnUnhealthy: z.number().nonoptional(),
  retryDelay: z.number().nonoptional(),
});
export type FormValues = z.infer<typeof formSchema>;

export function fromSettings(settings?: Settings): FormValues {
  if (!settings) {
    return {
      repo: '',
      notificationTypes: [],
      retriesOnUnhealthy: 0,
      retryDelay: 2 * 60 * 1000,
    };
  }
  return {
    ...settings,
    retryDelay: settings.retryDelay / 60000,
  };
}

export function toSettings(formValues: FormValues): Settings {
  return {
    ...formValues,
    retriesOnUnhealthy: Number(formValues.retriesOnUnhealthy),
    retryDelay: Number(formValues.retryDelay) * 60000,
  };
}
