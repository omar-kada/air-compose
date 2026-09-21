import '@testing-library/jest-dom';

class ResizeObserverMock {
  observe = () => undefined;
  unobserve = () => undefined;
  disconnect = () => undefined;
}

vi.stubGlobal('ResizeObserver', ResizeObserverMock);
