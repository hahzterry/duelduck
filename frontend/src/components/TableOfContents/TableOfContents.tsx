import type { CSSProperties, MouseEvent } from 'react';
import { useEffect, useRef } from 'react';
import cx from 'classnames';
import NextLink from 'next/link';

import styles from './styles.module.scss';

export type TocLink = { label: string; id: string };

interface Props {
  title: string;
  links: TocLink[];
  activeLinkId?: string | null;
  wrapperClass?: string;
  wrapperStyle?: CSSProperties;
  titleClass?: string;
  titleStyle?: CSSProperties;
  listClass?: string;
  listStyle?: CSSProperties;
}

export default function TableOfContents({
  title,
  links,
  activeLinkId,
  wrapperClass,
  wrapperStyle,
  titleClass,
  titleStyle,
  listClass,
  listStyle,
}: Props) {
  const list = useRef<HTMLUListElement>(null);

  useEffect(() => {
    if (!list.current) {
      return;
    }

    const item = list.current.querySelector(`[data-id="${activeLinkId}"]`);

    if (!item) {
      return;
    }

    list.current.scrollTo({
      top:
        item.getBoundingClientRect().top -
        list.current.getBoundingClientRect().top +
        list.current.scrollTop -
        list.current.clientHeight / 2,
      behavior: 'smooth',
    });
  }, [activeLinkId]);

  const handleLinkClick = (
    event: MouseEvent<HTMLAnchorElement>,
    id: string,
  ) => {
    event.preventDefault();

    const element = document.getElementById(id);

    if (!element) {
      return;
    }

    const headerHeight = document.querySelector('header')?.clientHeight ?? 0;

    document.body.scrollTo({
      top:
        element.getBoundingClientRect().top +
        document.body.scrollTop -
        headerHeight,
      behavior: 'smooth',
    });
  };

  return (
    <div className={cx(styles.wrapper, wrapperClass)} style={wrapperStyle}>
      <p className={cx(styles.title, titleClass)} style={titleStyle}>
        {title}
      </p>
      <ul ref={list} className={cx(styles.list, listClass)} style={listStyle}>
        {links.map((link) => (
          <li key={link.id} data-active={activeLinkId === link.id}>
            <NextLink
              data-id={link.id}
              href={`#${link.id}`}
              onClick={(e) => handleLinkClick(e, link.id)}
            >
              {link.label}
            </NextLink>
          </li>
        ))}
      </ul>
    </div>
  );
}
