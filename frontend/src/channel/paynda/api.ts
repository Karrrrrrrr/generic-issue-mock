import type {
  ChannelAPI,
  Card,
  Cardholder,
  ListResponse,
  Transaction,
} from "@/channel/types";
import { authorizationPayload, request } from "@/channel/shared";

const baseURL = "/paynda/ui";

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
