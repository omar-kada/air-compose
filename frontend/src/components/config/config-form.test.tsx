import { render, screen, fireEvent } from '@testing-library/react';
import { useForm } from 'react-hook-form';
import type { FormValues } from './config-form-schema';
import { ConfigForm } from './config-form';

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('./service-card', () => ({
  ServiceCard: ({ service, onRemove }: { service: { name: string }; onRemove: () => void }) => (
    <div data-slot="service-card" data-name={service.name}>
      <button data-slot="remove-service" onClick={onRemove}>
        Remove
      </button>
    </div>
  ),
}));

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
  return <ConfigForm form={form} disabled={disabled} />;
}

describe('ConfigForm', () => {
  it('renders service cards for each service', () => {
    const { container } = render(<TestWrapper />);
    const cards = container.querySelectorAll('[data-slot="service-card"]');
    expect(cards).toHaveLength(1);
    expect(container.querySelector('[data-slot="service-card"][data-name="web"]')).not.toBeNull();
  });

  it('renders global env vars form', () => {
    const { container } = render(<TestWrapper />);
    expect(container.querySelector('[data-slot="env-vars-array-form"]')).not.toBeNull();
  });

  it('renders add service button when not disabled', () => {
    render(<TestWrapper />);
    expect(screen.getByText('translated:CONFIGURATION.FORM.ADD_SERVICE')).toBeTruthy();
  });

  it('hides add service button when disabled', () => {
    render(<TestWrapper disabled={true} />);
    expect(screen.queryByText('translated:CONFIGURATION.FORM.ADD_SERVICE')).toBeNull();
  });

  it('renders multiple service cards', () => {
    function TestWrapperMulti() {
      const form = useForm<FormValues>({
        defaultValues: {
          services: [
            { name: 'web', envVars: [{ key: '', value: '' }] },
            { name: 'db', envVars: [{ key: '', value: '' }] },
          ],
          globalEnvVars: [],
        },
      });
      return <ConfigForm form={form} />;
    }
    const { container } = render(<TestWrapperMulti />);
    expect(container.querySelectorAll('[data-slot="service-card"]')).toHaveLength(2);
  });

  it('adds a new service when Add Service button is clicked', () => {
    const { container } = render(<TestWrapper />);
    fireEvent.click(
      screen.getByText('translated:CONFIGURATION.FORM.ADD_SERVICE').closest('button') as HTMLElement,
    );
    expect(container.querySelectorAll('[data-slot="service-card"]')).toHaveLength(2);
  });

  it('removes a service when Remove button is clicked', () => {
    const { container } = render(<TestWrapper />);
    fireEvent.click(container.querySelector('[data-slot="remove-service"]') as HTMLElement);
    expect(container.querySelectorAll('[data-slot="service-card"]')).toHaveLength(0);
  });
});
