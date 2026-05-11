export interface Post {
  id: string;
  project_id: string;
  image: string;
  project: Project;
  title: string;
  slug: string;
  author_id: string;
  author: Author;
  category?: Category;
  category_id?: string;
  description: string;
  status: string;

  metadata: Metadata | undefined;

  created_at: string;
  updated_at: string;
  views?: number;
}

export interface Project {
  id: string;
  name: string;
  path: string;
  created_at: string;
  updated_at: string;
}

export interface Author {
  id: string;
  email: string;
  name: string;
  role: string;
  position: string;
  about: string;

  avatar: string | undefined;
  background: string | null;

  x: string | null;
  discord: string | null;
  telegram: string | null;
  instagram: string | null;
  linkedin: string | null;
  youtube: string | null;
  github: string | null;
  website: string | null;

  projects: any;
  created_at: string;
  updated_at: string;
}

export interface Category {
  id: string;
  project_id: string;
  name: string;
  post_count: number;
  created_at: string;
  updated_at: string;
}

export type Categories = Category[];

export interface FaqItem {
  id?: string | number;
  question: string;
  answer: string;
}

export interface KeyTakeawayItem {
  id?: string;
  content?: string;
  text?: string;
}

export interface KeyTakeawaysObject {
  title?: string;
  items: KeyTakeawayItem[];
}

export interface Metadata {
  html: string;
  json: Json;
  faq?: FaqItem[];
  keywords?: string[];
  words?: number;
  characters?: number;
  key_takeaways?: KeyTakeawaysObject | KeyTakeawayItem[];
}

export interface Json {
  type: string;
  content: Content[];
}

export interface Content {
  type: string;
  attrs: Attrs;
  content: Content2[];
}

export interface Attrs {
  id: string;
  textAlign: any;
}

export interface Content2 {
  text: string;
  type: string;
}
