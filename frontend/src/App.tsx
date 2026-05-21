import { useQuery } from '@tanstack/react-query';
import { useEffect, useState, type ReactNode } from 'react';
import { BrowserRouter, Navigate, Route, Routes, useLocation, useNavigate } from 'react-router-dom';
import {
  ConfigPage,
  DeploymentsPage,
  EnvironementHealth,
  ErrorAlert,
  InitPage,
  LoginPage,
  NavBar,
  RegisterPage,
  StatusPage,
  Topbar,
} from './components';
import { getStateQueryOptions, useRegisteration, useUser } from './hooks';
import { cn, ROUTES } from './lib';

function RouteBasedTopBar({ children }: { children: ReactNode }) {
  const { data: registration, isPending, error } = useRegisteration();
  const { data: user, isPending: userPending, error: userError } = useUser();
  const { data: state, isPending: statePending } = useQuery(
    getStateQueryOptions({
      enabled: !!user,
    }),
  );
  const navigate = useNavigate();
  const location = useLocation();

  const [showTopBar, setShowTopBar] = useState(!!user);

  useEffect(() => {
    setShowTopBar(!!registration?.registered && !!user && state?.initialized === true);
  }, [setShowTopBar, registration, user, state]);

  useEffect(() => {
    const waitAndNavigate = (pending: boolean, condition: boolean, route: string) => {
      if (!pending && condition && location.pathname !== route) {
        setLastRoute(location.pathname);
        navigate(route);
      }
      return pending || condition;
    };
    if (
      waitAndNavigate(isPending, !registration?.registered, ROUTES.REGISTER) ||
      waitAndNavigate(userPending, !user, ROUTES.LOGIN) ||
      waitAndNavigate(statePending, !state?.initialized, ROUTES.INIT)
    ) {
      return;
    }
    if ([ROUTES.INIT, ROUTES.REGISTER, ROUTES.LOGIN].includes(location.pathname)) {
      navigate(getLastRoute() || ROUTES.DEPLOYMENTS);
    }
  }, [navigate, location, registration, user, isPending, userPending, state, statePending]);

  const mergedError = error ?? userError;
  return (
    <div className={cn('flex flex-col h-dvh', showTopBar ? 'pb-12 md:pb-0' : '')}>
      {showTopBar && (
        <>
          <Topbar className="max-w-7xl mx-4">
            {/* Top navigation bar, on big screens */}
            <div className="flex">
              <NavBar className="hidden md:flex bg-sidebar items-center flex-1" />
            </div>
            <EnvironementHealth></EnvironementHealth>
          </Topbar>
          {/* Bottom navigation bar, on small screens */}
          <NavBar className="flex md:hidden bg-sidebar h-12 border-t w-full fixed items-center justify-around bottom-0 left-0 right-0 z-50" />
        </>
      )}
      <ErrorAlert title={mergedError?.message ?? null} />
      <div className="w-full flex justify-around min-h-0 h-full">
        <div className="max-w-7xl flex-1 mx-4 ">{children}</div>
      </div>
    </div>
  );
}

function App() {
  return (
    <BrowserRouter>
      <RouteBasedTopBar>
        <Routes>
          <Route path={ROUTES.ROOT} element={<Navigate to={ROUTES.DEPLOYMENTS}></Navigate>} />
          <Route path={ROUTES.INIT} element={<InitPage />} />
          <Route path={ROUTES.REGISTER} element={<RegisterPage />} />
          <Route path={ROUTES.LOGIN} element={<LoginPage />} />
          <Route path={ROUTES.DEPLOYMENTS} element={<DeploymentsPage />} />
          <Route path={ROUTES.DEPLOYMENT(':id')} element={<DeploymentsPage />} />
          <Route path={ROUTES.STATUS} element={<StatusPage />} />
          <Route path={ROUTES.LOGS} element={<div> logs </div>} />
          <Route path={ROUTES.CONFIG} element={<ConfigPage />} />
        </Routes>
      </RouteBasedTopBar>
    </BrowserRouter>
  );
}

const LAST_ROUTE = 'LAST_ROUTE';
function setLastRoute(route: string) {
  if (![ROUTES.REGISTER, ROUTES.LOGIN].includes(route)) {
    localStorage.setItem(LAST_ROUTE, route);
  }
}

function getLastRoute(): string | null {
  return localStorage.getItem(LAST_ROUTE);
}

export default App;
