import type { TransactionStatus, TransactionType } from "@/channel/enums";
import type { SimulationStage } from "./simulation";
import type { Card, ListResponse, PageRequest } from "@/channel/types";

export interface CardFundingRequest {
  card: Card;
  amount: number;
  withdraw: boolean;
  requestId?: string;
}

export interface AuthorizationTransaction {
  id: number;
  transaction_type: TransactionType;
  status: TransactionStatus;
  amount: string;
  currency: string;
  created_at: string;
}

export interface Authorization {
  id: number;
  account_id: number;
  account_name: string;
  card_id: number;
  status: TransactionStatus;
  amount: string;
  settled?: string;
  reversed?: string;
  refunded?: string;
  remaining: string;
  currency: string;
  merchant_name: string;
  created_at: string;
  notification_status?: "contract_pending";
  card_number?: string;
  authorization_code?: string;
  merchant_country?: string;
  merchant_mcc?: string;
  raw_payload?: unknown;
  transactions?: AuthorizationTransaction[];
}

export type AuthorizationListRequest = PageRequest & Record<string, string | number | undefined>;

export interface AuthorizationAPI {
  list: (query: AuthorizationListRequest) => Promise<ListResponse<Authorization>>;
  detail?: (authorization: Authorization) => Promise<Authorization>;
  stage: (request: {
    authorization: Authorization;
    stage: SimulationStage;
    amount: number;
    requestId?: string;
  }) => Promise<unknown>;
}

export interface ManagedVirtualAccount {
  id: number;
  account_id: number;
  account_name: string;
  name: string;
  wallet_id?: number;
  balance: string;
  currency: string;
}

export interface VirtualAccountListRequest extends PageRequest {
  account_id?: number;
}

export interface VirtualAccountFundingRequest {
  account: ManagedVirtualAccount;
  amount: number;
  requestId?: string;
}

export interface VirtualAccountAPI {
  list: (query: VirtualAccountListRequest) => Promise<ListResponse<ManagedVirtualAccount>>;
  create: (request: { account_id: number; name: string; currency: string }) => Promise<unknown>;
  topUp: (request: VirtualAccountFundingRequest) => Promise<unknown>;
  withdraw?: (request: VirtualAccountFundingRequest) => Promise<unknown>;
}
