import { render, screen, fireEvent } from '@testing-library/react';
import { useForm } from 'react-hook-form';
import type { FormValues } from './config-form-schema';
import { ConfigForm } from './config-form';

const { envVarArrayFormSpy } = vi.hoisted(() => ({
  envVarArrayFormSpy: vi.fn((_props: unknown) => <div data-slot="env-vars-array-form" />),
}));

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
  EnvVarArrayForm: (props: unknown) => envVarArrayFormSpy(props),
}));

function TestWrapper({ disabled }: { disabled?: boolean }) {
  const form = useForm<FormValues>({
    defaultValues: {
      services: [{ name: 'web', envVars: [{ key: '', value: '' }] }],
      globalEnvVars: [],
    },
  });
  return <ConfigForm form={form} className="flex-1 p-4" disabled={disabled} />;
}

describe('ConfigForm', () => {
  beforeEach(() => {
    envVarArrayFormSpy.mockClear();
  });

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
    render(<TestWrapper disabled />);
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
      screen
        .getByText('translated:CONFIGURATION.FORM.ADD_SERVICE')
        .closest('button') as HTMLElement,
    );
    expect(container.querySelectorAll('[data-slot="service-card"]')).toHaveLength(2);
  });

  it('removes a service when Remove button is clicked', () => {
    const { container } = render(<TestWrapper />);
    fireEvent.click(container.querySelector('[data-slot="remove-service"]') as HTMLElement);
    expect(container.querySelectorAll('[data-slot="service-card"]')).toHaveLength(0);
  });

  it('does not re-render children when re-rendered with same props (React.memo)', () => {
    const { rerender } = render(<TestWrapper />);

    // EnvVarArrayForm mock should have rendered exactly once
    expect(envVarArrayFormSpy).toHaveBeenCalledTimes(1);

    // Re-render with identical props — React.memo should prevent ConfigForm
    // from re-rendering because form (stable ref), className, and disabled
    // are all identical.
    rerender(<TestWrapper />);

    // Without React.memo: ConfigForm re-renders on rerender → EnvVarArrayForm
    // mock is called again → spy count = 2 (test FAILS)
    // With React.memo: ConfigForm skips re-render → EnvVarArrayForm mock
    // not called again → spy count = 1 (test PASSES)
    expect(envVarArrayFormSpy).toHaveBeenCalledTimes(1);
  });
});
