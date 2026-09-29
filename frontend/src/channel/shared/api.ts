import type { Card, Cardholder, ChannelAPI, ListResponse, PageRequest, Transaction } from "@/channel/types";
import { authorizationPayload, request } from "@/channel/shared";
import type { RefundSimulationRequest } from "@/channel/shared/simulation";
import type { WebhookDeliveryStatus } from "@/channel/enums";
import type {
  Authorization,
  AuthorizationAPI,
  CardFundingRequest,
  ManagedVirtualAccount,
  VirtualAccountAPI,
} from "./contracts";

export interface Account {
  id: number;
  wallet_id?: number;
  name: string;
  balance: string;
  created_at: string;
  currency: string;
}

export interface AccountListRequest {
  id?: number;
  name?: string;
}

export interface Webhook {
  id: number;
  account_id: number;
  account_name: string;
  event: WebhookEvent;
  target_url: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export type WebhookEvent = string;

export interface AuthorizationConfig {
  account_id: number;
  account_name: string;
  target_url: string;
  enabled: boolean;
  timeout_millis: number;
  updated_at: string;
}

export interface CardProduct {
  id: number;
  prefix: string;
}

export interface VirtualAccount {
  id: number;
  name: string;
  currency: string;
  balance: string;
  created_at: string;
}

export interface Wallet {
  id: number;
  account_id: number;
  account_name: string;
  kind: "account" | "virtual_account" | "card";
  currency: string;
  amount: string;
}

export interface WebhookRecord {
  id: number;
  account_id: number;
  account_name: string;
  event: string;
  target_url: string;
  source_id: string;
  payload: string;
  request_headers: string;
  response_body: string;
  response_headers: string;
  status_code: number;
  status: WebhookDeliveryStatus;
  attempt_count: number;
  delivered_at: string | null;
  error_message: string;
  created_at: string;
}

export interface WebhookRecordListRequest {
  page_number: number;
  page_size: number;
  account_id?: number;
  event?: string;
  status?: WebhookDeliveryStatus;
  created_from?: string;
  created_to?: string;
}

export interface WalletListRequest extends PageRequest {
  account_id?: number;
}

export interface WebhookListRequest extends PageRequest {
  account_id?: number;
  event?: string;
}

interface Page<Item> {
  items: Item[];
  total: number;
}

type AccountData = Omit<Account, "balance"> & { available: string };
type VirtualAccountData = Omit<ManagedVirtualAccount, "balance"> & { available: string };
type CardData = Omit<Card, "card_currency" | "card_status" | "balance" | "reserved" | "funding_source" | "virtual_account_id"> & {
  currency: Card["card_currency"];
  status: Card["card_status"];
  available: string;
  pending_out: string;
  virtual_account_id: number | null;
  card_type: "single" | "share" | "virtual_account_single";
};
type CardholderData = Omit<Cardholder, "phone_number"> & { mobile: string };
type TransactionData = Omit<Transaction, "transaction_type" | "transacted_at" | "merchant_category_code"> & {
  type: Transaction["transaction_type"];
  created_at: string;
  merchant_mcc: string;
};
type AuthorizationDetailData = Omit<Authorization, "transactions"> & {
  card: CardData;
  transactions: TransactionData[];
};
type WalletData = Omit<Wallet, "kind" | "amount"> & { type: Wallet["kind"]; available: string };
type WebhookRecordData = Omit<WebhookRecord, "payload" | "request_headers" | "response_headers"> & {
  payload: unknown;
  request_headers: unknown;
  response_headers: unknown;
};

function toAccount(item: AccountData): Account {
  return { ...item, balance: item.available };
}

function toCard(item: CardData): Card {
  return {
    ...item,
    virtual_account_id: item.virtual_account_id ?? undefined,
    card_currency: item.currency,
    card_status: item.status,
    balance: item.available,
    reserved: item.pending_out,
    funding_source: item.card_type === "virtual_account_single" ? "虚拟账户供资独立卡" : item.card_type === "share" ? "虚拟账户共享资金" : "卡资金",
  };
}

function toTransaction(item: TransactionData): Transaction {
  return {
    ...item,
    transaction_type: item.type,
    transacted_at: item.created_at,
    merchant_category_code: item.merchant_mcc,
  };
}

function toJSONText(value: unknown): string {
  return typeof value === "string" ? value : JSON.stringify(value) ?? "";
}

function toWebhookRecord(item: WebhookRecordData): WebhookRecord {
  return {
    ...item,
    payload: toJSONText(item.payload),
    request_headers: toJSONText(item.request_headers),
    response_headers: toJSONText(item.response_headers),
  };
}

export function createManagementAPI(baseURL: string) {
  const accountApi = {
    async listAll(): Promise<Account[]> {
      const result = (await request.get<Page<AccountData>>(`${baseURL}/accounts`)).data;
      return result.items.map(toAccount);
    },
    async list(pageNumber = 1, pageSize = 20, filters: AccountListRequest = {}): Promise<ListResponse<Account>> {
      const result = (await request.get<Page<AccountData>>(`${baseURL}/accounts`, {
        params: { ...filters, page_number: pageNumber, page_size: pageSize },
      })).data;
      return { data: result.items.map(toAccount), total_items: result.total };
    },
    async create(payload: Pick<Account, "name" | "currency">) {
      return toAccount((await request.post<AccountData>(`${baseURL}/accounts`, payload)).data);
    },
    async update(id: number, payload: Pick<Account, "name">) {
      return toAccount((await request.post<AccountData>(`${baseURL}/accounts/rename`, { id, ...payload })).data);
    },
  };

  const webhookApi = {
    async events() {
      return (await request.get<{ items: WebhookEvent[] }>(`${baseURL}/webhooks/events`)).data.items;
    },
    async list(query: WebhookListRequest): Promise<ListResponse<Webhook>> {
      const result = (await request.get<Page<Webhook>>(`${baseURL}/webhooks`, { params: query })).data;
      return { data: result.items, total_items: result.total };
    },
    async create(payload: Omit<Webhook, "id" | "created_at" | "updated_at" | "account_name">) {
      return (await request.post<Webhook>(`${baseURL}/webhooks`, payload)).data;
    },
    async update(webhook: Pick<Webhook, "id" | "account_id">, payload: Pick<Webhook, "target_url" | "enabled">) {
      return (await request.post<Webhook>(`${baseURL}/webhooks/update`, {
        id: webhook.id, account_id: webhook.account_id, ...payload,
      })).data;
    },
    async remove(webhook: Pick<Webhook, "id" | "account_id">) {
      await request.post(`${baseURL}/webhooks/delete`, { id: webhook.id, account_id: webhook.account_id });
    },
  };

  const authorizationConfigApi = {
    async get(accountId: number) {
      return (await request.get<AuthorizationConfig>(`${baseURL}/authorization-config`, {
        params: {
          account_id: accountId,
        },
      })).data;
    },
    async update(
      payload: Pick<
        AuthorizationConfig,
        "account_id" | "target_url" | "enabled" | "timeout_millis"
      >,
    ) {
      return (await request.post<AuthorizationConfig>(`${baseURL}/authorization-config`, payload)).data;
    },
  };

  const managementApi = {
    async cardProducts(page: PageRequest): Promise<ListResponse<CardProduct>> {
      const result = (await request.get<Page<CardProduct>>(`${baseURL}/card-products`, { params: page })).data;
      return { data: result.items, total_items: result.total };
    },
    async virtualAccounts(page: PageRequest): Promise<ListResponse<VirtualAccount>> {
      const result = (await request.get<Page<VirtualAccountData & { created_at: string }>>(`${baseURL}/virtual-accounts`, {
        params: page,
      })).data;
      return {
        data: result.items.map(item => ({ ...item, balance: item.available })),
        total_items: result.total,
      };
    },
  };

  const refundApi = {
    async simulate(payload: RefundSimulationRequest) {
      return (await request.post(`${baseURL}/simulate/refunds`, {
        authorization_id: payload.authorization_id,
        card_id: payload.card_id,
        amount: payload.amount,
        currency: payload.currency,
        merchant_name: payload.merchant_name || undefined,
        merchant_mcc: payload.merchant_category_code || undefined,
        merchant_country: payload.merchant_country || undefined,
        request_id: payload.request_id,
      })).data;
    },
  };

  const api: ChannelAPI = {
    async listCardholders(page) {
      const result = (await request.get<Page<CardholderData>>(`${baseURL}/cardholders`, { params: page })).data;
      return { data: result.items.map(item => ({ ...item, phone_number: item.mobile })), total_items: result.total };
    },
    async listCards(page) {
      const { card_status, ...filters } = page ?? {};
      const result = (await request.get<Page<CardData>>(`${baseURL}/cards`, {
        params: { ...filters, status: card_status },
      })).data;
      return { data: result.items.map(toCard), total_items: result.total };
    },
    async updateCardStatus(payload) {
      return (await request.post<CardData>(`${baseURL}/cards/status`, {
        id: payload.id, account_id: payload.account_id, status: payload.card_status,
      })).data;
    },
    async listTransactions(page) {
      const { transaction_type, ...filters } = page ?? {};
      const result = (await request.get<Page<TransactionData>>(`${baseURL}/transactions`, {
        params: { ...filters, type: transaction_type },
      })).data;
      return { data: result.items.map(toTransaction), total_items: result.total };
    },
    async simulateAuthorization(payload) {
      await request.post(`${baseURL}/simulate/authorizations`, authorizationPayload(payload));
    },
    async applyTransactionStep(id, action, amount) {
      await request.post(`${baseURL}/transactions/stages`, {
        id, type: action === "reverse" ? "void" : action, amount,
      });
    },
  };

  const fundsApi = {
    async list(query: WalletListRequest): Promise<ListResponse<Wallet>> {
      const result = (await request.get<Page<WalletData>>(`${baseURL}/wallets`, { params: query })).data;
      return {
        data: result.items.map(item => ({ ...item, kind: item.type, amount: item.available })),
        total_items: result.total,
      };
    },
    async adjustAccount(account: Account, amount: number) {
      await request.post(`${baseURL}/accounts/adjust`, {
        account_id: account.id, amount, currency: account.currency,
      });
    },
  };

  const webhookRecordApi = {
    async list(query: WebhookRecordListRequest): Promise<ListResponse<WebhookRecord>> {
      const result = (await request.get<Page<WebhookRecordData>>(`${baseURL}/webhook-records`, { params: query })).data;
      return { data: result.items.map(toWebhookRecord), total_items: result.total };
    },
    async replay(record: Pick<WebhookRecord, "id" | "account_id">) {
      return toWebhookRecord((await request.post<WebhookRecordData>(`${baseURL}/webhook-records/replay`, {
        id: record.id, account_id: record.account_id,
      })).data);
    },
  };

  const authorizationApi: AuthorizationAPI = {
    async list(query) {
      const result = (await request.get<Page<Authorization>>(`${baseURL}/authorizations`, { params: query })).data;
      return { data: result.items, total_items: result.total };
    },
    async detail(authorization) {
      const item = (await request.get<AuthorizationDetailData>(`${baseURL}/authorizations/detail`, {
        params: { id: authorization.id, account_id: authorization.account_id },
      })).data;
      return {
        ...item,
        card_number: item.card.card_number,
        transactions: item.transactions.map(stage => ({ ...stage, transaction_type: stage.type })),
      };
    },
    async stage(input) {
      const endpoint = { clear: "clearings", reverse: "reversals", refund: "refunds" }[input.stage];
      await request.post(`${baseURL}/simulate/${endpoint}`, {
        authorization_id: input.authorization.id, amount: input.amount, request_id: input.requestId,
      });
    },
  };

  const virtualAccountApi: VirtualAccountAPI = {
    async list(query) {
      const result = (await request.get<Page<VirtualAccountData>>(`${baseURL}/virtual-accounts`, { params: query })).data;
      return {
        data: result.items.map(item => ({ ...item, balance: item.available })),
        total_items: result.total,
      };
    },
    async create(input) {
      await request.post(`${baseURL}/virtual-accounts`, input);
    },
    async topUp(input) {
      await request.post(`${baseURL}/virtual-accounts/fund`, {
        virtual_account_id: input.account.id, account_id: input.account.account_id,
        amount: input.amount, request_id: input.requestId,
      });
    },
    async withdraw(input) {
      await request.post(`${baseURL}/virtual-accounts/fund`, {
        virtual_account_id: input.account.id, account_id: input.account.account_id,
        amount: input.amount, request_id: input.requestId, withdraw: true,
      });
    },
  };

  async function fundCard(input: CardFundingRequest) {
    await request.post(`${baseURL}/cards/fund`, {
      card_id: input.card.id, account_id: input.card.account_id,
      kind: input.withdraw ? "card_withdraw" : "card_top_up", amount: input.amount, request_id: input.requestId,
    });
  }

  return {
    fundCard,
    virtualAccountApi,
    authorizationApi,
    accountApi,
    webhookApi,
    authorizationConfigApi,
    managementApi,
    refundApi,
    api,
    fundsApi,
    webhookRecordApi,
  };
}

export type ManagementAPI = ReturnType<typeof createManagementAPI>;
