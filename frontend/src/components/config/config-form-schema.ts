import type { Config } from '@/api/api';
import yaml from 'js-yaml';
import z from 'zod';

export const formSchema = z.object({
  globalEnvVars: z.array(
    z.object({
      key: z.string(),
      value: z.string(),
    }),
  ),
  services: z.array(
    z.object({
      name: z.string().min(1, { message: 'CONFIGURATION.FORM.SERVICE_NAME_REQUIRED' }),
      envVars: z.array(
        z.object({
          key: z.string(),
          value: z.string(),
        }),
      ),
    }),
  ),
});
export type FormValues = z.infer<typeof formSchema>;

export function fromConfig(config?: Config): FormValues {
  return {
    globalEnvVars: objectToArray(config?.globalVariables).concat([{ key: '', value: '' }]),
    services: config?.services
      ? Object.entries(config.services).map(([serviceKey, serviceValue]) => ({
          name: serviceKey,
          envVars: objectToArray(serviceValue).concat([{ key: '', value: '' }]),
        }))
      : [],
  };
}

function objectToArray<T>(obj?: { [key: string]: T }): Array<{ key: string; value: T }> {
  if (!obj) {
    return [];
  }
  return Object.entries(obj).map(([key, value]) => ({ key, value }));
}

export function toConfig(data: FormValues): Config {
  return {
    globalVariables: envArrayToObject(data.globalEnvVars),
    services: envArrayToObject(
      data.services?.map((service) => {
        return { key: service.name, value: envArrayToObject(service.envVars) };
      }),
    ),
  };
}

function envArrayToObject<T>(vars: { key: string; value: T }[]): { [key: string]: T } {
  if (!vars) return {};
  return vars
    .filter((env) => env.key && env.key.trim() !== '')
    .reduce((acc, env) => ({ ...acc, [env.key]: env.value }), {});
}

export function toYaml(formData: FormValues): string {
  if (!formData) return '';
  const config = toConfig(formData);
  return yaml.dump(config);
}
