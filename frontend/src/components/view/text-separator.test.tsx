import { render, screen } from '@testing-library/react';
import { TextSeparator } from './text-separator';

describe('TextSeparator', () => {
  it('renders text between separators', () => {
    const { container } = render(<TextSeparator text="OR" />);
    expect(screen.getByText('OR')).toBeTruthy();
    expect(container.querySelectorAll('[data-slot="separator"]').length).toBe(2);
  });
});
