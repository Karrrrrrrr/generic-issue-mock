import type { CardStatus, TransactionStatus, WalletTransferKind } from "@/channel/enums";
import type { AuthorizationTransaction } from "@/channel/shared/contracts";
export interface Page<Item> {
  items: Item[];
  total: number;
}
export interface Account {
  wallet_id: number;
  id: number;
  name: string;
  balance: number;
  currency: string;
  created_at: string;
}
export interface Owned {
  id: number;
  account_id: number;
  account_name: string;
  created_at: string;
}
export interface VirtualAccount extends Owned {
  wallet_id: number;
  name: string;
  balance: number;
  currency: string;
}
export interface Card extends Owned {
  wallet_id: number;
  cvv: string;
  expires_at: string;
  card_type: "single" | "share" | "virtual_account_single";
  virtual_account_id: number;
  card_number: string;
  card_bin: string;
  status: CardStatus;
  balance: number;
  reserved: number;
  currency: string;
}
export interface Authorization extends Owned {
  card_id: number;
  amount: number;
  remaining: number;
  settled: number;
  reversed: number;
  refunded: number;
  currency: string;
  merchant_name: string;
  status: TransactionStatus;
}
export interface AuthorizationDetail extends Authorization {
  card_number: string;
  authorization_code: string;
  merchant_country: string;
  merchant_mcc: string;
  raw_payload: unknown;
  transactions: AuthorizationTransaction[];
}
export interface Transfer extends Owned {
  request_id: string;
  kind: WalletTransferKind;
  amount: string;
  currency: string;
  source_wallet_id: number;
  target_wallet_id: number;
}
export interface Product {
  id: number;
  prefix: string;
}
async function request<Item>(path: string, options?: RequestInit): Promise<Item> {
  const response = await fetch(`/pingpong/ui/${path}`, options);
  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.message || "请求失败");
  }
  return data as Item;
}
export const api = {
  get<Item>(path: string, params: Record<string, string | number | undefined> = {}) {
    const query = new URLSearchParams();
    for (const [key, value] of Object.entries(params)) {
      if (value !== undefined) {
        query.set(key, String(value));
      }
    }
    return request<Item>(`${path}?${query}`);
  },
  post<Item>(path: string, body: unknown) {
    return request<Item>(path, {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify(body),
    });
  },
  put<Item>(path: string, body: unknown) {
    return request<Item>(path, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify(body),
    });
  },
};
export async function accountOptions() {
  const accounts: Account[] = [];
  for (let page = 1;; page++) {
    const result = await api.get<Page<Account>>("accounts", {
      page_number: page,
      page_size: 100,
    });
    accounts.push(...result.items);
    if (accounts.length >= result.total || result.items.length === 0) {
      break;
    }
  }
  return accounts.map(account => ({
    label: `${account.name} · ${account.id}`,
    value: account.id,
  }));
}
export function requestID() {
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, value => value.toString(16).padStart(2, "0")).join("");
}
