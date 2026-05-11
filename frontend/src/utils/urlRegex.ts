const ipRegex = { v4: () => /(?:\d{1,3}\.){3}\d{1,3}/ };

export function urlRegex(options: { strict?: boolean; exact?: boolean } = {}) {
  options = {
    strict: true,
    ...options,
  };

  const protocol = `(?:(?:[a-z]+:)?//)${options.strict ? '' : '?'}`;
  const auth = '(?:\\S+(?::\\S*)?@)?';
  const ip = ipRegex.v4().source;
  const host = '(?:(?:[a-z\\u00a1-\\uffff0-9][-_]*)*[a-z\\u00a1-\\uffff0-9]+)';
  const domain =
    '(?:\\.(?:[a-z\\u00a1-\\uffff0-9]-*)*[a-z\\u00a1-\\uffff0-9]+)*';

  const tld = `(?:\\.(?:[a-z\\u00a1-\\uffff]{2,}))\\.?`;
  const port = '(?::\\d{2,5})?';
  const path = '(?:[/?#][^\\s"]*)?';
  const prefix = options.strict
    ? `(?:${protocol}|www\\.)`
    : `(?:${protocol}|www\\.)?`;
  const regex = `^${prefix}${auth}(?:localhost|${ip}|${host}${domain}${tld})${port}${path}`;

  return options.exact
    ? new RegExp(`(?:^${regex}$)`, 'i')
    : new RegExp(regex, 'ig');
}
