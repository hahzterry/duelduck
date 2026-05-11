import Link from 'next/link';

import styles from './styles.module.scss';

interface SeoPaginationProps {
  currentPage: number;
  baseUrl: string;
  maxPages?: number;
  totalPages?: number;
}

/**
 * SEO-friendly pagination component that renders actual <a> tags for crawlers.
 * This is rendered server-side and provides crawlable links to paginated content.
 */
export const SeoPagination = ({
  currentPage,
  baseUrl,
  maxPages = 50,
  totalPages,
}: SeoPaginationProps) => {
  const effectiveMaxPages = totalPages
    ? Math.min(totalPages, maxPages)
    : maxPages;

  // Generate page numbers to show
  // Show: 1, ..., current-1, current, current+1, ..., last
  const getPageNumbers = (): (number | 'ellipsis')[] => {
    const pages: (number | 'ellipsis')[] = [];

    if (effectiveMaxPages <= 7) {
      // Show all pages if total is small
      for (let i = 1; i <= effectiveMaxPages; i++) {
        pages.push(i);
      }

      return pages;
    }

    // Always show first page
    pages.push(1);

    if (currentPage > 3) {
      pages.push('ellipsis');
    }

    // Pages around current
    const start = Math.max(2, currentPage - 1);
    const end = Math.min(effectiveMaxPages - 1, currentPage + 1);

    for (let i = start; i <= end; i++) {
      pages.push(i);
    }

    if (currentPage < effectiveMaxPages - 2) {
      pages.push('ellipsis');
    }

    // Always show last page
    if (effectiveMaxPages > 1) {
      pages.push(effectiveMaxPages);
    }

    return pages;
  };

  const getPageUrl = (page: number): string => {
    if (page === 1) {
      return baseUrl;
    }

    return `${baseUrl}/page/${page}`;
  };

  const pageNumbers = getPageNumbers();
  const prevPage = currentPage > 1 ? currentPage - 1 : null;
  const nextPage = currentPage < effectiveMaxPages ? currentPage + 1 : null;

  return (
    <nav className={styles.seoPagination} aria-label="Pagination">
      {/* Previous link */}
      {prevPage ? (
        <Link
          href={getPageUrl(prevPage)}
          className={styles.seoPagination__arrow}
          rel="prev"
        >
          <span aria-hidden="true">&laquo;</span>
          <span className={styles.seoPagination__srOnly}>Previous page</span>
        </Link>
      ) : (
        <span
          className={`${styles.seoPagination__arrow} ${styles['seoPagination__arrow--disabled']}`}
        >
          <span aria-hidden="true">&laquo;</span>
        </span>
      )}

      {/* Page numbers */}
      <div className={styles.seoPagination__pages}>
        {pageNumbers.map((page, index) =>
          page === 'ellipsis' ? (
            <span
              key={`ellipsis-${index}`}
              className={styles.seoPagination__ellipsis}
            >
              &hellip;
            </span>
          ) : (
            <Link
              key={page}
              href={getPageUrl(page)}
              className={`${styles.seoPagination__page} ${
                page === currentPage
                  ? styles['seoPagination__page--active']
                  : ''
              }`}
              aria-current={page === currentPage ? 'page' : undefined}
            >
              {page}
            </Link>
          ),
        )}
      </div>

      {/* Next link */}
      {nextPage ? (
        <Link
          href={getPageUrl(nextPage)}
          className={styles.seoPagination__arrow}
          rel="next"
        >
          <span className={styles.seoPagination__srOnly}>Next page</span>
          <span aria-hidden="true">&raquo;</span>
        </Link>
      ) : (
        <span
          className={`${styles.seoPagination__arrow} ${styles['seoPagination__arrow--disabled']}`}
        >
          <span aria-hidden="true">&raquo;</span>
        </span>
      )}
    </nav>
  );
};
