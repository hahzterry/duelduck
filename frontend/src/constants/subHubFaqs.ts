interface FaqItem {
  question: string;
  answer: string;
}

export const subHubFaqs: Record<string, FaqItem[]> = {
  bitcoin: [
    {
      question: 'What Bitcoin events can I predict on DuelDuck?',
      answer:
        'You can predict BTC price milestones ($100K, $200K, $300K), halving cycle effects, ETF inflow targets, mining difficulty changes, and adoption metrics.',
    },
    {
      question: 'Which tokens can I use to enter Bitcoin duels?',
      answer:
        'DuelDuck operates on Solana. You can enter duels using SOL, USDC, and other supported SPL tokens.',
    },
    {
      question: 'How is this different from leveraged BTC trading?',
      answer:
        'DuelDuck is a prediction market with fixed entry costs. There is no leverage, no margin calls, and no liquidation risk. You simply predict an outcome and share the pool if correct.',
    },
    {
      question: 'What happens if a BTC price target seems impossible?',
      answer:
        'Duels remain open until their deadline. If the target is not reached by the deadline, the market resolves in favor of the "No" side and the pool is distributed accordingly.',
    },
    {
      question: 'How are Bitcoin prediction outcomes verified?',
      answer:
        'Outcomes are resolved using on-chain price oracles, major exchange price feeds, and verifiable blockchain data specified in each duel\u2019s resolution criteria.',
    },
  ],

  ethereum: [
    {
      question: 'What Ethereum events can I predict?',
      answer:
        'ETH price milestones ($5K, $10K), protocol upgrades (Pectra, Verkle), staking APY changes, L2 TVL targets, and ETH/BTC ratio movements.',
    },
    {
      question: 'How are Ethereum predictions settled?',
      answer:
        'Via on-chain data (ETH price oracles, beacon chain metrics), Ethereum Foundation announcements, and L2Beat/DefiLlama data for L2 metrics.',
    },
    {
      question: 'Can I predict Ethereum upgrade timelines?',
      answer:
        'Yes. Duels on upgrade activation dates are popular. Resolution uses the Ethereum mainnet block at which the upgrade activates.',
    },
    {
      question: 'Is ETH staking APY predictable?',
      answer:
        'Users create duels on staking yield ranges. Resolution uses beacon chain consensus rewards data at the specified snapshot date.',
    },
    {
      question: 'How does this differ from trading ETH?',
      answer:
        'DuelDuck is a prediction market with fixed entry costs and no liquidation. You predict specific outcomes, not trade ETH directly.',
    },
  ],

  defi: [
    {
      question: 'What DeFi events can I predict?',
      answer:
        'Protocol TVL milestones, yield farming APY ranges, governance votes, liquidation cascades, and cross-chain bridge volumes.',
    },
    {
      question: 'How are DeFi predictions resolved?',
      answer:
        'Using on-chain data from DefiLlama, Dune Analytics, protocol governance portals, and verified smart contract state.',
    },
    {
      question: 'Can I predict rug pulls or exploits?',
      answer:
        'DuelDuck does not support speculative harm markets. Duels must have verifiable, constructive resolution criteria.',
    },
    {
      question: 'Which DeFi protocols are covered?',
      answer:
        'Any protocol with publicly verifiable on-chain data. Popular categories include Aave, Uniswap, Lido, MakerDAO, and emerging protocols.',
    },
    {
      question: 'Do I need DeFi experience to participate?',
      answer:
        'No. DuelDuck simplifies DeFi forecasting into yes/no duels. If you follow DeFi news, you can participate.',
    },
  ],

  etf: [
    {
      question: 'What crypto ETF events can I predict?',
      answer:
        'SEC approval/rejection of spot ETFs, ETF inflow milestones ($1B, $10B), new asset ETF filings (SOL, XRP), and institutional adoption timelines.',
    },
    {
      question: 'How are ETF predictions resolved?',
      answer:
        'Using SEC filings, official fund prospectuses, and verified AUM data from Bloomberg/Morningstar.',
    },
    {
      question: 'Can I predict which assets get ETFs next?',
      answer:
        "Yes. Popular duels include 'Will a spot SOL ETF be approved by 2027?' resolved via SEC public filings.",
    },
    {
      question: 'How do ETF predictions relate to crypto prices?',
      answer:
        'ETF approvals historically correlate with price movements. DuelDuck lets you predict the regulatory event itself, separate from price action.',
    },
  ],

  memecoin: [
    {
      question: 'What memecoin events can I predict?',
      answer:
        'Market cap milestones, exchange listings, social media viral moments, trading volume spikes, and community governance decisions.',
    },
    {
      question: 'How are memecoin predictions resolved?',
      answer:
        'Using CoinGecko/CoinMarketCap data for market caps, exchange announcements for listings, and on-chain metrics for volume.',
    },
    {
      question: 'Can I create duels about new memecoins?',
      answer:
        'Yes, as long as the token has verifiable on-chain data. The duel must reference a specific metric and data source for resolution.',
    },
    {
      question: 'Is predicting memecoins risky?',
      answer:
        'DuelDuck prediction markets have fixed entry costs and no liquidation risk. You risk only your entry fee, unlike leveraged memecoin trading.',
    },
  ],

  regulation: [
    {
      question: 'What crypto regulation events can I predict?',
      answer:
        'SEC enforcement actions, congressional crypto bills, MiCA implementation deadlines, stablecoin legislation, CBDC launches, and international crypto bans/approvals.',
    },
    {
      question: 'How are regulation predictions resolved?',
      answer:
        'Using official government publications: SEC filings, Federal Register entries, EU Official Journal, congressional vote records, and central bank announcements.',
    },
    {
      question:
        'Can I use regulation predictions to hedge my crypto portfolio?',
      answer:
        'Yes. If you hold crypto assets and worry about regulatory headwinds, you can take a position in a regulation duel to hedge that risk.',
    },
    {
      question: 'Are international crypto regulations covered?',
      answer:
        'Yes. Markets cover US (SEC, CFTC), EU (MiCA), UK (FCA), Asia (MAS, FSA), and emerging market crypto regulation.',
    },
  ],

  dogecoin: [
    {
      question: 'What is DuelDuck?',
      answer:
        'A decentralized, peer-to-peer prediction market platform on Solana. Every duel is a smart contract with clear terms, binary outcomes, and automatic settlement.',
    },
    {
      question: 'Why Dogecoin?',
      answer:
        'DOGE is highly sentiment-driven and community-led \u2014 ideal for live probability markets. Price moves, adoption campaigns, and social milestones all generate predictable event cycles.',
    },
    {
      question: 'How are odds formed?',
      answer:
        'By the live distribution of stakes between outcomes. Every new bet updates the implied probability instantly \u2014 a self-regulating view of sentiment with no external control.',
    },
    {
      question: 'Who confirms results?',
      answer:
        'The duel creator, manually or via API, using open public data: exchange price feeds, blockchain explorers, and official Dogecoin network updates.',
    },
    {
      question: 'Can anyone create markets?',
      answer:
        'Yes. Any verified Solana wallet can launch duels and earn creator fees.',
    },
    {
      question: 'Are payouts automatic?',
      answer:
        'Yes. Settlement is instant once results are confirmed. Winners receive proportional payouts on-chain.',
    },
  ],

  'us-2028': [
    {
      question: 'What US 2028 election events can I predict?',
      answer:
        'Party nominations, primary results, debate outcomes, swing state predictions, popular vote margins, Electoral College outcomes, and VP picks.',
    },
    {
      question: 'When do US 2028 election markets open?',
      answer:
        'Markets are already active for early predictions like party nominees and primary timing. More specific markets open as the race develops.',
    },
    {
      question: 'How are election predictions settled?',
      answer:
        'Using official certified election results from state election boards and the Electoral College certification.',
    },
    {
      question: 'Are prediction markets more accurate than polls?',
      answer:
        'Research shows prediction markets often outperform polls for binary outcomes because participants have financial incentives for accuracy.',
    },
    {
      question: 'Can I predict individual state results?',
      answer:
        'Yes. Swing state duels (Pennsylvania, Georgia, Arizona, Wisconsin, Michigan, Nevada) are among the most popular election markets.',
    },
  ],

  'uk-2029': [
    {
      question: 'What UK election events can I predict?',
      answer:
        'Overall winner, seat counts by party, specific constituency results, leadership challenges, coalition scenarios, and voter turnout milestones.',
    },
    {
      question: 'How are UK election predictions resolved?',
      answer:
        'Using official results declared by returning officers and certified by the Electoral Commission.',
    },
    {
      question: 'Can I predict a snap election before 2029?',
      answer:
        'Yes. DuelDuck supports duels on snap election timing. Resolution uses the official dissolution of Parliament.',
    },
    {
      question: 'Are Scottish and Welsh elections covered?',
      answer:
        'Devolved elections can be added as separate duels. The UK 2029 hub focuses on the Westminster general election.',
    },
  ],

  'france-2027': [
    {
      question: 'What France 2027 events can I predict?',
      answer:
        'First round qualifiers, second round winner, candidate polling milestones, party endorsements, and voter turnout thresholds.',
    },
    {
      question: 'How are French election predictions resolved?',
      answer:
        'Using official results from the Conseil constitutionnel and the Ministry of the Interior.',
    },
    {
      question: 'Can I predict who will run in the French election?',
      answer:
        'Yes. Pre-election duels cover candidate declarations, party primaries, and the official 500 parrainages threshold.',
    },
    {
      question: 'Is the content available in French?',
      answer:
        'DuelDuck is currently in English only. French localization is planned for a future release.',
    },
  ],

  'germany-2029': [
    {
      question: 'What Germany 2029 events can I predict?',
      answer:
        'Bundestag seat distribution, coalition formations (traffic light, grand coalition, Jamaica), Chancellor candidate selection, party vote share thresholds (5% hurdle), and state election outcomes.',
    },
    {
      question: 'How are German election predictions resolved?',
      answer:
        'Using official results from the Bundeswahlleiter (Federal Returning Officer) and certified Bundestag seat allocation.',
    },
    {
      question: 'Can I predict coalition negotiations?',
      answer:
        'Yes. Post-election coalition duels are popular. Resolution uses the official coalition agreement signing and Chancellor election in the Bundestag.',
    },
    {
      question: 'Are snap elections possible in Germany?',
      answer:
        'Yes. The Chancellor can call a vote of confidence leading to early elections. DuelDuck supports duels on snap election timing.',
    },
  ],

  fed: [
    {
      question: 'What Fed events can I predict?',
      answer:
        'FOMC rate decisions (hold, cut, hike), dot plot projections, QT tapering timing, balance sheet size milestones, and Fed Chair press conference signals.',
    },
    {
      question: 'How are Fed predictions resolved?',
      answer:
        'Using official FOMC statements, Federal Reserve press releases, and FRED economic data.',
    },
    {
      question: 'How often are new Fed markets available?',
      answer:
        'New markets align with the FOMC schedule (8 meetings per year) plus interim economic data releases (CPI, NFP, GDP).',
    },
    {
      question: 'Can I predict the terminal rate?',
      answer:
        'Yes. Terminal rate duels are resolved based on the peak federal funds rate in the current cycle, verified by FOMC minutes.',
    },
  ],

  stocks: [
    {
      question: 'What stock events can I predict?',
      answer:
        'Index milestones (S&P 6000, Nasdaq 20K), earnings beats/misses, IPO day performance, market corrections, and sector rotation.',
    },
    {
      question: 'How are stock predictions resolved?',
      answer:
        'Using official market closing data from NYSE/Nasdaq, SEC earnings filings (10-Q/10-K), and verified financial data providers.',
    },
    {
      question: 'Is this like options trading?',
      answer:
        'No. DuelDuck stock predictions have fixed entry costs, no leverage, no margin, and no liquidation. You predict binary outcomes.',
    },
    {
      question: 'Can I predict individual stock prices?',
      answer:
        "Yes. Duels can target specific stock price milestones (e.g., 'Will AAPL close above $250 by Q2 2026?').",
    },
  ],

  economy: [
    {
      question: 'What are DuelDuck Economy Markets?',
      answer:
        'A space where the economy turns into an open game of predictions. Real events become markets: central bank decisions, inflation data, recessions, gold prices, and stock moves.',
    },
    {
      question: 'What is the platform built on?',
      answer:
        'The entire ecosystem runs on the Solana blockchain. Every bet, outcome, and payout is instantly recorded in smart contracts. No intermediaries, no delays.',
    },
    {
      question: 'Which currencies are supported?',
      answer:
        'Bet in whatever you prefer \u2014 Bitcoin, Ethereum, Solana, DOGE, USDT, or even trending memecoins. Every transaction happens directly between users on-chain.',
    },
    {
      question: 'How are winners determined?',
      answer:
        'Once an event ends, the platform verifies the result using official data sources \u2014 economic reports, Fed statements, or market prices. The pool is automatically distributed among correct forecasters.',
    },
    {
      question: 'Is it safe?',
      answer:
        'Completely. All transactions are processed on-chain, and every movement of funds is visible in the public ledger. DuelDuck never holds your money \u2014 it stays locked in a smart contract until the result is confirmed.',
    },
    {
      question: 'Can I create my own markets?',
      answer:
        'Yes. DuelDuck is a community-driven platform. You can create your own event, set the conditions, invite others, and earn a share of the pool.',
    },
    {
      question: 'What makes DuelDuck different from other platforms?',
      answer:
        'It\u2019s not a bookmaker and not an exchange. The community forms the odds through its bets, and every shift is recorded on the blockchain. Each prediction becomes part of the collective view of the global economy.',
    },
  ],

  'us-politics': [
    {
      question: 'What US political events can I predict?',
      answer:
        'Bill passage (infrastructure, crypto regulation, healthcare), executive orders, Supreme Court decisions, cabinet confirmations, government shutdowns, and approval rating milestones.',
    },
    {
      question: 'How are US politics predictions resolved?',
      answer:
        'Using official sources: congress.gov vote records, White House executive orders, Supreme Court opinions, and Federal Register publications.',
    },
    {
      question: "What's the difference between Politics and Elections hubs?",
      answer:
        'The Politics hub covers ongoing governance (legislation, SCOTUS, policy). The Elections hub covers specific election races (US 2028, UK 2029).',
    },
    {
      question: 'Can I predict government shutdowns?',
      answer:
        'Yes. Government shutdown duels are resolved based on the official lapse in appropriations as reported by OMB.',
    },
    {
      question: 'Are state-level politics covered?',
      answer:
        'Yes. Users can create duels on state legislation, gubernatorial actions, and ballot measures with verifiable resolution criteria.',
    },
  ],
};
