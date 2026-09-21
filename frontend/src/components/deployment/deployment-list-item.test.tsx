import { render, screen, fireEvent } from '@testing-library/react';
import type { Deployment } from '@/api/api';
import { DeploymentListItem, DeploymentItemSkeleton } from './deployment-list-item';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => `translated:${key}`,
    i18n: { language: 'en' },
  }),
}));

const mockDeployment: Deployment = {
  id: '123',
  title: 'deploy-web-v1',
  author: 'alice',
  time: '2025-01-01T00:00:00Z',
  endTime: '2025-01-01T00:05:00Z',
  diff: '',
  status: 'success',
};

describe('DeploymentListItem', () => {
  it('renders the deployment title', () => {
    render(
      <DeploymentListItem deployment={mockDeployment} isSelected={false} onSelect={() => {}} />,
    );
    expect(screen.getByText('deploy-web-v1')).not.toBeNull();
  });

  it('renders the deployment ID', () => {
    render(
      <DeploymentListItem deployment={mockDeployment} isSelected={false} onSelect={() => {}} />,
    );
    expect(screen.getByText(/#123/)).not.toBeNull();
  });

  it('renders a status badge icon and the chevron icon', () => {
    const { container } = render(
      <DeploymentListItem deployment={mockDeployment} isSelected={false} onSelect={() => {}} />,
    );
    // One SVG from DeploymentStatusBadge, one SVG from ChevronRight
    expect(container.querySelectorAll('svg')).toHaveLength(2);
  });

  it('calls onSelect with the deployment when clicked', () => {
    const onSelect = vi.fn();
    const { container } = render(
      <DeploymentListItem deployment={mockDeployment} isSelected={false} onSelect={onSelect} />,
    );
    fireEvent.click(container.firstChild as HTMLElement);
    expect(onSelect).toHaveBeenCalledWith(mockDeployment);
  });

  it('applies the selected style when isSelected is true', () => {
    const { container } = render(
      <DeploymentListItem deployment={mockDeployment} isSelected onSelect={() => {}} />,
    );
    expect((container.firstChild as HTMLElement).className).toContain('bg-sidebar');
  });

  it('does not apply the selected style when isSelected is false', () => {
    const { container } = render(
      <DeploymentListItem deployment={mockDeployment} isSelected={false} onSelect={() => {}} />,
    );
    expect((container.firstChild as HTMLElement).className).not.toContain('bg-sidebar');
  });
});

describe('DeploymentItemSkeleton', () => {
  it('renders without errors', () => {
    const { container } = render(<DeploymentItemSkeleton />);
    expect(container).not.toBeNull();
  });

  it('renders two skeleton placeholders', () => {
    const { container } = render(<DeploymentItemSkeleton />);
    expect(container.querySelectorAll('[data-slot="skeleton"]')).toHaveLength(2);
  });
});
