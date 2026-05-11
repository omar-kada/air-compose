import { useLogin, useRegisteration } from '@/hooks';
import { useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { LoginForm } from './login';
import { Button } from './ui/button';
import { Card, CardContent } from './ui/card';
import { TextSeparator } from './view';

export function LoginPage() {
  const { t } = useTranslation();

  const { login, isPending } = useLogin();
  const { data: registration, isPending: registrationPending } = useRegisteration();
  const serverUrl = import.meta.env.VITE_SERVER_URL;

  const navigate = useNavigate();

  const loginWithOidc = useCallback(() => {
    navigate(`${serverUrl}/api/oidc/login`);
  }, []);

  return (
    <div className="p-4 space-y-4 h-full flex items-center flex-col justify-center">
      <h2 className="text-xl">{t('LOGIN.FORM.TITLE')}</h2>
      <Card>
        <CardContent className="space-y-4">
          {registration?.oidc && (
            <>
              <div className="flex w-full justify-around">
                <Button onClick={loginWithOidc}>{t('LOGIN.FORM.LOGIN_WITH_OIDC')}</Button>
              </div>
              <TextSeparator text={t('LOGIN.FORM.OR')} />
            </>
          )}
          <LoginForm
            onSubmit={login}
            className="m-auto max-w-lg flex-1"
            loading={isPending || registrationPending}
          ></LoginForm>
        </CardContent>
      </Card>
    </div>
  );
}
