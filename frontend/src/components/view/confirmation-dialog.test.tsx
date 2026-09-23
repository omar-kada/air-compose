import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { ConfirmationDialog } from './confirmation-dialog';

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

describe('ConfirmationDialog', () => {
  it('opens dialog when trigger is clicked', async () => {
    const onConfirm = vi.fn().mockResolvedValue(undefined);
    render(
      <ConfirmationDialog title="Confirm?" description="Are you sure?" onConfirm={onConfirm}>
        <button>Open</button>
      </ConfirmationDialog>,
    );
    fireEvent.click(screen.getByText('Open'));
    await waitFor(() => {
      expect(screen.getByText('Confirm?')).toBeTruthy();
    });
    expect(screen.getByText('Are you sure?')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'translated:ACTION.CONFIRM' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'translated:ACTION.CANCEL' })).toBeTruthy();
  });

  it('calls onConfirm when confirm button is clicked', async () => {
    const onConfirm = vi.fn().mockResolvedValue(undefined);
    render(
      <ConfirmationDialog title="Confirm?" description="Are you sure?" onConfirm={onConfirm}>
        <button>Open</button>
      </ConfirmationDialog>,
    );
    fireEvent.click(screen.getByText('Open'));
    await waitFor(() => {
      expect(screen.getByText('Confirm?')).toBeTruthy();
    });
    fireEvent.click(screen.getByRole('button', { name: 'translated:ACTION.CONFIRM' }));
    await waitFor(() => {
      expect(onConfirm).toHaveBeenCalled();
    });
  });

  it('closes immediately when onConfirm returns undefined', async () => {
    const onConfirm = vi.fn().mockReturnValue(undefined);
    render(
      <ConfirmationDialog title="Confirm?" description="Are you sure?" onConfirm={onConfirm}>
        <button>Open</button>
      </ConfirmationDialog>,
    );
    fireEvent.click(screen.getByText('Open'));
    await waitFor(() => {
      expect(screen.getByText('Confirm?')).toBeTruthy();
    });
    fireEvent.click(screen.getByRole('button', { name: 'translated:ACTION.CONFIRM' }));
    await waitFor(() => {
      expect(onConfirm).toHaveBeenCalled();
    });
    await waitFor(() => {
      expect(screen.queryByText('Confirm?')).toBeNull();
    });
  });

  it('shows spinner during loading and closes after promise resolves', async () => {
    let resolveFn: () => void = () => {};
    const onConfirm = vi.fn().mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          resolveFn = resolve;
        }),
    );
    render(
      <ConfirmationDialog title="Confirm?" description="Are you sure?" onConfirm={onConfirm}>
        <button>Open</button>
      </ConfirmationDialog>,
    );
    fireEvent.click(screen.getByText('Open'));
    await waitFor(() => {
      expect(screen.getByText('Confirm?')).toBeTruthy();
    });
    fireEvent.click(screen.getByRole('button', { name: /CONFIRM/ }));
    await waitFor(() => {
      const confirmButton = screen.getByRole('button', { name: /CONFIRM/ });
      expect(confirmButton.querySelector('[aria-label="Loading"]')).not.toBeNull();
    });
    resolveFn();
    await waitFor(() => {
      expect(screen.queryByText('Confirm?')).toBeNull();
    });
  });
});
