export const sortCurrencyArray = (arr: any[]) => {
  return arr.sort((a, b) => {
    if (a.title === 'DUCK POINTS') return -1;
    if (b.title === 'DUCK POINTS') return 1;
    if (a.title === 'SOL') return -1;
    if (b.title === 'SOL') return 1;
    if (a.title === 'USDC') return -1;
    if (b.title === 'USDC') return 1;

    return 0;
  });
};

export const sortTokensArray = (arr: any[]) => {
  return arr.sort((a, b) => {
    if (a.tokenSymbol === 'SOL') return -1;
    if (b.tokenSymbol === 'SOL') return 1;
    if (a.tokenSymbol === 'USDC') return -1;
    if (b.tokenSymbol === 'USDC') return 1;
    if (a.tokenSymbol === 'DUCK POINTS') return -1;
    if (b.tokenSymbol === 'DUCK POINTS') return 1;

    return 0;
  });
};

export const firstLetterToUpperCase = (letter: string) =>
  letter.replace(/(^\w|\s\w)/g, (m) => m.toUpperCase());
