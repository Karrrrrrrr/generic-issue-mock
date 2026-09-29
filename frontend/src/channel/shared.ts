import axios from "axios";
import type { AuthorizationSimulationRequest } from "./shared/simulation";

export const request = axios.create({ timeout: 10_000 });

request.interceptors.response.use(
  (response) => response,
  (error) => {
    const message = error.response?.data?.message || error.response?.data?.msg;
    return Promise.reject(new Error(message || "请求失败，请检查服务是否已启动"));
  },
);

export function authorizationPayload(payload: AuthorizationSimulationRequest) {
  return {
    card_id: payload.cardId,
    amount: payload.amount,
    currency: payload.currency,
    merchant_name: payload.merchantName || undefined,
    merchant_mcc: payload.merchantMCC || undefined,
    merchant_country: payload.merchantCountry || undefined,
    request_id: payload.requestId,
  };
}
