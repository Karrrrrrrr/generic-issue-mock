import type { Account as UIAccount, CardProduct } from "@/channel/shared/api";
import type { Authorization as UIAuthorization, AuthorizationAPI, CardFundingRequest, VirtualAccountAPI } from "@/channel/shared/contracts";
import type { AuthorizationSimulationRequest, RefundSimulationRequest } from "@/channel/shared/simulation";
import type { Card as UICard, ChannelAPI, ListResponse, Transaction } from "@/channel/types";
import {
  api,
  requestID,
  type AuthorizationDetail,
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
    wallet_id: account.wallet_id,
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
    wallet_id: card.wallet_id,
    cvv: card.cvv,
    expires_at: card.expires_at,
    funding_source: card.card_type === "virtual_account_single" ? "虚拟账户供资独立卡" : card.card_type === "share" ? "虚拟账户共享资金" : "卡资金",
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
  async update(id: number, request: { name: string }): Promise<UIAccount> {
    return toAccount(await api.put<Account>(`accounts/${id}`, request));
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
      card_number: query?.card_number,
      created_from: query?.created_from,
      created_to: query?.created_to,
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
    merchant_category_code: request.merchantMCC,
    merchant_country: request.merchantCountry,
    request_id: request.requestID,
  });
}

function toAuthorization(item: Authorization): UIAuthorization {
  return {
    id: item.id,
    account_id: item.account_id,
    account_name: item.account_name,
    card_id: item.card_id,
    amount: String(item.amount),
    remaining: String(item.remaining),
    settled: String(item.settled),
    reversed: String(item.reversed),
    refunded: String(item.refunded),
    currency: item.currency,
    merchant_name: item.merchant_name,
    status: item.status,
    notification_status: item.notification_status,
    created_at: item.created_at,
  };
}

export const authorizationApi: AuthorizationAPI = {
  async list(filters) {
    return (await listAll<Authorization>("authorizations", filters)).map(toAuthorization);
  },
  async detail(authorization) {
    const item = await api.get<AuthorizationDetail>(`authorizations/${authorization.id}`, {
      account_id: authorization.account_id,
    });
    return {
      ...toAuthorization(item),
      card_number: item.card_number,
      authorization_code: item.authorization_code,
      merchant_country: item.merchant_country,
      merchant_mcc: item.merchant_mcc,
      raw_payload: item.raw_payload,
      transactions: item.transactions,
    };
  },
  async stage(request) {
    if (!request.requestID) {
      throw new Error("模拟交易缺少请求 ID");
    }
    await api.post(`authorizations/${request.authorization.id}/stages`, {
      stage: request.stage === "reverse" ? "void" : request.stage,
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
    wallet_id: account.wallet_id,
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

export async function loadSimulationCards(): Promise<UICard[]> {
  return (await listAll<Card>("cards")).map(toCard);
}

export async function simulateRefund(request: RefundSimulationRequest & { requestID: string }) {
  await api.post("simulate/refunds", {
    authorization_id: request.authorization_id,
    card_id: request.card_id,
    amount: request.amount,
    currency: request.currency,
    merchant_name: request.merchant_name,
    merchant_category_code: request.merchant_category_code,
    merchant_country: request.merchant_country,
    request_id: request.requestID,
  });
}

const pendingTransactionRequests = new Map<string, string>();

export const transactionApi: Pick<ChannelAPI, "listTransactions" | "applyTransactionStep"> = {
  async listTransactions(query) {
    const result = await api.get<Page<Transaction>>("transactions", {
      page_no: query?.page_number ?? 1,
      page_size: query?.page_size ?? 20,
      account_id: query?.account_id,
      id: query?.id,
      card_id: query?.card_id,
      authorization_id: query?.authorization_id,
      transaction_type: query?.transaction_type,
      status: query?.status,
      created_from: query?.created_from,
      created_to: query?.created_to,
    });
    return { data: result.items, total_items: result.total };
  },
  async applyTransactionStep(id, action, amount) {
    const key = JSON.stringify([id, action, amount]);
    const orderID = pendingTransactionRequests.get(key) ?? requestID();
    pendingTransactionRequests.set(key, orderID);
    await api.post(`transactions/${id}/stages`, {
      transaction_type: action === "reverse" ? "void" : action,
      amount,
      request_id: orderID,
    });
    pendingTransactionRequests.delete(key);
  },
};
