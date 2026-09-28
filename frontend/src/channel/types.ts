import type { AuthorizationSimulationRequest, SimulationStage } from "./shared/simulation";

export interface Cardholder {
  account_id: string;
  account_name: string;
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone_number: string;
  status: string;
  created_at: string;
}

export interface Card {
  account_id: string;
  account_name: string;
  wallet_id: string;
  id: string;
  cardholder_id: string;
  card_number: string;
  card_bin: string;
  card_currency: string;
  card_status: string;
  cvv: string;
  expires_at: string;
  created_at: string;
  funding_source: string;
  balance: string;
}

export interface Transaction {
  account_id: string;
  account_name: string;
  id: string;
  card_id: string;
  authorization_id: string;
  transaction_type: string;
  status: string;
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
  account_id?: string;
  id?: string;
  card_number?: string;
  card_status?: string;
  created_from?: string;
  created_to?: string;
}

export interface TransactionListRequest extends PageRequest {
  account_id?: string;
  id?: string;
  card_id?: string;
  authorization_id?: string;
  transaction_type?: string;
  status?: string;
  created_from?: string;
  created_to?: string;
}

export interface CardStatusUpdateRequest {
  id: string;
  account_id: string;
  card_status: string;
}

export interface ChannelAPI {
  listCardholders(page?: PageRequest): Promise<ListResponse<Cardholder>>;

  listCards(query?: CardListRequest): Promise<ListResponse<Card>>;

  updateCardStatus(payload: CardStatusUpdateRequest): Promise<Card>;

  listTransactions(query?: TransactionListRequest): Promise<ListResponse<Transaction>>;

  simulateAuthorization(payload: AuthorizationSimulationRequest): Promise<void>;

  applyTransactionStep(
    id: string,
    action: SimulationStage,
    amount?: number,
  ): Promise<void>;
}
