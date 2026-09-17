export interface Cardholder {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone_number: string;
  status: string;
  created_at: string;
}

export interface Card {
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

export interface ChannelAPI {
  listCardholders(): Promise<ListResponse<Cardholder>>;

  createCardholder(payload: {
    firstName: string;
    lastName: string;
    email: string;
    mobile: string;
  }): Promise<Cardholder>;

  listCards(): Promise<ListResponse<Card>>;

  createCard(cardholderID: string, currency: string): Promise<Card>;

  updateCardStatus(id: string, status: string): Promise<Card>;

  listTransactions(): Promise<ListResponse<Transaction>>;

  simulateAuthorization(payload: {
    cardID: string;
    amount: number;
    currency: string;
    merchantName: string;
    merchantMCC: string;
    merchantCountry: string;
    merchantCity: string;
  }): Promise<void>;

  applyTransactionStep(
      id: string,
      action: "clear" | "reverse" | "refund",
      amount?: number,
  ): Promise<void>;
}
