import { ChangeEvent, FormEventHandler, useState } from 'react';

import { RoundedInput } from '~components/Inputs/RoundedInput';
import { Typography } from '~components/Typography';
import { emailRegex } from '~constants/main';
import { useSidebarStore } from '~store/sidebarStore';
import { SIDEBAR_CONTENT } from '~types/general';

import styles from './styles.module.scss';

export const SignInWithEmail = () => {
  const [error, setError] = useState(false);
  const { setEmail, email, openSidebar } = useSidebarStore();

  const handleSubmit: FormEventHandler<HTMLFormElement> = async (e) => {
    e.preventDefault();
    if (!email?.length || !emailRegex.test(email)) {
      setError(true);

      return;
    }

    openSidebar(SIDEBAR_CONTENT.CONFIRM_EMAIL);
  };

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    setEmail(e.target.value);
    setError(false);
  };

  return (
    <div className={styles['content']}>
      <div className={styles['content__titles']}>
        <Typography text="Sign in with Email" />
        <Typography text="Please enter your email and we will send you a confirmation code" />
      </div>
      <RoundedInput
        placeholder="Enter your email"
        value={email || ''}
        onChange={handleChange}
        onSubmit={handleSubmit}
        isError={error}
        errorMsg={error ? 'Something went wrong' : ''}
      />
    </div>
  );
};
