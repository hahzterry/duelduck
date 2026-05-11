import { isValidElement } from 'react';

import { DUEL_DUCK_API } from '~constants/api';

export const pxToRem = (px: number) => `${px / 16}rem`;

export const isUUID = (str: string): boolean => {
  const uuidPattern: RegExp =
    /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/;

  return uuidPattern.test(str);
};

export const isDeadlineEnd = (deadline: string) => {
  const now = new Date();
  const deadlineDate = new Date(deadline);

  return now > deadlineDate;
};

export const createMediaLink = (image_url?: string) =>
  image_url ? `${DUEL_DUCK_API}/media${image_url}` : '/skeletonImage.svg';

export const createMediaStaticLink = (image_url?: string) =>
  image_url
    ? image_url?.startsWith('/user_uploads')
      ? `${DUEL_DUCK_API}/media${image_url}`
      : `${DUEL_DUCK_API}/static${image_url}`
    : '/skeletonImage.svg';

export const createUserAvatar = (
  image_url?: string,
  isBackground?: boolean,
) => {
  if (image_url?.startsWith('https')) {
    return image_url;
  }

  return image_url
    ? createMediaStaticLink(image_url)
    : isBackground
      ? '/publicProfileBg.svg'
      : '/duckProfile.svg';
};

export const isArrayOfReactElements = (arr: any[]) => {
  return Array.isArray(arr) && arr.every(isValidElement);
};

export function isReactNode(value: any): boolean {
  return (
    value === null ||
    value === undefined ||
    typeof value === 'boolean' ||
    typeof value === 'string' ||
    typeof value === 'number' ||
    isValidElement(value) ||
    (Array.isArray(value) && value.every(isReactNode))
  );
}

export const getTextWidth = (text: string, fontParams: string) => {
  const windowAny = window as any;

  if (!windowAny.textCanvas) {
    windowAny.textCanvas = document.createElement('canvas');
  }

  const context = windowAny.textCanvas.getContext('2d');

  context.font = fontParams;

  return context.measureText(text).width;
};
