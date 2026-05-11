export interface Options {
  pagination?: Pagination;
  order?: Order;
  filters?: Filter[];
  challenge_id?: string;
}

export interface Pagination {
  page_size?: number;
  page_num?: number;
}

export interface Order {
  order_by: string;
  order_type: string;
}

export interface Filter {
  column: string;
  operator: string;
  value: string;
  where_or: boolean;
}
