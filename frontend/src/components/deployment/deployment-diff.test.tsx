import { render, screen } from '@testing-library/react';
import type { FileDiff } from '@/api/api';
import { DeploymentDiff, DeploymentDiffSkeleton } from './deployment-diff';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => `translated:${key}`,
    i18n: { language: 'en' },
  }),
}));

vi.mock('./file-diff', () => ({
  FileDiffView: ({
    fileDiff,
  }: {
    fileDiff: { oldFile: string; newFile: string; diff: string };
  }) => (
    <div data-testid="file-diff-view" data-old={fileDiff.oldFile} data-new={fileDiff.newFile} />
  ),
}));

const mockFileDiff = (over: Partial<FileDiff> = {}): FileDiff => ({
  oldFile: 'src/config.ts',
  newFile: 'src/config.ts',
  diff: '@@ -1,3 +1,3 @@',
  ...over,
});

describe('DeploymentDiff', () => {
  it('renders the updated files label', () => {
    render(<DeploymentDiff fileDiffs={[mockFileDiff()]} />);
    expect(screen.queryByText('translated:DIFF.UPDATED_FILES')).not.toBeNull();
  });

  it('renders the badge with the correct file count', () => {
    render(
      <DeploymentDiff
        fileDiffs={[mockFileDiff({ oldFile: 'a.ts' }), mockFileDiff({ oldFile: 'b.ts' })]}
      />,
    );
    expect(screen.queryByText('2')).not.toBeNull();
  });

  it('renders a FileDiffView for each file diff', () => {
    render(
      <DeploymentDiff
        fileDiffs={[mockFileDiff({ oldFile: 'a.ts' }), mockFileDiff({ oldFile: 'b.ts' })]}
      />,
    );
    expect(screen.getAllByTestId('file-diff-view')).toHaveLength(2);
  });

  it('renders children as a card description (data-slot)', () => {
    render(
      <DeploymentDiff fileDiffs={[mockFileDiff()]}>Some context about the diff</DeploymentDiff>,
    );
    const description = document.querySelector('[data-slot="card-description"]');
    expect(description).not.toBeNull();
    expect(description?.textContent).toBe('Some context about the diff');
  });

  it('does not render a card description when children are not provided', () => {
    render(<DeploymentDiff fileDiffs={[mockFileDiff()]} />);
    expect(document.querySelector('[data-slot="card-description"]')).toBeNull();
  });

  it('renders badge with 0 when there are no file diffs', () => {
    render(<DeploymentDiff fileDiffs={[]} />);
    expect(screen.queryByText('0')).not.toBeNull();
  });
});

describe('DeploymentDiffSkeleton', () => {
  it('renders three skeleton placeholders', () => {
    const { container } = render(<DeploymentDiffSkeleton />);
    expect(container.querySelectorAll('[data-slot="skeleton"]')).toHaveLength(3);
  });
});
