import type { AuthorizationSimulationRequest, SimulationStage } from "./shared/simulation";
import type { CardStatus, TransactionStatus, TransactionType } from "./enums";

export interface Cardholder {
  account_id: number;
  account_name: string;
  id: number;
  first_name: string;
  last_name: string;
  email: string;
  phone_number: string;
  status: "normal";
  created_at: string;
}

export interface Card {
  account_id: number;
  account_name: string;
  wallet_id?: number;
  id: number;
  cardholder_id?: number;
  card_number: string;
  card_bin: string;
  card_currency: string;
  card_status: CardStatus;
  cvv?: string;
  expires_at?: string;
  created_at: string;
  funding_source?: string;
  balance: string;
  reserved?: string;
  virtual_account_id?: number;
}

export interface Transaction {
  account_id: number;
  account_name: string;
  id: number;
  card_id: number;
  authorization_id: number;
  transaction_type: TransactionType;
  status: TransactionStatus;
  amount: string;
  currency: string;
  merchant_name: string;
  merchant_category_code: string;
  transacted_at: string;
}

export interface ListResponse<T> {
  total_items: number;
  data: T[];
}

export interface PageRequest {
  page_number: number;
  page_size: number;
}

export interface CardListRequest extends PageRequest {
  account_id?: number;
  id?: number;
  card_number?: string;
  card_status?: CardStatus;
  created_from?: string;
  created_to?: string;
}

export interface TransactionListRequest extends PageRequest {
  account_id?: number;
  id?: number;
  card_id?: number;
  authorization_id?: number;
  transaction_type?: TransactionType;
  status?: TransactionStatus;
  created_from?: string;
  created_to?: string;
}

export interface CardStatusUpdateRequest {
  id: number;
  account_id: number;
  card_status: CardStatus;
}

export interface ChannelAPI {
  listCardholders(page?: PageRequest): Promise<ListResponse<Cardholder>>;

  listCards(query?: CardListRequest): Promise<ListResponse<Card>>;

  updateCardStatus(payload: CardStatusUpdateRequest): Promise<unknown>;

  listTransactions(query?: TransactionListRequest): Promise<ListResponse<Transaction>>;

  simulateAuthorization(payload: AuthorizationSimulationRequest): Promise<void>;

  applyTransactionStep(
    id: number,
    action: SimulationStage,
    amount?: number,
  ): Promise<void>;
}
