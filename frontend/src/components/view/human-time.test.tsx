import { render, screen } from '@testing-library/react';
import { HumanTime } from './human-time';

const mockUseRelativeTime = vi.hoisted(() => vi.fn());

vi.mock('@/hooks', () => ({
  useRelativeTime: mockUseRelativeTime,
}));

describe('HumanTime', () => {
  beforeEach(() => {
    mockUseRelativeTime.mockReturnValue(null);
  });

  it('renders defaultValue when time is undefined', () => {
    render(<HumanTime defaultValue="Never" />);
    expect(screen.getByText('Never')).toBeTruthy();
  });

  it('renders defaultValue when relativeTime is falsy', () => {
    render(<HumanTime time="2024-01-01T10:00:00Z" defaultValue="Never" />);
    expect(screen.getByText('Never')).toBeTruthy();
  });

  it('renders relativeTime when both time and relativeTime are provided', () => {
    mockUseRelativeTime.mockReturnValue('2 minutes ago');
    render(<HumanTime time="2024-01-01T10:00:00Z" />);
    expect(screen.getByText('2 minutes ago')).toBeTruthy();
    expect(document.querySelector('[data-slot="tooltip-trigger"]')).not.toBeNull();
  });
});
