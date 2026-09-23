const mockFn = vi.hoisted(() =>
  vi.fn((opts: any) => ({
    mutationFn: vi.fn().mockResolvedValue({ data: { success: true } }),
    ...(opts?.mutation ?? {}),
  })),
);

vi.mock('@/api/api', async (importOriginal) => {
  const original = await importOriginal<typeof import('@/api/api')>();
  return { ...original, getUserAPIChangePasswordMutationOptions: mockFn };
});

import { useChangePass } from './use-change-pass';
import { renderHookWithQuery } from '@/tests/test-utils';

describe('useChangePass', () => {
  it('returns changePass as mutateAsync', () => {
    const { result } = renderHookWithQuery(() => useChangePass());
    expect(result.current.changePass).toBeInstanceOf(Function);
  });

  it('changePass calls mutationFn with provided data', async () => {
    const { result } = renderHookWithQuery(() => useChangePass());
    const mutationFn = mockFn.mock.results[0].value.mutationFn;
    await (result.current.changePass as any)({ data: { password: 'new' } });
    expect(mutationFn).toHaveBeenCalledWith({ data: { password: 'new' } }, expect.anything());
  });
});
