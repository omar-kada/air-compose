import { EventType, type GitCredentails, type Settings } from '@/api/api';
import z from 'zod';

export const repoSchema = z
  .object({
    repo: z.string().min(1, { message: 'ONBOARDING.FORM.repo_REQUIRED' }),
    branch: z.string().optional(),
    privateRepo: z.boolean().default(false).optional(),
    token: z.string().optional(),
    username: z.string().optional(),
  })
  .superRefine((data, ctx) => {
    if (data.privateRepo && !data.token) {
      ctx.addIssue({
        code: 'custom',
        message: 'ONBOARDING.FORM.token_REQUIRED',
        path: ['token'],
      });
    }
    if (data.privateRepo && !data.username) {
      ctx.addIssue({
        code: 'custom',
        message: 'ONBOARDING.FORM.username_REQUIRED',
        path: ['username'],
      });
    }
  });

export type RepoFormValues = z.infer<typeof repoSchema>;

export const cronSchema = z.object({
  cron: z.string().optional(),
});

export type CronFormValues = z.infer<typeof cronSchema>;

export const notificationSchema = z
  .object({
    enableNotifications: z.boolean().default(false).optional(),
    notificationURL: z.string().optional(),
    notificationTypes: z.array(z.enum(EventType)),
  })
  .superRefine((data, ctx) => {
    if (data.enableNotifications && !data.notificationURL) {
      ctx.addIssue({
        code: 'custom',
        message: 'ONBOARDING.FORM.NOTIFICATION_URL_REQUIRED',
        path: ['notificationURL'],
      });
    }
  });

export type NotificationFormValues = z.infer<typeof notificationSchema>;

export type OnboardingFormValues = RepoFormValues & CronFormValues & NotificationFormValues;

export function toSettings(formValues: OnboardingFormValues): Settings {
  return {
    repo: formValues.repo?.trim(),
    branch: formValues.branch?.trim(),
    username: formValues.username?.trim(),
    token: formValues.token?.trim(),
    cron: formValues.cron?.trim(),
    notificationURL: formValues.notificationURL?.trim(),
    notificationTypes: formValues.notificationTypes,
  };
}

export function toGitCredentaials(formValues: RepoFormValues): GitCredentails {
  return {
    repo: formValues.repo?.trim(),
    branch: formValues.branch?.trim(),
    username: formValues.username?.trim(),
    token: formValues.token?.trim(),
  };
}

export function toRepoFormValues(settings: Settings): RepoFormValues {
  return {
    repo: settings.repo,
    branch: settings.branch,
    privateRepo: settings.token != null,
    username: settings.username,
    token: settings.token,
  };
}

export function toNotificationFormValues(settings: Settings): NotificationFormValues {
  return {
    enableNotifications: settings.notificationURL != null,
    notificationURL: settings.notificationURL,
    notificationTypes: settings.notificationTypes,
  };
}

export function toCronFormValues(settings: Settings): CronFormValues {
  return {
    cron: settings.cron,
  };
}
