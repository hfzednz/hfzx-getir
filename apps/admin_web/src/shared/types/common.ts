export type Id = string;

export interface City {
  id: Id;
  name: string;
  code: string;
  timezone?: string;
  active?: boolean;
}

export interface Paginated<T> {
  items: T[];
  page: number;
  pageSize: number;
  total: number;
  hasMore: boolean;
}

export interface ApiErrorBody {
  error: {
    code: string;
    message: string;
    traceId: string;
  };
}
