import type { Card, Cardholder, ChannelAPI, ListResponse, Transaction } from "@/channel/types";
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
  currency?: string;
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

export interface CardProduct {
  id: number;
  prefix: string;
}

export interface VirtualAccount {
  id: number;
  name: string;
  currency: string;
  funding_source: string;
  balance: string;
  spend: string;
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

export function createManagementAPI(baseURL: string) {
  const accountApi = {
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
    async update(id: number, payload: Pick<Account, "name">) {
      return (await request.put<Account>(`${baseURL}/accounts/${id}`, payload)).data;
    },
  };

  const webhookApi = {
    async events() {
      return (await request.get<WebhookEvent[]>(`${baseURL}/webhooks/events`)).data;
    },
    async list(accountID?: number) {
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
    async update(id: number, payload: Pick<Webhook, "target_url" | "enabled">) {
      return (await request.put<Webhook>(`${baseURL}/webhooks/${id}`, payload)).data;
    },
    async remove(id: number) {
      await request.delete(`${baseURL}/webhooks/${id}`);
    },
  };

  const managementApi = {
    async cardProducts() {
      return (await request.get<{ items: CardProduct[] }>(`${baseURL}/card-products`)).data.items;
    },
    async virtualAccounts() {
      return (await request.get<VirtualAccount[]>(`${baseURL}/virtual-accounts`)).data;
    },
  };

  const refundApi = {
    async simulate(payload: RefundSimulationRequest) {
      return (await request.post<Transaction>(`${baseURL}/simulate/refunds`, payload)).data;
    },
  };

  const api: ChannelAPI = {
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

  const fundsApi = {
    async list(accountID?: number) {
      return (
        await request.get<Wallet[]>(baseURL + "/funds", {
          params: {
            account_id: accountID,
          },
        })
      ).data;
    },

    async adjustAccount(account: Account, amount: number) {
      if (!account.wallet_id) {
        throw new Error("账户钱包不存在");
      }
      await request.post(baseURL + "/funds/transfer", {
        account_id: account.id,
        source_id: amount < 0 ? account.wallet_id : undefined,
        target_id: amount > 0 ? account.wallet_id : undefined,
        amount: String(Math.abs(amount)),
      });
    },

    async transferResource(input: {
      accountID: number;
      walletID: number;
      cardID?: number;
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
        card_id: input.cardID,
        source_id: input.withdraw ? input.walletID : accountWallet.id,
        target_id: input.withdraw ? accountWallet.id : input.walletID,
        amount: String(input.amount),
      });
    },
  };

  const webhookRecordApi = {
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
  const authorizationApi: AuthorizationAPI = {
    async list(filters) {
      return (await request.get<Authorization[]>(`${baseURL}/authorization-balances`, {
        params: filters,
      })).data;
    },
    async detail(authorization) {
      return (await request.get<Authorization>(`${baseURL}/authorizations/${authorization.id}/detail`, {
        params: { account_id: authorization.account_id },
      })).data;
    },
    async stage(input) {
      await request.post(`${baseURL}/authorizations/${input.authorization.id}/${input.stage}`, {
        amount: String(input.amount),
      });
    },
  };

  const virtualAccountApi: VirtualAccountAPI = {
    async list() {
      const [response, wallets] = await Promise.all([
        request.get<Omit<ManagedVirtualAccount, "balance" | "currency">[]>(`${baseURL}/managed-virtual-accounts`),
        fundsApi.list(),
      ]);
      return response.data.map((account) => {
        const wallet = wallets.find((item) => item.id === account.wallet_id && item.account_id === account.account_id);
        if (!wallet) {
          throw new Error("虚拟账户钱包不存在");
        }
        return { ...account, balance: wallet.amount, currency: wallet.currency };
      });
    },
    async create(input) {
      await request.post(`${baseURL}/managed-virtual-accounts`, input);
    },
    async topUp(input) {
      if (!input.account.wallet_id) {
        throw new Error("虚拟账户钱包不存在");
      }
      await fundsApi.transferResource({
        accountID: input.account.account_id,
        walletID: input.account.wallet_id,
        amount: input.amount,
        withdraw: false,
      });
    },
    async withdraw(input) {
      if (!input.account.wallet_id) {
        throw new Error("虚拟账户钱包不存在");
      }
      await fundsApi.transferResource({
        accountID: input.account.account_id,
        walletID: input.account.wallet_id,
        amount: input.amount,
        withdraw: true,
      });
    },
  };

  async function fundCard(input: CardFundingRequest) {
    if (!input.card.wallet_id) {
      throw new Error("卡钱包不存在");
    }
    await fundsApi.transferResource({
      accountID: input.card.account_id,
      cardID: input.card.id,
      walletID: input.card.wallet_id,
      amount: input.amount,
      withdraw: input.withdraw,
    });
  }

  return {
    fundCard,
    virtualAccountApi,
    authorizationApi,
    accountApi,
    webhookApi,
    managementApi,
    refundApi,
    api,
    fundsApi,
    webhookRecordApi,
  };
}

export type ManagementAPI = ReturnType<typeof createManagementAPI>;
