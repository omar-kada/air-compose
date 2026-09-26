import { render, fireEvent } from '@testing-library/react';
import { FileDiffView } from './file-diff';

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('@/hooks', () => ({
  useIsMobile: () => false,
}));

vi.mock('@/hooks/theme-provider', () => ({
  useTheme: () => ({ theme: 'light' }),
}));

vi.mock('@git-diff-view/react', () => ({
  DiffView: () => <div data-slot="diff-view" />,
  DiffModeEnum: { Unified: 'unified', Split: 'split' },
}));

vi.mock('lucide-react', () => ({
  ChevronDown: () => <svg data-slot="chevron-down" />,
  ChevronUp: () => <svg data-slot="chevron-up" />,
  FileDiff: () => <svg data-slot="file-diff-icon" />,
}));

describe('FileDiffView', () => {
  const mockFileDiff = { oldFile: 'old.txt', newFile: 'new.txt', diff: '' };

  it('renders file name when collapsed', () => {
    const { container } = render(<FileDiffView fileDiff={mockFileDiff} />);
    expect(container.textContent).toContain('old.txt');
    expect(container.textContent).toContain('> new.txt');
    expect(container.querySelector('[data-slot="chevron-down"]')).not.toBeNull();
    expect(container.querySelector('[data-slot="diff-view"]')).toBeNull();
  });

  it('renders same file name without arrow when oldFile equals newFile', () => {
    const { container } = render(
      <FileDiffView fileDiff={{ oldFile: 'same.txt', newFile: 'same.txt', diff: '' }} />,
    );
    expect(container.textContent).toContain('same.txt');
    expect(container.textContent).not.toContain('> ');
  });

  it('renders DiffView when autoOpen is true', () => {
    const { container } = render(<FileDiffView fileDiff={mockFileDiff} autoOpen />);
    expect(container.querySelector('[data-slot="diff-view"]')).not.toBeNull();
    expect(container.querySelector('[data-slot="chevron-up"]')).not.toBeNull();
    expect(container.querySelector('[data-slot="chevron-down"]')).toBeNull();
  });

  it('applies className to the collapsible', () => {
    const { container } = render(<FileDiffView fileDiff={mockFileDiff} className="custom-class" />);
    const collapsible = container
      .querySelector('[data-slot="file-diff-icon"]')
      ?.closest('[class*="custom-class"]');
    expect(collapsible).not.toBeNull();
  });

  it('toggles expand when trigger is clicked', () => {
    const { container } = render(<FileDiffView fileDiff={mockFileDiff} />);
    expect(container.querySelector('[data-slot="diff-view"]')).toBeNull();
    const trigger = container.querySelector('[data-slot="file-diff-icon"]')?.closest('button');
    expect(trigger).not.toBeNull();
    fireEvent.click(trigger as HTMLElement);
    expect(container.querySelector('[data-slot="diff-view"]')).not.toBeNull();
  });
});
