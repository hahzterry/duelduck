import { useEffect, useState } from 'react';

import { TextToast } from '~components/Popups/ErrorToast/components/TextToast';
import { useRefAndState } from '~hooks/useRefAndState';
import { ErrorToastTextPart } from '~store/errorToastStore';
import { getTextWidth } from '~utils/general';

import styles from './styles.module.scss';

const getStringText = (text: ErrorToastTextPart[] | string) =>
  typeof text === 'string' ? text : text.reduce((acc, t) => acc + t.text, '');

export const AnimatedChangeTextComponent = ({
  text,
  font,
}: {
  text: ErrorToastTextPart[] | string;
  font: {
    fontFamily: string;
    fontWeight: string;
    fontSize: string;
  };
}) => {
  const [maxWidth, setMaxWidth] = useState(0);
  const [getOldText, setOldText, oldText] = useRefAndState(text);

  useEffect(() => {
    const newText = getStringText(text);
    const oldTextString = getStringText(getOldText('ref'));

    if (newText.length > oldTextString.length) {
      setOldText(text);
    }

    setMaxWidth(
      getTextWidth(
        newText,
        `${font.fontWeight} ${font.fontSize} ${font.fontFamily}`,
      ),
    );
  }, [text]);

  return (
    <span
      onTransitionEnd={() => {
        setOldText(text);
      }}
      style={{
        maxWidth,
      }}
      className={styles.animatedChangeTextComponent}
    >
      {typeof oldText === 'string' ? oldText : <TextToast parts={oldText} />}
    </span>
  );
};
