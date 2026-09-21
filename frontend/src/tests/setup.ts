import '@testing-library/jest-dom';

class ResizeObserverMock {
  observe() {
    void this;
  }
  unobserve() {
    void this;
  }
  disconnect() {
    void this;
  }
}

vi.stubGlobal('ResizeObserver', ResizeObserverMock);
