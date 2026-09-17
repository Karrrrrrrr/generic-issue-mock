import type { Card, Cardholder, ChannelAPI, ListResponse, Transaction, } from "@/channel/types";
import { authorizationPayload, request } from "@/channel/shared";

const baseURL = "/slash/ui";

export interface Account {
  id: string;
  name: string;
  created_at: string;
}

export const accountApi = {
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
  event: string;
  target_url: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export const webhookApi = {
  async list() {
    return (await request.get<Webhook[]>(`${baseURL}/webhooks`)).data;
  },
  async create(payload: Omit<Webhook, "id" | "created_at" | "updated_at">) {
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

export async function applyTransactionAmount(id: string, action: "clear" | "reverse" | "refund", amount: number) {
  return (await request.post<Transaction>(`${baseURL}/transactions/${id}/${action}`, { amount })).data;
}

export const refundApi = {
  async simulate(payload: {
    card_id: string;
    amount: number;
    currency: string;
    merchant_name: string;
    merchant_category_code: string;
    merchant_country: string;
    merchant_city: string;
  }) {
    return (await request.post<Transaction>(`${baseURL}/simulate/refunds`, payload)).data;
  },
};

export const api: ChannelAPI = {
  async listCardholders() {
    return (
        await request.get<ListResponse<Cardholder>>(`${baseURL}/cardholders`)
    ).data;
  },
  async createCardholder(payload) {
    return (
        await request.post<Cardholder>(`${baseURL}/cardholders`, {
          first_name: payload.firstName,
          last_name: payload.lastName,
          email: payload.email,
          phone_number: payload.mobile,
        })
    ).data;
  },
  async listCards() {
    return (await request.get<ListResponse<Card>>(`${baseURL}/cards`)).data;
  },
  async createCard(cardholderID, currency) {
    return (
        await request.post<Card>(`${baseURL}/cards`, {
          cardholder_id: cardholderID,
          card_currency: currency,
        })
    ).data;
  },
  async updateCardStatus(id, status) {
    return (
        await request.put<Card>(`${baseURL}/cards/${id}/status`, {
          card_status: status,
        })
    ).data;
  },
  async listTransactions() {
    return (
        await request.get<ListResponse<Transaction>>(`${baseURL}/transactions`)
    ).data;
  },
  async simulateAuthorization(payload) {
    await request.post(
        `${baseURL}/simulate/authorizations`,
        authorizationPayload(payload),
    );
  },
  async applyTransactionStep(id, action, amount) {
    await request.post(`${baseURL}/transactions/${id}/${action}`, { amount });
  },
};
