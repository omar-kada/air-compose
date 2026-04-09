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
});
export type FormValues = z.infer<typeof formSchema>;

export function fromSettings(settings?: Settings): FormValues {
  if (!settings) {
    return {
      repo: '',
      notificationTypes: [],
    };
  }
  return settings;
}

export function toSettings(formValues: FormValues): Settings {
  return formValues;
}
