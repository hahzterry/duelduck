import colors from '~styles/colors';

export const roundToLastNonZeroDecimal = (num: number): number => {
  const numStr = num.toString();
  const lastNonZeroIndex = numStr
    .split('')
    .findIndex((char) => char !== '0' && char !== '.');
  const decimalIndex = numStr.indexOf('.');

  if (decimalIndex === -1 || lastNonZeroIndex === -1) {
    return num;
  }

  return parseFloat(num.toFixed(lastNonZeroIndex));
};

type NumberScale = 1e3 | 1e6 | 1e9;

export const formatNumber = (
  num: number = 0,
  minThreshold: number = 0,
  scale: NumberScale = 1e3,
): string => {
  const isNegative = num < 0;
  const absNum = Math.abs(num);

  if (num === 0) return '0';

  if (absNum < minThreshold) {
    return isNegative ? `-${absNum}` : absNum.toString();
  }

  let suffix = '';
  let value = absNum;

  if (scale === 1e9 && absNum >= 1e9) {
    value = absNum / 1e9;
    suffix = 'B';
  } else if (scale === 1e6 && absNum >= 1e6) {
    value = absNum / 1e6;
    suffix = 'M';
  } else if (scale === 1e3 && absNum >= 1e3) {
    value = absNum / 1e3;
    suffix = 'K';
  }

  const result = value.toFixed(0).replace(/\.0$/, '') + suffix;

  return isNegative ? `-${result}` : result;
};

export const formatNumberMobile = (
  num: number = 0,
  maximumFractionDigits?: number,
): string => {
  maximumFractionDigits =
    typeof maximumFractionDigits === 'undefined' ? 2 : maximumFractionDigits;
  const isNegative = num < 0;
  const absNum = Math.abs(num);

  let resVal;

  if (num.toString().length < 3) {
    return isNegative
      ? '-' + Math.abs(num).toLocaleString('en-US', { maximumFractionDigits })
      : num.toLocaleString('en-US', { maximumFractionDigits });
  }

  if (absNum >= 1e9) {
    resVal =
      (absNum / 1e9).toLocaleString('en-US', { maximumFractionDigits }) + 'B';
  } else if (absNum >= 1e6) {
    resVal =
      (absNum / 1e6).toLocaleString('en-US', { maximumFractionDigits }) + 'M';
  } else if (absNum >= 1e3) {
    resVal =
      (absNum / 1e3).toLocaleString('en-US', { maximumFractionDigits }) + 'K';
  } else {
    resVal = absNum.toLocaleString('en-US', { maximumFractionDigits });
  }

  return isNegative ? '-' + resVal : (resVal as string);
};

export const formatNumberMobileWithModifyer = (
  num: number = 0,
  maximumFractionDigits?: number,
): { val: string; modifyer: string } => {
  maximumFractionDigits =
    typeof maximumFractionDigits === 'undefined' ? 2 : maximumFractionDigits;

  if (num.toString().length < 3) {
    return {
      val: num.toLocaleString('en-US', { maximumFractionDigits }),
      modifyer: '',
    };
  }

  let resVal;
  let modifyer = '';

  if (num >= 1e9) {
    resVal = (num / 1e9).toFixed(0).replace(/\.0$/, '');
    modifyer = 'B';
  } else if (num >= 1e6) {
    resVal = (num / 1e6).toFixed(0).replace(/\.0$/, '');
    modifyer = 'M';
  } else if (num >= 1e3) {
    resVal = (num / 1e3).toFixed(1).replace(/\.0$/, '');
    modifyer = 'K';
  } else {
    resVal = num.toLocaleString('en-US', { maximumFractionDigits });
  }

  return { val: resVal as string, modifyer };
};

export const addSpaceInPrice = (price: number) =>
  price
    .toLocaleString('en-US', {
      maximumFractionDigits: 2,
    })
    .replace(/\B(?=(\d{3})+(?!\d))/g, ' ');

export const formatReward = (value: number): string => {
  if (value === 0) return '0';
  if (value >= 0.01) return formatNumberMobile(value, 2);

  return '<0.01';
};

export const formatBalance = (
  value: number,
  {
    ifHigherThanValue = 1000,
    minimumValueFractionDigits = 1,
    maximumValueFractionDigits = 4,
  } = {
    ifHigherThanValue: 1000,
    minimumValueFractionDigits: 1,
    maximumValueFractionDigits: 4,
  },
) => {
  return value.toLocaleString('en-US', {
    maximumFractionDigits:
      value > ifHigherThanValue
        ? minimumValueFractionDigits
        : maximumValueFractionDigits,
  });
};

export const clamp = (v: number, min: number, max: number) =>
  Math.min(max, Math.max(min, v));

export const formatQuestionPrice = (price: number): string => {
  const decimalPart = price % 1 === 0 ? '' : price.toString().split('.')[1];
  const integerPart = Number(price.toString().split('.')[0]);

  return `${integerPart.toLocaleString('en-us')}${decimalPart ? `.${decimalPart}` : ''}`;
};

export function toDynamicFix(value: number | string): number {
  if (isNaN(+value)) return 0;

  const num = parseFloat(String(value));

  if (Number.isInteger(num)) return num;

  const stringValue = num.toString();
  const decimalIndex = stringValue.indexOf('.');

  if (decimalIndex === -1) return num;

  const fractionalPart = stringValue.slice(decimalIndex + 1);

  const firstSignificantIndex = fractionalPart.search(/[^0]/);

  if (firstSignificantIndex === -1) return num;

  const fix = firstSignificantIndex + 4;

  const multiplier = 10 ** fix;

  return Math.floor(num * multiplier) / multiplier;
}

export const getColorNumber = (num: number) => {
  if (num === 0) return colors.primary;
  else if (num > 0) return colors.success;
  else return colors.error;
};
