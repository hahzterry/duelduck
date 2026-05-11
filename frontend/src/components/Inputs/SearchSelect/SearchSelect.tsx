'use client';
import { CSSProperties, RefObject, useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import cx from 'classnames';
import { StaticImport } from 'next/dist/shared/lib/get-img-props';

import { FadeAnimation } from '~components/Animations/FadeAnimation';
import { LazyImage } from '~components/Lazy/LazyImage';
import { Typography } from '~components/Typography';
import { useOuterClick } from '~hooks/useOuterClick';
import empty from '~icons/empty-duck-logo.svg';
import { Cross } from '~icons/JsxSvg/Cross';
import { Search } from '~icons/JsxSvg/Search';
import search from '~icons/search-icon.svg';

import styles from './styles.module.scss';

interface Props {
  refContainerDropdown: RefObject<HTMLElement | null>;
  searchInputPlaceholder?: string;
  elements: {
    value: string;
    title: string;
    icon: string | StaticImport;
  }[];
  onSearchChange?: (value: string) => void;
  searchValue?: string;
  stylesSearchIcon?: CSSProperties;
  searchIcon?: string;
  textOpenButtonSearch?: string;
  buttonOpenStyles?: CSSProperties;
  emptyText?: string;
  emptyIcon?: string;
  onSelect?: (value: string) => void;
}

export const SearchSelect = ({
  refContainerDropdown,
  searchInputPlaceholder = 'Search here',
  elements,
  emptyText = 'Nothing was found, please try another SPL Token',
  onSearchChange,
  searchValue = '',
  searchIcon = search,
  onSelect,
  emptyIcon = empty,
  textOpenButtonSearch,
  stylesSearchIcon,
  buttonOpenStyles,
}: Props) => {
  const [isOpened, setIsOpened] = useState(false);

  const refInputText = useRef<HTMLInputElement | null>(null);
  const refDropdown = useRef<HTMLDivElement>(null);
  const refOpenButton = useRef<HTMLDivElement>(null);
  const [positionDropdownOpen, setPositionDropdownOpen] = useState<{
    width: string;
    top: string;
    left: string;
  } | null>(null);

  useEffect(() => {
    const handlePositionUpdate = () => {
      const openButton = refOpenButton.current;
      const container = refContainerDropdown.current;

      if (openButton && container) {
        setPositionDropdownOpen({
          top: `${openButton.getBoundingClientRect().y}px`,
          width: `${container.clientWidth}px`,
          left: `${container.getBoundingClientRect().x}px`,
        });
      }
    };

    handlePositionUpdate();
    const observer = new ResizeObserver(handlePositionUpdate);

    refOpenButton.current && observer.observe(refOpenButton.current);
    refContainerDropdown.current &&
      observer.observe(refContainerDropdown.current);

    document.body.addEventListener('scroll', handlePositionUpdate);

    return () => {
      document.body.removeEventListener('scroll', handlePositionUpdate);
      observer.disconnect();
    };
  }, [isOpened]);

  useOuterClick(
    refDropdown,
    () => {
      setIsOpened(false);
    },
    'click',
  );

  return (
    <>
      {typeof document !== 'undefined' &&
        createPortal(
          <FadeAnimation
            customWrapperClass={styles.searchSelect__dropdown}
            customRef={refDropdown}
            isVisible={isOpened}
            style={{ ...positionDropdownOpen }}
          >
            <div className={styles.searchSelect__inputSearch}>
              <div className={styles.searchSelect__input}>
                <Search />
                <input
                  placeholder={searchInputPlaceholder}
                  ref={refInputText}
                  value={searchValue}
                  onChange={(e) => onSearchChange?.(e.target.value.trim())}
                />
              </div>
              <Cross
                onClick={() => {
                  setIsOpened(false);
                  onSearchChange?.('');
                }}
              />
            </div>
            {elements.length ? (
              elements.map((e, i, arr) => (
                <div
                  className={cx(styles.searchSelect__elementContainer, {
                    [`${styles['searchSelect__elementContainer--last']}`]:
                      arr.length - 1 === i,
                    [`${styles['searchSelect__elementContainer--first']}`]:
                      0 === i,
                  })}
                  onClick={() => {
                    onSelect?.(e.value);
                    setIsOpened(false);
                  }}
                  key={i}
                >
                  <div className={styles.searchSelect__element}>
                    <div>
                      <LazyImage
                        src={e.icon}
                        width={18}
                        height={18}
                        alt={e.title}
                      />
                    </div>
                    <Typography text={e.title} />
                  </div>
                </div>
              ))
            ) : (
              <div className={styles.searchSelect__empty}>
                <LazyImage src={emptyIcon} alt={'No elements'} />
                <Typography text={emptyText} />
              </div>
            )}
          </FadeAnimation>,
          document.body,
        )}
      <div
        className={styles.searchSelect}
        ref={refOpenButton}
        style={buttonOpenStyles}
        onClick={() => {
          setIsOpened(true);
          setTimeout(() => {
            refInputText.current?.focus();
          }, 100);
        }}
      >
        <LazyImage
          src={searchIcon}
          alt={'Search'}
          style={stylesSearchIcon}
          className={styles.searchSelect__iconSearch}
        />
        {!!textOpenButtonSearch && (
          <Typography
            text={textOpenButtonSearch}
            className={styles.searchSelect__text}
          />
        )}
      </div>
    </>
  );
};
