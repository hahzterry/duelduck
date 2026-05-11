import type { Duel } from '~types/duel';
import type { Tournament } from '~types/tournament';

type BuildMainPageJsonLdArgs = {
  baseUrl: string;
  tournaments: Tournament[];
  duels: Duel[];
};

function buildTournamentSchema(tournament: Tournament, baseUrl: string) {
  const tournamentUrl = tournament.slug
    ? `${baseUrl}/tournament/${tournament.slug}/duels`
    : `${baseUrl}/tournament/${tournament.id}/duels`;

  return {
    '@type': 'SportsEvent',
    name: tournament.name,
    description: tournament.description,
    url: tournamentUrl,
    image: tournament.image_url,
    startDate: tournament.start_date,
    endDate: tournament.finish_date,
    eventStatus: 'https://schema.org/EventScheduled',
    eventAttendanceMode: 'https://schema.org/OnlineEventAttendanceMode',
    organizer: {
      '@id': `${baseUrl}#organization`,
    },
    offers: {
      '@type': 'Offer',
      price: '0',
      priceCurrency: 'USD',
      availability: 'https://schema.org/InStock',
      url: tournamentUrl,
    },
    additionalProperty: [
      {
        '@type': 'PropertyValue',
        name: 'Reward Pool USDC',
        value: tournament.reward_pool_usdc,
        unitText: 'USDC',
      },
      {
        '@type': 'PropertyValue',
        name: 'Players Count',
        value: tournament.players_count,
      },
    ],
  };
}

function buildDuelSchema(duel: Duel, baseUrl: string) {
  const duelUrl = duel.slug
    ? `${baseUrl}/share-duel/${duel.slug}`
    : `${baseUrl}/share-duel/${duel.id}`;

  return {
    '@type': 'Question',
    name: duel.question,
    text: duel.question,
    url: duelUrl,
    image: duel.image_url,
    dateCreated: duel.created_at,
    author: {
      '@type': 'Person',
      name: duel.username || 'Anonymous',
    },
    answerCount: duel.players_count || 0,
    additionalProperty: [
      {
        '@type': 'PropertyValue',
        name: 'Topic',
        value: duel.topic,
      },
      {
        '@type': 'PropertyValue',
        name: 'Deadline',
        value: duel.deadline,
      },
      {
        '@type': 'PropertyValue',
        name: 'Price',
        value: duel.duel_price,
      },
      {
        '@type': 'PropertyValue',
        name: 'Yes Votes',
        value: duel.yes_count || 0,
      },
      {
        '@type': 'PropertyValue',
        name: 'No Votes',
        value: duel.no_count || 0,
      },
    ],
  };
}

export function buildMainPageJsonLd({
  baseUrl,
  tournaments,
  duels,
}: BuildMainPageJsonLdArgs) {
  const schemas: object[] = [];

  if (tournaments.length > 0) {
    schemas.push({
      '@context': 'https://schema.org',
      '@type': 'ItemList',
      name: 'Active Tournaments on Duel Duck',
      description:
        'List of active prediction tournaments on Duel Duck platform where users can compete and win rewards.',
      itemListOrder: 'Descending',
      numberOfItems: tournaments.length,
      itemListElement: tournaments.map((tournament, index) => ({
        '@type': 'ListItem',
        position: index + 1,
        item: buildTournamentSchema(tournament, baseUrl),
      })),
    });
  }

  if (duels.length > 0) {
    schemas.push({
      '@context': 'https://schema.org',
      '@type': 'ItemList',
      name: 'Active Prediction Duels on Duel Duck',
      description:
        'List of active prediction questions on Duel Duck where users can vote Yes or No on outcomes in crypto, sports, and gaming.',
      itemListOrder: 'Descending',
      numberOfItems: duels.length,
      itemListElement: duels.slice(0, 10).map((duel, index) => ({
        '@type': 'ListItem',
        position: index + 1,
        item: buildDuelSchema(duel, baseUrl),
      })),
    });
  }

  return schemas;
}
