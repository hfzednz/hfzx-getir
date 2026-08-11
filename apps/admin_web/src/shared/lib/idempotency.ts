import { v4 as uuidv4 } from "uuid";

/** Generate an Idempotency-Key (UUID v4). */
export function createIdempotencyKey(): string {
  return uuidv4();
}
