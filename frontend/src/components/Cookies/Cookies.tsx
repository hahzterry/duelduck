import { useState } from 'react';

import { ButtonBase } from '~components/Buttons/ButtonBase';
import { Typography } from '~components/Typography';
import { Cookie } from '~icons/JsxSvg/Cookie';

import styles from './styles.module.scss';

const popupKey = 'showPopup';

export const Cookies = () => {
  const getDefaultShowPopup = () => {
    const showPopupValue = localStorage.getItem(popupKey);

    return showPopupValue !== 'false';
  };

  const [showPopup, setShowPopup] = useState(getDefaultShowPopup());

  const handleAccept = () => {
    localStorage.setItem(popupKey, 'false');
    setShowPopup(false);
  };

  const handleClose = () => {
    localStorage.setItem(popupKey, 'false');
    setShowPopup(false);
  };

  if (!showPopup) {
    return null;
  }

  return (
    <div className={styles.cookiesWrapper} id={'cookies-popup'}>
      <div className={styles.top}>
        <Cookie />
        <Typography>
          Ducks enjoy cookies. Share yours to improve your experience on our
          website.
        </Typography>
      </div>
      <div className={styles.actions}>
        <ButtonBase
          text={'Close'}
          onClick={handleClose}
          className={styles.close}
        />
        <ButtonBase text={'Accept all'} onClick={handleAccept} />
      </div>
    </div>
  );
};
