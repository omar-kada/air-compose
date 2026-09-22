import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { ChangePasswordDialog } from './change-password-dialog';

const mockChangePass = vi.hoisted(() => vi.fn().mockResolvedValue({ data: { success: true } }));

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('sonner', async () => {
  const { createSonnerMock } = await import('@/tests/mock-factories');
  return createSonnerMock();
});

vi.mock('@/hooks/user/use-change-pass', () => ({
  useChangePass: () => ({
    changePass: mockChangePass,
    isPending: false,
    error: null,
  }),
}));

vi.mock('../view', async () => {
  const { createErrorAlertMock } = await import('@/tests/mock-factories');
  return createErrorAlertMock();
});

describe('ChangePasswordDialog', () => {
  it('renders password fields and action buttons when dialog is opened', async () => {
    render(
      <ChangePasswordDialog>
        <button>Open Dialog</button>
      </ChangePasswordDialog>,
    );
    fireEvent.click(screen.getByText('Open Dialog'));
    await waitFor(() => {
      expect(document.querySelectorAll('input[type="password"]')).toHaveLength(2);
    });
    expect(screen.getByRole('button', { name: 'translated:ACTION.CONFIRM' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'translated:ACTION.CANCEL' })).toBeTruthy();
  });

  it('toggles password visibility when eye icon button is clicked', async () => {
    render(
      <ChangePasswordDialog>
        <button>Open Dialog</button>
      </ChangePasswordDialog>,
    );
    fireEvent.click(screen.getByText('Open Dialog'));
    await waitFor(() => {
      expect(document.querySelectorAll('input[type="password"]')).toHaveLength(2);
    });
    const toggleButtons = screen.getAllByRole('button', { name: 'Show password' });
    expect(toggleButtons).toHaveLength(2);
    fireEvent.click(toggleButtons[0]);
    await waitFor(() => {
      expect(document.querySelectorAll('input[type="password"]')).toHaveLength(1);
      expect(document.querySelectorAll('input[type="text"]')).toHaveLength(1);
    });
  });

  it('calls changePass on form submit with valid data', async () => {
    render(
      <ChangePasswordDialog>
        <button>Open Dialog</button>
      </ChangePasswordDialog>,
    );
    fireEvent.click(screen.getByText('Open Dialog'));
    await waitFor(() => {
      expect(document.querySelectorAll('input[type="password"]')).toHaveLength(2);
    });
    const inputs = document.querySelectorAll('input[type="password"]');
    fireEvent.change(inputs[0] as HTMLInputElement, { target: { value: 'oldpass123' } });
    fireEvent.change(inputs[1] as HTMLInputElement, { target: { value: 'newpass123' } });
    fireEvent.click(screen.getByRole('button', { name: 'translated:ACTION.CONFIRM' }));
    await waitFor(() => {
      expect(mockChangePass).toHaveBeenCalledWith({
        data: { oldPass: 'oldpass123', newPass: 'newpass123' },
      });
    });
  });

  it('does not call changePass when fields are too short', async () => {
    render(
      <ChangePasswordDialog>
        <button>Open Dialog</button>
      </ChangePasswordDialog>,
    );
    fireEvent.click(screen.getByText('Open Dialog'));
    await waitFor(() => {
      expect(document.querySelectorAll('input[type="password"]')).toHaveLength(2);
    });
    fireEvent.click(screen.getByRole('button', { name: 'translated:ACTION.CONFIRM' }));
    await waitFor(() => {
      expect(mockChangePass).not.toHaveBeenCalled();
    });
  });
});
