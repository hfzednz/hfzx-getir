export interface WebSession {
  accessToken: string;
  refreshToken?: string;
  principalId: string;
  roles: string[];
  phone?: string;
  expiresAt?: string;
}
