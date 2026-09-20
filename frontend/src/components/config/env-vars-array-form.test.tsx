import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { useForm } from 'react-hook-form';
import type { FormValues } from './config-form-schema';
import { EnvVarArrayForm } from './env-vars-array-form';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

function FormWrapper({
  name,
  defaultValues,
}: {
  name: 'globalEnvVars' | `services.${number}.envVars`;
  defaultValues: FormValues;
}) {
  const form = useForm<FormValues>({ defaultValues });
  return <EnvVarArrayForm control={form.control} name={name} />;
}

describe('EnvVarArrayForm', () => {
  describe('global env vars', () => {
    const name = 'globalEnvVars';
    const defaultValues: FormValues = {
      globalEnvVars: [{ key: '', value: '' }],
      services: [],
    };

    it('appends a new empty entry when the last key field is filled', async () => {
      render(<FormWrapper name={name} defaultValues={defaultValues} />);

      // Initially there is 1 key input and 1 value textarea
      expect(screen.getAllByPlaceholderText('CONFIGURATION.FORM.KEY')).toHaveLength(1);
      expect(screen.getAllByPlaceholderText('CONFIGURATION.FORM.VALUE')).toHaveLength(1);

      // Type in the first key input
      fireEvent.change(screen.getByPlaceholderText('CONFIGURATION.FORM.KEY'), {
        target: { value: 'FOO' },
      });

      // After typing, a new empty entry should be appended
      await waitFor(() => {
        expect(screen.getAllByPlaceholderText('CONFIGURATION.FORM.KEY')).toHaveLength(2);
        expect(screen.getAllByPlaceholderText('CONFIGURATION.FORM.VALUE')).toHaveLength(2);
      });

      // The new entry should have an empty key and value
      const keyInputs = screen.getAllByPlaceholderText(
        'CONFIGURATION.FORM.KEY',
      ) as HTMLInputElement[];
      expect(keyInputs[0].value).toBe('FOO');
      expect(keyInputs[1].value).toBe('');
    });
  });

  describe('per-service env vars', () => {
    const name = 'services.0.envVars' as `services.${number}.envVars`;
    const defaultValues: FormValues = {
      globalEnvVars: [],
      services: [{ name: 'web', envVars: [{ key: '', value: '' }] }],
    };

    it('appends a new empty entry when the last key field is filled', async () => {
      render(<FormWrapper name={name} defaultValues={defaultValues} />);

      // Initially there is 1 key input and 1 value textarea
      expect(screen.getAllByPlaceholderText('CONFIGURATION.FORM.KEY')).toHaveLength(1);
      expect(screen.getAllByPlaceholderText('CONFIGURATION.FORM.VALUE')).toHaveLength(1);

      // Type in the first key input
      fireEvent.change(screen.getByPlaceholderText('CONFIGURATION.FORM.KEY'), {
        target: { value: 'FOO' },
      });

      // After typing, a new empty entry should be appended
      await waitFor(() => {
        expect(screen.getAllByPlaceholderText('CONFIGURATION.FORM.KEY')).toHaveLength(2);
        expect(screen.getAllByPlaceholderText('CONFIGURATION.FORM.VALUE')).toHaveLength(2);
      });

      // The new entry should have an empty key and value
      const keyInputs = screen.getAllByPlaceholderText(
        'CONFIGURATION.FORM.KEY',
      ) as HTMLInputElement[];
      expect(keyInputs[0].value).toBe('FOO');
      expect(keyInputs[1].value).toBe('');
    });
  });
});
