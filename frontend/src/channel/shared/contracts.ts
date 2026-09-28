import type { TransactionStatus, TransactionType } from "@/channel/enums";
import type { SimulationStage } from "./simulation";
import type { Card } from "@/channel/types";

export interface CardFundingRequest {
  card: Card;
  amount: number;
  withdraw: boolean;
  requestID?: string;
}

export interface AuthorizationTransaction {
  id: string;
  transaction_type: TransactionType;
  status: TransactionStatus;
  amount: string;
  currency: string;
  created_at: string;
}

export interface Authorization {
  id: string;
  account_id: string;
  account_name: string;
  card_id: string;
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

export interface AuthorizationAPI {
  list: (filters: Record<string, string | undefined>) => Promise<Authorization[]>;
  detail?: (authorization: Authorization) => Promise<Authorization>;
  stage: (request: {
    authorization: Authorization;
    stage: SimulationStage;
    amount: number;
    requestID?: string;
  }) => Promise<unknown>;
}

export interface ManagedVirtualAccount {
  id: string;
  account_id: string;
  account_name: string;
  name: string;
  wallet_id?: string;
  balance: string;
  currency: string;
}

export interface VirtualAccountFundingRequest {
  account: ManagedVirtualAccount;
  amount: number;
  requestID?: string;
}

export interface VirtualAccountAPI {
  list: () => Promise<ManagedVirtualAccount[]>;
  create: (request: { account_id: string; name: string; currency: string }) => Promise<unknown>;
  topUp: (request: VirtualAccountFundingRequest) => Promise<unknown>;
  withdraw?: (request: VirtualAccountFundingRequest) => Promise<unknown>;
}
