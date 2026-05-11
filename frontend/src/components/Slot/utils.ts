export const debounce = <T extends (...args: any[]) => any>(
  fn: T,
  delay: number,
): ((...args: Parameters<T>) => void) => {
  let timer: number;

  return (...args: Parameters<T>) => {
    clearTimeout(timer);
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    timer = setTimeout(() => {
      fn(...args);
    }, delay);
  };
};
