import './linkStyles.scss';

import { memo, useEffect, useState } from 'react';
import axios from 'axios';
import cx from 'classnames';
import Image from 'next/image';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

import { LazyImage } from '~/components/Lazy/LazyImage';
import { ResponsibleGamingBanner } from '~/components/ResponsibleGamingBanner';
import bigDuckSvg from '~/icons/bigduck.svg';
import { ButtonBase } from '~components/Buttons/ButtonBase';
import { SocialLink } from '~components/Buttons/SocialLink';
import { Typography } from '~components/Typography';
import useMediaQuery from '~hooks/useMediaQuery';
import { DiscordIcon } from '~icons/JsxSvg/DiscordIcon';
import { Dune } from '~icons/JsxSvg/Dune';
import { Dune2 } from '~icons/JsxSvg/Dune2';
import { Dune2Active } from '~icons/JsxSvg/Dune2Active';
import { DuneActive } from '~icons/JsxSvg/DuneActive';
import { Instagram } from '~icons/JsxSvg/Instagram';
import { OverdoneMonochrome } from '~icons/JsxSvg/OverdoneMonochrome';
import { TelegramIcon } from '~icons/JsxSvg/Telegram';
import { Wiki } from '~icons/JsxSvg/Wiki';
import { X } from '~icons/JsxSvg/X';
import { Youtube } from '~icons/JsxSvg/Youtube';
import colors from '~styles/colors';
import { CUSTOM_EVENT_KEYS } from '~types/general';
import { customEvent } from '~utils/customEvent';

import Copyright from './components/Copyright/Copyright';
import styles from './styles.module.scss';

const footerSections = [
  {
    title: 'Platform',
    links: [
      { label: 'Home', href: '/' },
      { label: 'Duels', href: '/duels' },
      { label: 'Tournaments', href: '/tournaments' },
      { label: 'Create a Duel', href: '/create-duel' },
      { label: 'API', href: '/api' },
    ],
  },
  {
    title: 'EXPLORE',
    links: [
      { label: 'Public FAQ', href: '/faq' },
      { label: 'Blog', href: '/blog' },
    ],
  },
  {
    title: 'Company',
    links: [
      { label: 'Branding', href: '/branding' },
      { label: 'Platform Metrics', href: '/dashboard' },
      { label: 'Legal & Policies', href: '/terms-and-conditions' },
    ],
  },
];

const socials = [
  {
    href: 'https://t.me/duelduck',
    icon: <TelegramIcon />,
    text: 'Telegram',
    hoverBackgroundColor: '#4C9AD3',
    hoverIconColor: 'rgba(255, 255, 255, 1)',
  },
  {
    href: 'https://www.youtube.com/@DuelDuck',
    icon: <Youtube />,
    text: 'Youtube',
    hoverBackgroundColor: 'white',
    hoverIconColor: 'rgba(234, 51, 61, 1)',
  },
  {
    href: 'https://x.com/duel_duck',
    icon: <X />,
    text: 'X (Twitter)',
    hoverBackgroundColor: '#000',
    hoverIconColor: 'white',
  },
  {
    href: 'https://discord.gg/bDdTRheD2a',
    icon: <DiscordIcon />,
    text: 'Discord',
    hoverBackgroundColor: '#5F6AEB',
    hoverIconColor: 'rgba(255, 255, 255, 1)',
  },
  {
    href: 'https://dune.com/duelduck/duel-duck',
    icon: <Dune2 />,
    text: 'Dune',
    hoverBackgroundColor: 'white',
    hoverIcon: <Dune2Active />,
  },
];

const socialsStas = [
  {
    href: 'https://x.com/stan_duel_duck',
    icon: <X />,
    text: 'X',
    hoverBackgroundColor: '#000',
    hoverIconColor: 'white',
  },
  {
    href: 'https://www.instagram.com/iamhoruna',
    icon: <Instagram />,
    text: 'Instagram',
    hoverBackgroundColor:
      'radial-gradient(circle at 30% 107%, #fdf497 0%, #fdf497 5%, #fd5949 45%,#d6249f 60%,#285AEB 90%)',
    hoverIconColor: 'white',
  },
  {
    href: 'https://en.wikipedia.org/wiki/Stanislav_Horuna',
    icon: <Wiki />,
    text: 'Wiki',
    hoverBackgroundColor: 'white',
    hoverIconColor: '#000',
  },
];

const isNeedLogin = ['/tasks', '/create-duel'];

const pageHideFooter: string[] = [];

const TELEGRAM_TOKEN = '7548392247:AAFU7WB3FmY-gFNUOp2SjG96xggKjHAIDdw';
const TELEGRAM_CHAT_ID = -4759031640;

const FooterLeftMainPage = () => {
  const [contact, setContact] = useState<string>('');
  const [comment, setComment] = useState<string>('');
  const [isDisabled, setIsDisabled] = useState<boolean>(false);
  const [isUserInput, setIsUserInput] = useState<boolean>(false);

  const [sending, setSending] = useState(false);

  useEffect(() => {
    setIsDisabled(contact.length < 3);
  }, [contact]);

  const sendToTelegram = async (message: string) => {
    const url = `https://api.telegram.org/bot${TELEGRAM_TOKEN}/sendMessage`;
    const params = {
      chat_id: TELEGRAM_CHAT_ID,
      text: message,
      parse_mode: 'HTML',
    };

    try {
      await axios.post(url, params);

      return true;
    } catch (error) {
      console.error('Telegram send error:', error);

      return false;
    }
  };

  const handleSubmit = async () => {
    setIsUserInput(true);
    if (isDisabled) return;
    setSending(true);

    const message = 'Contact: ' + contact + '\n' + 'Message: ' + comment;

    const success = await sendToTelegram(message);

    if (success) {
      setContact('');
      setComment('');
    }

    setSending(false);
  };

  return (
    <div className={styles.footerLeft}>
      <div className={styles.top}>
        <LazyImage
          loading="lazy"
          alt={'Stan'}
          className={styles.image}
          src="/StanFooter.webp"
          width={150}
          height={150}
        />
        <div className={styles.info}>
          <div className={styles.nameAndPosition}>
            <Typography className={styles.name}>Stan horuna</Typography>
            <Typography className={styles.position}>DuelDuck CEO</Typography>
          </div>
          <div className={styles.bio}>
            {socialsStas.map(
              ({ href, icon, hoverIconColor, hoverBackgroundColor }, i) => (
                <SocialLink
                  key={i}
                  href={href}
                  icon={icon}
                  hoverIconColor={hoverIconColor}
                  hoverBackgroundColor={hoverBackgroundColor}
                />
              ),
            )}
          </div>
        </div>
      </div>
      <div className={styles.footerForm}>
        <input
          suppressHydrationWarning
          placeholder="Leave your contact"
          value={contact}
          onChange={(e) => {
            setIsUserInput(true);
            setContact(e.target.value);
          }}
          style={{
            color: contact.length ? colors.textPrimary : colors.textSecondary,
          }}
        />
        <textarea
          suppressHydrationWarning
          value={comment}
          placeholder={'Any additional comments (Optional)'}
          onChange={(e) => setComment(e.target.value)}
          style={{
            color: contact.length ? colors.textPrimary : colors.textSecondary,
          }}
        />
        <ButtonBase
          isDisabled={isDisabled && isUserInput}
          className={styles.submit}
          text={sending ? 'Sending...' : 'Submit'}
          onClick={handleSubmit}
        />
      </div>
    </div>
  );
};

const TabletFooterLinks = () => {
  return (
    <div className={styles.tabletLinks}>
      <div className={styles.tabletLinks__horisontalLine} />
      <div className={styles.tabletLinks__links}>
        {socials.map(
          (
            {
              href,
              icon,
              hoverIconColor,
              hoverBackgroundColor,
              hoverIcon,
              text,
            },
            i,
          ) => (
            <SocialLink
              key={i}
              href={href}
              icon={icon}
              hoverIcon={hoverIcon}
              hoverIconColor={hoverIconColor}
              hoverBackgroundColor={hoverBackgroundColor}
              text={text}
            />
          ),
        )}
      </div>
      <div className={styles.tabletLinks__horisontalLine} />
      <div className={styles.tabletLinks__ddLinksContainer}>
        {footerSections.map((e, i) => (
          <div className={styles.tabletLinks__ddLinks} key={i}>
            <Typography text={e.title} />
            <ul className={styles.links}>
              {e.links.map(({ label, href }) => (
                <li
                  key={label}
                  className={styles.linkItem}
                  onClick={() => {
                    if (isNeedLogin.includes(href)) {
                      customEvent.emit(CUSTOM_EVENT_KEYS.LOGIN);
                    }
                  }}
                >
                  <Link
                    href={href}
                    target={href.startsWith('/') ? undefined : '_blank'}
                    rel={`noopener noreferrer ${href.startsWith('/') ? '' : 'nofollow'}`}
                  >
                    <Typography>{label}</Typography>
                    {label === 'Analytics at' ? (
                      <div>
                        <Dune />
                        <DuneActive />
                      </div>
                    ) : null}
                  </Link>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
      <ResponsibleGamingBanner variant="footer" />
      <div className={styles.tabletLinks__horisontalLine} />
    </div>
  );
};

const FooterLeftOtherPage = memo(() => {
  return (
    <div className={styles.footerLeft}>
      <div className={styles.top}>
        <Image
          alt={'Duel Duck'}
          className={`${styles.image} ${styles.hideOnTablet}`}
          src={bigDuckSvg}
          data-duck={'duck'}
          width={150}
          height={150}
        />
        <div className={styles.info}>
          <div className={styles.nameAndPosition}>
            <span>
              <Typography
                className={styles.name}
                customStyles={{ fontSize: '32px' }}
              >
                DUEL{' '}
              </Typography>
              <Typography
                className={styles.name}
                customStyles={{ fontSize: '32px' }}
                color={colors.primary}
              >
                DUCK
              </Typography>
            </span>
          </div>
          <div className={`${styles.bio} ${styles.hideOnTablet}`}>
            {socials.map(
              (
                { href, icon, hoverIconColor, hoverBackgroundColor, hoverIcon },
                i,
              ) => (
                <SocialLink
                  key={i}
                  href={href}
                  icon={icon}
                  hoverIcon={hoverIcon}
                  hoverIconColor={hoverIconColor}
                  hoverBackgroundColor={hoverBackgroundColor}
                />
              ),
            )}
          </div>
          <p className={styles.legalCompanyName}>DUELDUCK CR LIMITADA</p>
        </div>
      </div>
    </div>
  );
});

export const NewFooter = memo(() => {
  const pathname = usePathname();
  const { isSmallTablet } = useMediaQuery();

  return (
    <footer
      className={styles.wrapper}
      style={{
        display: pageHideFooter.includes(pathname || '') ? 'none' : undefined,
      }}
      data-create={pathname === '/create-duel'}
      data-browse={pathname === '/duels'}
      data-is-home={pathname === '/'}
    >
      <div className={styles.footerContainer}>
        <div className={styles.footerContent}>
          <div
            className={cx(
              styles.footerLeftWrapper,
              pathname !== '/' && styles.hideOnPhone,
            )}
          >
            {pathname === '/' ? (
              <FooterLeftMainPage />
            ) : (
              <FooterLeftOtherPage />
            )}
          </div>
          {/* Tablet layout - shown only on tablet (not phone) */}
          <div className={styles.showOnlyOnTabletNotPhone}>
            <TabletFooterLinks />
          </div>
          {/* Desktop & Phone layout */}
          <div
            className={`${styles.footerRight} ${styles.hideOnTabletNotPhone}`}
          >
            <div
              className={`${styles.nameAndPosition} ${styles.phoneNameContainer}`}
            >
              <span style={{ margin: '0 auto' }}>
                <Typography
                  className={styles.name}
                  customStyles={{ fontSize: '32px' }}
                >
                  DUEL{' '}
                </Typography>
                <Typography
                  className={styles.name}
                  customStyles={{ fontSize: '32px' }}
                  color={colors.primary}
                >
                  DUCK
                </Typography>
              </span>
            </div>
            {footerSections.map(({ title, links }, sectionIndex) => (
              <div
                key={title}
                className={styles.section}
                data-last-section={sectionIndex === footerSections.length - 1}
                data-is-home={pathname === '/'}
              >
                <Typography className={styles.sectionTitle}>{title}</Typography>
                <ul className={styles.links}>
                  {links.map(({ label, href }) => (
                    <li
                      key={label}
                      className={styles.linkItem}
                      onClick={() => {
                        if (isNeedLogin.includes(href)) {
                          customEvent.emit(CUSTOM_EVENT_KEYS.LOGIN);
                        }
                      }}
                    >
                      <Link
                        href={href}
                        target={href.startsWith('/') ? undefined : '_blank'}
                        rel={`noopener noreferrer ${href.startsWith('/') ? '' : 'nofollow'}`}
                      >
                        <Typography>{label}</Typography>
                        {label === 'Analytics at' ? (
                          <div>
                            <Dune />
                            <DuneActive />
                          </div>
                        ) : null}
                      </Link>
                    </li>
                  ))}
                </ul>
              </div>
            ))}

            <ResponsibleGamingBanner variant="footer" />

            <div
              className={`${styles.bio} ${styles.phoneBioContainer}`}
              data-is-home={pathname === '/'}
            >
              {socials.map(
                (
                  {
                    href,
                    icon,
                    hoverIconColor,
                    hoverBackgroundColor,
                    hoverIcon,
                    text,
                  },
                  i,
                ) => (
                  <SocialLink
                    key={i}
                    href={href}
                    icon={icon}
                    hoverIcon={hoverIcon}
                    hoverIconColor={hoverIconColor}
                    hoverBackgroundColor={hoverBackgroundColor}
                    ariaLabelText={text}
                  />
                ),
              )}
            </div>
            {isSmallTablet && (
              <div className={styles.copyrightsMobile}>
                <Typography text={'2026 © DUELDUCK CR LIMITADA'} />{' '}
                <Link
                  href="https://overdone.it/"
                  target="_blank"
                  rel="nofollow noopener noreferrer"
                >
                  Developed by <OverdoneMonochrome width={20} height={20} />
                </Link>
              </div>
            )}
          </div>
        </div>
        {!isSmallTablet && <Copyright />}
      </div>
    </footer>
  );
});
