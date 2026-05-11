import { RefObject } from 'react';

import { FadeAnimation } from '~components/Animations/FadeAnimation';
import { Typography } from '~components/Typography';
import colors from '~styles/colors';
import { ColorType } from '~types/general';

import styles from './styles.module.scss';

export type PopupMessageTexts = { color: ColorType; text: string }[];

interface Props {
  texts?: PopupMessageTexts;
  isOpen: boolean;
  ref?: RefObject<HTMLDivElement | null>;
}

export const MessagePopup = ({ texts, isOpen, ref }: Props) => {
  console.log(texts, isOpen);

  return (
    <FadeAnimation
      customWrapperClass={styles.messagePopup}
      isVisible={isOpen}
      customRef={ref}
    >
      <span>
        {texts?.map((e, i) => (
          <Typography text={e.text} color={colors[e.color]} key={i} />
        ))}
      </span>
    </FadeAnimation>
  );
};
