import { IS_STAGE } from '~constants/api';

// export const STAGE_AC_HEADER_NAME = 'x-stage-ac';
//
// export const STAGE_AC_HEADER_VALUE = 'run-stage-request';

export const getStageRequestHeaders = (): Record<string, string> => {
  if (!IS_STAGE) {
    return {};
  }

  return {};
};

export const withStageRequestHeaders = (headers?: HeadersInit): HeadersInit => {
  if (!IS_STAGE) {
    return headers || {};
  }

  const stageHeaders = getStageRequestHeaders();

  if (!headers) {
    return stageHeaders;
  }

  if (headers instanceof Headers) {
    const nextHeaders = new Headers(headers);

    Object.entries(stageHeaders).forEach(([key, value]) => {
      nextHeaders.set(key, value);
    });

    return nextHeaders;
  }

  if (Array.isArray(headers)) {
    return [...headers, ...Object.entries(stageHeaders)];
  }

  return {
    ...headers,
    ...stageHeaders,
  };
};
