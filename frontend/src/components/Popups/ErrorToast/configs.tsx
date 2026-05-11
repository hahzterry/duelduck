import { ReactNode } from 'react';

import { ErrorToastData, ErrorToastTextPart } from '~store/errorToastStore';
import colors from '~styles/colors';

export const configBottomCircleClose = ({
  textParts,
  additionalContent,
  titleParts,
}: {
  additionalContent?: ReactNode;
  textParts?: ErrorToastTextPart[];
  titleParts?: ErrorToastTextPart[];
}): ErrorToastData => {
  return {
    backgroundColor: colors.backgroundSecond,
    textColor: colors.textSecondary,
    accentTextColor: colors.textPrimary,
    stylesToast: {
      padding: '24px',
    },
    stylesContainer: {
      borderRadius: '15px',
      width: '324px',
      minWidth: 'min(280px, 100% - 40px)',
      maxWidth: 'min(324px, 100% - 40px)',
    },
    progressColor: colors.primary,
    additionalContent,
    progressType: 'cross-circle',
    positionClose: 'outside-bottom',
    timeClose: 5000,
    titleParts,
    textParts,
  };
};
