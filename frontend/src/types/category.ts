export type CategoryStatus = 'active' | 'archived' | 'draft';

export interface CategoryFaq {
  question: string;
  answer: string;
}

export interface CategoryStats {
  tvl_usdc: number;
  participants_count: number;
  active_markets_count: number;
  average_prize_pool_usdc: number;
}

/**
 * Lightweight category record used by the seed data in categoryApi.ts.
 * Will be replaced by CategoryApiResponse once the backend is live.
 */
export interface CategoryData {
  id: string;
  slug: string;
  parent_slug: string | null;
  title: string;
  description: string;
  og_image: string | null;
  keywords: string[];
  status: CategoryStatus;
  noindex: boolean;
  created_at: string;
  updated_at: string;
}

export interface SubcategoryInfo {
  id: string;
  slug: string;
  title: string;
  description: string;
  status: CategoryStatus;
}

/**
 * Unified category data returned by the backend API.
 * Combines all data the frontend needs for a category page.
 *
 * Endpoint: GET /api/categories/:slug
 */
export interface CategoryApiResponse {
  id: string;
  slug: string;
  parent_slug: string | null;

  // Display
  title: string;
  description: string;
  hero_title: string;
  hero_description: string;

  // SEO
  seo_title: string;
  seo_description: string;
  og_image: string | null;
  keywords: string[];
  noindex: boolean;

  // Lifecycle
  status: CategoryStatus;

  // Stats — null when no data is available yet
  stats: CategoryStats | null;

  // Content
  faqs: CategoryFaq[];

  // Nested subcategories (only for root categories)
  subcategories: SubcategoryInfo[];

  // Timestamps
  created_at: string;
  updated_at: string;
}
