import { usePathname } from 'next/navigation';

import { TextButton } from '~components/Buttons/TextButton';

import styles from './styles.module.scss';

const directLinks = [
  { text: 'Duels', href: '/duels' },
  { text: 'Tournaments', href: '/tournaments' },
  { text: 'API', href: '/api' },
  { text: 'Blog', href: '/blog' },
];

export const Links = () => {
  const pathname = usePathname();

  return (
    <nav className={styles.links}>
      {directLinks.map((item) => (
        <TextButton
          href={item.href}
          activeColor={'primary'}
          fontWeight={pathname === item.href ? 'sb' : 'r'}
          isActive={pathname === item.href}
          key={item.text}
          text={item.text}
        />
      ))}
      <TextButton
        href={'/blog'}
        activeColor={'primary'}
        fontWeight={pathname === '/blog' ? 'sb' : 'r'}
        isActive={pathname === '/blog'}
        text={'Learn'}
      />
    </nav>
  );
};
