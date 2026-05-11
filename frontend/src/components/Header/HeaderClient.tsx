'use client';

import './Header.scss';

import cx from 'classnames';
import Link from 'next/link';

import { HeaderBorderedButton } from '~components/Buttons/HeaderBorderedButton';
import { Links } from '~components/Header/components/Links/Links';
import { BigDuck } from '~icons/JsxSvg/BigDuck';
import { Login } from '~icons/JsxSvg/Login';
import { useSidebarStore } from '~store/sidebarStore';

import styles from './styles.module.scss';

export const HeaderClient = () => {
  const { isSidebarOpen } = useSidebarStore();

  return (
    <div
      className={cx('header', {
        ['siderOpen']: isSidebarOpen,
      })}
    >
      <div className="header__left">
        <Link aria-label="Duel Duck Logo" className="logo" href={'/'}>
          <BigDuck />
        </Link>
        <Links />
      </div>
      <div className={'header__right'}>
        <HeaderBorderedButton
          text={'Play now'}
          className={styles.playNow}
          icon={<Login />}
        />
      </div>
    </div>
  );
};
