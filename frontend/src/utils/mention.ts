import type { MentionChallengeLeader } from '~types/leaders';

type BuildMentionboardJsonLdArgs = {
  baseUrl: string;
  pageUrl: string;
  leaders: MentionChallengeLeader[];
};

export function buildMentionboardJsonLd({
  baseUrl,
  pageUrl,
  leaders,
}: BuildMentionboardJsonLdArgs) {
  return [
    {
      '@context': 'https://schema.org',
      '@type': 'WebPage',
      '@id': pageUrl,
      url: pageUrl,
      name: 'DuelDuck Mention Leaderboard',
      description:
        'Compete on the DuelDuck Mention Leaderboard by sharing content about @DuelDuck on X. Earn guaranteed USDC rewards based on your engagement score.',
      breadcrumb: {
        '@type': 'BreadcrumbList',
        itemListElement: [
          {
            '@type': 'ListItem',
            position: 1,
            name: 'Home',
            item: baseUrl,
          },
          {
            '@type': 'ListItem',
            position: 2,
            name: 'Mention Leaderboard',
            item: pageUrl,
          },
        ],
      },
      mainEntity: {
        '@type': 'ItemList',
        name: 'DuelDuck Mention Leaderboard Rankings',
        description:
          'Ranked list of participants in the DuelDuck Mention Challenge, ordered by total score and USDC rewards earned.',
        itemListOrder: 'Descending',
        numberOfItems: leaders?.length || 0,
        itemListElement: (leaders ?? [])
          .slice()
          .sort((a, b) => a.rank - b.rank)
          .map((leader) => {
            const displayName = leader.x_username || leader.username;

            return {
              '@type': 'ListItem',
              position: leader.rank,
              item: {
                '@type': 'Person',
                name: displayName,
                identifier: leader.user_id,
                url: leader.x_username
                  ? `https://x.com/${leader.x_username}`
                  : undefined,
                sameAs: leader.x_username
                  ? `https://x.com/${leader.x_username}`
                  : undefined,
                description: `Rank ${leader.rank} with ${leader.total_score.toFixed(2)} total score, earning ${leader.usdc_reward.toFixed(2)} USDC`,
                additionalProperty: [
                  {
                    '@type': 'PropertyValue',
                    name: 'Total Score',
                    value: leader.total_score,
                  },
                  {
                    '@type': 'PropertyValue',
                    name: 'DD Score',
                    value: leader.dd_score,
                  },
                  {
                    '@type': 'PropertyValue',
                    name: 'USDC Reward',
                    value: leader.usdc_reward,
                    unitText: 'USDC',
                  },
                  {
                    '@type': 'PropertyValue',
                    name: 'Rank',
                    value: leader.rank,
                  },
                ],
              },
            };
          }),
      },
    },
  ];
}

export const getXProfileByUsername = (username: string) =>
  `https://x.com/${username}`;
