export function validatePageParam(
  pageParam: string | undefined,
): number | null {
  if (!pageParam) return 1;

  const parsed = parseInt(pageParam, 10);

  if (isNaN(parsed) || parsed < 1 || !Number.isInteger(Number(pageParam))) {
    return null;
  }

  return parsed;
}
