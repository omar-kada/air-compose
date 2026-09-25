import { render } from '@testing-library/react';
import { useForm, useFieldArray } from 'react-hook-form';
import type { FormValues } from './config-form-schema';
import { ServiceCard } from './service-card';

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('./env-vars-array-form', () => ({
  EnvVarArrayForm: () => <div data-slot="env-vars-array-form" />,
}));

function TestWrapper({ disabled }: { disabled?: boolean }) {
  const form = useForm<FormValues>({
    defaultValues: {
      services: [{ name: 'web', envVars: [{ key: '', value: '' }] }],
      globalEnvVars: [],
    },
  });
  const { fields } = useFieldArray({ control: form.control, name: 'services' });
  return (
    <ServiceCard
      form={form}
      service={fields[0]}
      name="services.0"
      disabled={disabled}
      onRemove={vi.fn()}
    />
  );
}

describe('ServiceCard', () => {
  it('renders service logo for the service', () => {
    const { container } = render(<TestWrapper />);
    const avatar = container.querySelector('[data-slot="avatar"]');
    expect(avatar).not.toBeNull();
    expect(avatar?.querySelector('img')).not.toBeNull();
  });

  it('renders env vars form', () => {
    const { container } = render(<TestWrapper />);
    expect(container.querySelector('[data-slot="env-vars-array-form"]')).not.toBeNull();
  });

  it('renders remove button when not disabled', () => {
    const { container } = render(<TestWrapper />);
    expect(container.querySelector('[data-slot="button"]')).not.toBeNull();
  });

  it('hides remove button when disabled', () => {
    const { container } = render(<TestWrapper disabled />);
    expect(container.querySelector('[data-slot="button"]')).toBeNull();
  });

  it('renders service name input with placeholder', () => {
    const { container } = render(<TestWrapper />);
    const input = container.querySelector('[data-slot="input"]');
    expect(input).not.toBeNull();
    expect(input?.getAttribute('placeholder')).toBe('translated:CONFIGURATION.FORM.SERVICE_NAME');
  });
});
