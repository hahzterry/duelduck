import { ReactNode } from 'react';

export interface LegalDoc {
  title: string;
  lastUpdated: Date;
  content: ReactNode;
  renderTitle?: boolean;
  renderLastUpdated?: boolean;
}
