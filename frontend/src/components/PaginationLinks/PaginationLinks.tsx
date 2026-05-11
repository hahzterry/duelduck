interface PaginationLinksProps {
  prevUrl?: string;
  nextUrl?: string;
}

export const PaginationLinks = ({ prevUrl, nextUrl }: PaginationLinksProps) => {
  return (
    <>
      {prevUrl && <link rel="prev" href={prevUrl} />}
      {nextUrl && <link rel="next" href={nextUrl} />}
    </>
  );
};
