import { formSchema, fromConfig, toConfig, toYaml } from './config-form-schema';
import type { Config } from '@/api/api';

describe('config-form-schema', () => {
  describe('fromConfig', () => {
    it('returns default structure when config is undefined', () => {
      expect(fromConfig(undefined)).toEqual({
        globalEnvVars: [{ key: '', value: '' }],
        services: [],
      });
    });

    it('returns empty env arrays when config has no variables or services', () => {
      const config: Config = { globalVariables: {}, services: {} };
      expect(fromConfig(config)).toEqual({
        globalEnvVars: [{ key: '', value: '' }],
        services: [],
      });
    });

    it('converts global variables to array with a trailing empty placeholder', () => {
      const config: Config = {
        globalVariables: { FOO: 'bar', BAZ: 'qux' },
        services: {},
      };
      const result = fromConfig(config);
      expect(result.globalEnvVars).toEqual([
        { key: 'FOO', value: 'bar' },
        { key: 'BAZ', value: 'qux' },
        { key: '', value: '' },
      ]);
    });

    it('converts service env vars with a trailing empty placeholder', () => {
      const config: Config = {
        globalVariables: {},
        services: { web: { PORT: '3000' } },
      };
      const result = fromConfig(config);
      expect(result.services).toEqual([
        {
          name: 'web',
          envVars: [
            { key: 'PORT', value: '3000' },
            { key: '', value: '' },
          ],
        },
      ]);
    });

    it('converts multiple services preserving entry order', () => {
      const config: Config = {
        globalVariables: { DEBUG: 'true' },
        services: {
          web: { PORT: '3000' },
          db: { USER: 'admin' },
        },
      };
      const result = fromConfig(config);
      expect(result.services).toEqual([
        {
          name: 'web',
          envVars: [
            { key: 'PORT', value: '3000' },
            { key: '', value: '' },
          ],
        },
        {
          name: 'db',
          envVars: [
            { key: 'USER', value: 'admin' },
            { key: '', value: '' },
          ],
        },
      ]);
    });

    it('handles service with no env vars', () => {
      const config: Config = {
        globalVariables: {},
        services: { empty: {} },
      };
      const result = fromConfig(config);
      expect(result.services).toEqual([
        {
          name: 'empty',
          envVars: [{ key: '', value: '' }],
        },
      ]);
    });
  });

  describe('toConfig', () => {
    it('strips entries with empty keys from global env vars', () => {
      const formValues = {
        globalEnvVars: [
          { key: 'FOO', value: 'bar' },
          { key: '', value: '' },
        ],
        services: [],
      };
      expect(toConfig(formValues)).toEqual({
        globalVariables: { FOO: 'bar' },
        services: {},
      });
    });

    it('strips entries with whitespace-only keys', () => {
      const formValues = {
        globalEnvVars: [
          { key: 'FOO', value: 'bar' },
          { key: '   ', value: 'ignored' },
          { key: '', value: 'also-ignored' },
        ],
        services: [],
      };
      expect(toConfig(formValues)).toEqual({
        globalVariables: { FOO: 'bar' },
        services: {},
      });
    });

    it('preserves non-empty global env vars and drops placeholders', () => {
      const formValues = {
        globalEnvVars: [
          { key: 'A', value: '1' },
          { key: 'B', value: '2' },
          { key: '', value: '' },
        ],
        services: [],
      };
      expect(toConfig(formValues).globalVariables).toEqual({ A: '1', B: '2' });
    });

    it('strips services with empty names', () => {
      const formValues = {
        globalEnvVars: [{ key: '', value: '' }],
        services: [
          { name: '', envVars: [{ key: '', value: '' }] },
          {
            name: 'web',
            envVars: [
              { key: 'PORT', value: '3000' },
              { key: '', value: '' },
            ],
          },
        ],
      };
      expect(toConfig(formValues)).toEqual({
        globalVariables: {},
        services: {
          web: { PORT: '3000' },
        },
      });
    });

    it('strips whitespace-only entries inside service env vars', () => {
      const formValues = {
        globalEnvVars: [{ key: '', value: '' }],
        services: [
          {
            name: 'web',
            envVars: [
              { key: 'PORT', value: '3000' },
              { key: '', value: 'empty' },
              { key: '   ', value: 'whitespace' },
            ],
          },
        ],
      };
      expect(toConfig(formValues).services).toEqual({
        web: { PORT: '3000' },
      });
    });
  });

  describe('toConfig round-trip', () => {
    it('preserves config data through fromConfig → toConfig', () => {
      const config: Config = {
        globalVariables: { FOO: 'bar', BAZ: 'qux' },
        services: {
          web: { PORT: '3000', HOST: '0.0.0.0' },
        },
      };
      expect(toConfig(fromConfig(config))).toEqual(config);
    });

    it('preserves config with multiple services through round-trip', () => {
      const config: Config = {
        globalVariables: { DEBUG: 'true', LOG_LEVEL: 'info' },
        services: {
          web: { PORT: '3000', HOST: '0.0.0.0' },
          db: { USER: 'admin', PASS: 'secret' },
        },
      };
      expect(toConfig(fromConfig(config))).toEqual(config);
    });

    it('preserves empty config through round-trip', () => {
      const config: Config = { globalVariables: {}, services: {} };
      expect(toConfig(fromConfig(config))).toEqual(config);
    });
  });

  describe('toYaml', () => {
    it('produces a YAML string from form values', () => {
      const result = toYaml({
        globalEnvVars: [
          { key: 'FOO', value: 'bar' },
          { key: '', value: '' },
        ],
        services: [],
      });
      expect(result).toContain('globalVariables:');
      expect(result).toContain('FOO: bar');
    });

    it('includes services in the YAML output', () => {
      const result = toYaml({
        globalEnvVars: [{ key: '', value: '' }],
        services: [
          {
            name: 'web',
            envVars: [
              { key: 'PORT', value: '3000' },
              { key: '', value: '' },
            ],
          },
        ],
      });
      expect(result).toContain('services:');
      expect(result).toContain('web:');
      expect(result).toContain('PORT:');
    });

    it('returns empty string when formData is falsy', () => {
      expect(toYaml(undefined as unknown as Parameters<typeof toYaml>[0])).toBe('');
    });
  });

  describe('formSchema', () => {
    it('validates a valid form value with empty services', () => {
      const valid = { globalEnvVars: [], services: [] };
      expect(formSchema.parse(valid)).toEqual(valid);
    });

    it('validates a form value with a named service', () => {
      const valid = {
        globalEnvVars: [{ key: '', value: '' }],
        services: [{ name: 'web', envVars: [{ key: '', value: '' }] }],
      };
      expect(formSchema.parse(valid)).toEqual(valid);
    });

    it('rejects a service with an empty name', () => {
      const invalid = {
        globalEnvVars: [],
        services: [{ name: '', envVars: [] }],
      };
      expect(() => formSchema.parse(invalid)).toThrow();
    });

    it('accepts a service with a whitespace-only name (filtered later by toConfig)', () => {
      const valid = {
        globalEnvVars: [],
        services: [{ name: '   ', envVars: [{ key: '', value: '' }] }],
      };
      expect(formSchema.parse(valid)).toEqual(valid);
    });
  });
});
