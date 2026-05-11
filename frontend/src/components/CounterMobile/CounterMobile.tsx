import { memo, ReactNode, useEffect, useMemo, useState } from 'react';
import cx from 'classnames';

import { CoinTypeIcon } from '~components/CoinTypeIcon';
import { SlotCounter } from '~components/Slot';
import { CURRENCY } from '~types/general';
import { formatNumberMobileWithModifyer } from '~utils/numbers';

import styles from './styles.module.scss';

interface CounterProps {
  amount: number;
  currency: CURRENCY;
  icon?: ReactNode;
}

enum CHANGE_TYPE {
  INCREASE = 'INCREASE',
  DECREASE = 'DECREASE',
}

export const CounterMobile = memo(
  ({ amount, currency, icon }: CounterProps) => {
    const [currentAmount, setCurrentAmount] = useState(amount);
    const [changeType, setChangeType] = useState<CHANGE_TYPE | null>(null);

    useEffect(() => {
      if (amount === currentAmount) return;

      setChangeType(
        amount > currentAmount ? CHANGE_TYPE.INCREASE : CHANGE_TYPE.DECREASE,
      );
      const timeout = setTimeout(() => setChangeType(null), 1000);

      setCurrentAmount(amount);

      return () => clearTimeout(timeout);
    }, [amount]);

    const { val, modifyer } = useMemo(
      () => formatNumberMobileWithModifyer(currentAmount),
      [currentAmount],
    );

    return (
      <div
        className={cx(styles.counter, {
          [styles.increase as string]: changeType === CHANGE_TYPE.INCREASE,
          [styles.decrease as string]: changeType === CHANGE_TYPE.DECREASE,
          [styles.zero as string]: amount === 0,
        })}
      >
        {icon || (
          <CoinTypeIcon isUsdc={currency === CURRENCY.USDC} isFullWidthIcons />
        )}
        <div className={styles.counterWrapper}>
          <SlotCounter text={`${val}`} triggerOnce />
        </div>
        {modifyer.length > 0 && (
          <span style={{ color: '#565656' }}>{modifyer}</span>
        )}
      </div>
    );
  },
);
