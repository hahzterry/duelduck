import { regexLatinAndSymbols } from '~constants/main';
import { Options } from '~types/options';

export const shortenString = (
  str: string,
  maxLength: number,
  countSymbolsAfterDots: number,
) => {
  if (str.length <= maxLength) return str;

  const prefixLength = Math.max(0, maxLength - (3 + countSymbolsAfterDots));
  const prefix = str.slice(0, prefixLength);
  const suffix = countSymbolsAfterDots ? str.slice(-countSymbolsAfterDots) : '';

  return `${prefix}...${suffix}`;
};

export const capitalizeFirst = (str: string) =>
  str ? str[0]?.toUpperCase() + str.slice(1) : '';

export const optionsToSearchParams = (options: Options) => {
  const params = new URLSearchParams();

  if (options?.pagination) {
    if (options.pagination.page_size !== undefined) {
      params.append(
        'opts.pagination.page_size',
        options.pagination.page_size.toString(),
      );
    }

    if (options.pagination.page_num !== undefined) {
      params.append(
        'opts.pagination.page_num',
        options.pagination.page_num.toString(),
      );
    }
  }

  if (options?.order) {
    params.append('opts.order.order_by', options.order.order_by);
    params.append('opts.order.order_type', options.order.order_type);
  }

  if (options?.filters?.length) {
    options.filters.forEach((filter, index) => {
      params.append(`opts.filters[${index}].column`, filter.column);
      params.append(`opts.filters[${index}].operator`, filter.operator);
      params.append(`opts.filters[${index}].value`, filter.value);
      params.append(
        `opts.filters[${index}].where_or`,
        filter.where_or.toString(),
      );
    });
  }

  if (options?.challenge_id) {
    params.append('challenge_id', options.challenge_id);
  }

  return params.toString();
};

export const validateUsername = (value: string) => {
  if (value.length <= 3) return 'Username must be at least 3 characters';
  else if (value.length > 24)
    return 'Username must be less than 24 characters long';
  else if (!regexLatinAndSymbols.test(value))
    return 'Please use only letters, numbers, dots (.), underscores (_), hyphens (-), and spaces';
  else if (/^\s|\s$/.test(value))
    return 'Username cannot start or end with a space';
  else if (/\s{2,}/.test(value))
    return 'Username cannot contain consecutive spaces';

  return '';
};

export const capitalize = (str: string): string => {
  if (!str) return str;

  return str.charAt(0).toUpperCase() + str.slice(1);
};
