import type { Account as UIAccount, CardProduct } from "@/channel/shared/api";
import type { AuthorizationAPI, CardFundingRequest, VirtualAccountAPI } from "@/channel/shared/contracts";
import type { AuthorizationSimulationRequest } from "@/channel/shared/simulation";
import type { Card as UICard, ChannelAPI, ListResponse } from "@/channel/types";
import {
  api,
  type Account,
  type Authorization,
  type Card,
  type Page,
  type Product,
  type VirtualAccount,
} from "./api";

async function listAll<Item>(resource: string, filters: Record<string, string | undefined> = {}) {
  const items: Item[] = [];
  for (let page = 1; ; page++) {
    const result = await api.get<Page<Item>>(resource, {
      ...filters,
      page_no: page,
      page_size: 100,
    });
    items.push(...result.items);
    if (items.length >= result.total || result.items.length === 0) {
      return items;
    }
  }
}

function toAccount(account: Account): UIAccount {
  return {
    id: account.id,
    name: account.name,
    balance: String(account.balance),
    currency: account.currency,
    created_at: account.created_at,
  };
}

function toCard(card: Card): UICard {
  return {
    id: card.id,
    account_id: card.account_id,
    account_name: card.account_name,
    virtual_account_id: card.virtual_account_id,
    card_number: card.card_number,
    card_bin: card.card_bin,
    card_currency: card.currency,
    card_status: card.status,
    balance: String(card.balance),
    reserved: String(card.reserved),
    created_at: card.created_at,
  };
}

export const accountApi = {
  async listAll(): Promise<UIAccount[]> {
    return (await listAll<Account>("accounts")).map(toAccount);
  },
  async list(pageNumber = 1, pageSize = 20): Promise<ListResponse<UIAccount>> {
    const result = await api.get<Page<Account>>("accounts", {
      page_no: pageNumber,
      page_size: pageSize,
    });
    return { data: result.items.map(toAccount), total_items: result.total };
  },
  async create(request: { name: string }): Promise<UIAccount> {
    return toAccount(await api.post<Account>("accounts", request));
  },
};

export async function adjustAccountBalance(account: UIAccount, amount: number) {
  await api.post(`accounts/${account.id}/balance`, { amount });
}

export const cardApi: Pick<ChannelAPI, "listCards" | "updateCardStatus"> = {
  async listCards(query) {
    const result = await api.get<Page<Card>>("cards", {
      page_no: query?.page_number ?? 1,
      page_size: query?.page_size ?? 20,
      account_id: query?.account_id,
      card_id: query?.id,
      status: query?.card_status,
    });
    return { data: result.items.map(toCard), total_items: result.total };
  },
  async updateCardStatus(request) {
    await api.put(`cards/${request.id}/status`, {
      account_id: request.account_id,
      status: request.card_status,
    });
  },
};

export async function fundCard(request: CardFundingRequest) {
  if (!request.requestID) {
    throw new Error("资金操作缺少请求 ID");
  }
  await api.post(`cards/${request.card.id}/fund`, {
    account_id: request.card.account_id,
    action: request.withdraw ? "withdraw" : "top_up",
    amount: request.amount,
    request_id: request.requestID,
  });
}

export async function simulateAuthorization(request: AuthorizationSimulationRequest & { requestID: string }) {
  await api.post("simulate/authorizations", {
    card_id: request.cardID,
    amount: request.amount,
    currency: request.currency,
    merchant_name: request.merchantName,
    request_id: request.requestID,
  });
}

export const authorizationApi: AuthorizationAPI = {
  async list(filters) {
    const items = await listAll<Authorization>("authorizations", {
      account_id: filters.account_id,
      card_id: filters.card_id,
    });
    return items.map((item) => ({
      id: item.id,
      account_id: item.account_id,
      account_name: item.account_name,
      card_id: item.card_id,
      amount: String(item.amount),
      remaining: String(item.remaining),
      currency: item.currency,
      merchant_name: item.merchant_name,
      status: item.status,
      notification_status: item.notification_status,
      created_at: item.created_at,
    }));
  },
  async stage(request) {
    if (!request.requestID) {
      throw new Error("模拟交易缺少请求 ID");
    }
    await api.post(`authorizations/${request.authorization.id}/stages`, {
      stage: request.stage,
      amount: request.amount,
      request_id: request.requestID,
    });
  },
};

export const virtualAccountApi: VirtualAccountAPI = {
  async list() {
    return (await listAll<VirtualAccount>("virtual-accounts")).map((account) => ({
      id: account.id,
      account_id: account.account_id,
      account_name: account.account_name,
      name: account.name,
      balance: String(account.balance),
      currency: account.currency,
    }));
  },
  async create(request) {
    if (request.currency !== "USD") {
      throw new Error("当前仅支持 USD");
    }
    await api.post("virtual-accounts", {
      account_id: request.account_id,
      name: request.name,
    });
  },
  async topUp(request) {
    if (!request.requestID) {
      throw new Error("资金操作缺少请求 ID");
    }
    await api.post(`virtual-accounts/${request.account.id}/fund`, {
      account_id: request.account.account_id,
      amount: request.amount,
      request_id: request.requestID,
    });
  },
};

export async function loadProducts(): Promise<CardProduct[]> {
  return (await api.get<Page<Product>>("products")).items;
}
