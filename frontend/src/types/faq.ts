export enum MARKED_FAQ_TYPE {
  DISLIKE = -1,
  LIKE = 1,
  NOT_VOTED = 0,
}

export interface RepliedType {
  images: string[];
  image_url?: string;
  created_at: string;
  id: string;
  question_id: number;
  reply: string;
  username?: string;
}

export interface FaqTypeRes {
  marked: MARKED_FAQ_TYPE;
  is_answered: boolean;
  marks_value: number;
  answer_id: number;
  is_me: boolean;
  image_url?: string;
  id: number;
  question: string;
  username: string;
  answer: string;
  answer_created_at: string;
  question_created_at: string;
  question_images: string[] | null;
  answer_images: string[] | null;
  replies: RepliedType[];
}
