import { ChangeEventHandler } from 'react';

import { CoinTypeIcon } from '~components/CoinTypeIcon';
import { SwitchIOS } from '~components/SwitchIOS';
import { CURRENCY } from '~types/general';

import styles from './styles.module.scss';

interface Props {
  isChecked: boolean;
  id: string;
  onChange?: ChangeEventHandler<HTMLInputElement>;
  currency?: CURRENCY;
}

export const CurrencySwitch = ({
  isChecked,
  onChange,
  id,
  currency,
}: Props) => {
  return (
    <div className={styles.switch}>
      <CoinTypeIcon
        isUsdc={currency === CURRENCY.USDC}
        isWbt={currency === CURRENCY.WBT}
      />
      <SwitchIOS id={id} isChecked={isChecked} onChange={onChange} />
    </div>
  );
};
