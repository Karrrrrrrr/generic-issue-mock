import axios from 'axios'

export const request = axios.create({ timeout: 10_000 })

request.interceptors.response.use(
  (response) => response,
  (error) => {
    const message = error.response?.data?.message || error.response?.data?.msg
    return Promise.reject(new Error(message || '请求失败，请检查服务是否已启动'))
  },
)

export function authorizationPayload(payload: { cardID: string; amount: number; currency: string; merchantName: string; merchantMCC: string; merchantCountry: string }) {
  return {
    card_id: payload.cardID,
    transaction_amount: payload.amount,
    transaction_currency: payload.currency,
    merchant_name: payload.merchantName,
    merchant_category_code: payload.merchantMCC,
    merchant_country: payload.merchantCountry,
  }
}
