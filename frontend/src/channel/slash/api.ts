import type { Card, Cardholder, ChannelAPI, ListResponse, Transaction } from "@/channel/types";
import { authorizationPayload, request } from "@/channel/shared";

const baseURL = "/slash/ui";

export interface Account {
  id: string;
  wallet_id: string;
  name: string;
  balance: string;
  created_at: string;
}

export const accountApi = {
  async listAll(): Promise<Account[]> {
    const accounts: Account[] = [];
    let page = 1;
    while (true) {
      const result = await this.list(page, 100);
      accounts.push(...result.data);
      if (accounts.length >= result.total_items || result.data.length === 0) {
        return accounts;
      }
      page += 1;
    }
  },
  async list(pageNumber = 1, pageSize = 20) {
    return (
      await request.get<ListResponse<Account>>(`${baseURL}/accounts`, {
        params: {
          page_number: pageNumber,
          page_size: pageSize,
        },
      })
    ).data;
  },
  async create(payload: Pick<Account, "name">) {
    return (await request.post<Account>(`${baseURL}/accounts`, payload)).data;
  },
  async update(id: string, payload: Pick<Account, "name">) {
    return (await request.put<Account>(`${baseURL}/accounts/${id}`, payload)).data;
  },
};

export interface Webhook {
  id: string;
  account_id: string;
  account_name: string;
  event: WebhookEvent;
  target_url: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export type WebhookEvent = string;

export const webhookApi = {
  async events() {
    return (await request.get<WebhookEvent[]>(`${baseURL}/webhooks/events`)).data;
  },
  async list(accountID?: string) {
    return (
      await request.get<Webhook[]>(`${baseURL}/webhooks`, {
        params: {
          account_id: accountID || undefined,
        },
      })
    ).data;
  },
  async create(payload: Omit<Webhook, "id" | "created_at" | "updated_at" | "account_name">) {
    return (await request.post<Webhook>(`${baseURL}/webhooks`, payload)).data;
  },
  async update(id: string, payload: Pick<Webhook, "target_url" | "enabled">) {
    return (await request.put<Webhook>(`${baseURL}/webhooks/${id}`, payload)).data;
  },
  async remove(id: string) {
    await request.delete(`${baseURL}/webhooks/${id}`);
  },
};

export interface CardProduct {
  id: string;
  prefix: string;
  is_default: boolean;
}

export interface VirtualAccount {
  id: string;
  name: string;
  currency: string;
  funding_source: string;
  balance: string;
  spend: string;
  created_at: string;
}

export const managementApi = {
  async cardProducts() {
    return (await request.get<{ items: CardProduct[] }>(`${baseURL}/card-products`)).data.items;
  },
  async virtualAccounts() {
    return (await request.get<VirtualAccount[]>(`${baseURL}/virtual-accounts`)).data;
  },
};

export async function applyTransactionAmount(
  id: string,
  action: "clear" | "reverse" | "refund",
  amount: number,
) {
  return (
    await request.post<Transaction>(`${baseURL}/transactions/${id}/${action}`, {
      amount,
    })
  ).data;
}

export const refundApi = {
  async simulate(payload: {
    authorization_id?: string;
    card_id?: string;
    amount: number;
    currency?: string;
    merchant_name: string;
    merchant_category_code: string;
    merchant_country: string;
    merchant_city: string;
  }) {
    return (await request.post<Transaction>(`${baseURL}/simulate/refunds`, payload)).data;
  },
};

export const api: ChannelAPI = {
  async listCardholders(page) {
    return (
      await request.get<ListResponse<Cardholder>>(`${baseURL}/cardholders`, {
        params: page,
      })
    ).data;
  },
  async listCards(page) {
    return (
      await request.get<ListResponse<Card>>(`${baseURL}/cards`, {
        params: page,
      })
    ).data;
  },
  async updateCardStatus(payload) {
    return (
      await request.put<Card>(`${baseURL}/cards/${payload.id}/status`, {
        account_id: payload.account_id,
        card_status: payload.card_status,
      })
    ).data;
  },
  async listTransactions(page) {
    return (
      await request.get<ListResponse<Transaction>>(`${baseURL}/transactions`, {
        params: page,
      })
    ).data;
  },
  async simulateAuthorization(payload) {
    await request.post(`${baseURL}/simulate/authorizations`, authorizationPayload(payload));
  },
  async applyTransactionStep(id, action, amount) {
    await request.post(`${baseURL}/transactions/${id}/${action}`, {
      amount,
    });
  },
};

export interface Wallet {
  id: string;
  account_id: string;
  account_name: string;
  kind: "account" | "virtual_account" | "card";
  currency: string;
  amount: string;
}

export const fundsApi = {
  async list(accountID?: string) {
    return (
      await request.get<Wallet[]>(baseURL + "/funds", {
        params: {
          account_id: accountID,
        },
      })
    ).data;
  },

  async adjustAccount(account: Account, amount: number) {
    await request.post(baseURL + "/funds/transfer", {
      account_id: account.id,
      source_id: amount < 0 ? account.wallet_id : "",
      target_id: amount > 0 ? account.wallet_id : "",
      amount: String(Math.abs(amount)),
    });
  },

  async transferResource(input: {
    accountID: string;
    walletID: string;
    amount: number;
    withdraw: boolean;
  }) {
    const wallets = await this.list(input.accountID);
    const accountWallet = wallets.find((wallet) => wallet.kind === "account");
    if (!accountWallet) {
      throw new Error("关联账户钱包不存在");
    }
    await request.post(baseURL + "/funds/transfer", {
      account_id: input.accountID,
      source_id: input.withdraw ? input.walletID : accountWallet.id,
      target_id: input.withdraw ? accountWallet.id : input.walletID,
      amount: String(input.amount),
    });
  },
};

export interface WebhookRecord {
  id: string;
  account_id: string;
  account_name: string;
  event: string;
  target_url: string;
  source_id: string;
  payload: string;
  request_headers: string;
  response_body: string;
  response_headers: string;
  status_code: number;
  status: string;
  attempt_count: number;
  delivered_at: string | null;
  error_message: string;
  created_at: string;
}

export interface WebhookRecordListRequest {
  page_number: number;
  page_size: number;
  account_id?: string;
  event?: string;
  status?: string;
  created_from?: string;
  created_to?: string;
}

export const webhookRecordApi = {
  async list(query: WebhookRecordListRequest) {
    return (
      await request.get<ListResponse<WebhookRecord>>(`${baseURL}/webhook-records`, {
        params: query,
      })
    ).data;
  },
  async replay(record: Pick<WebhookRecord, "id" | "account_id">) {
    return (
      await request.post<WebhookRecord>(`${baseURL}/webhook-records/${record.id}/replay`, {
        account_id: record.account_id,
      })
    ).data;
  },
};
