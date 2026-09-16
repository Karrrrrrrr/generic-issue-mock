import axios from 'axios'

export const request = axios.create({ timeout: 10_000 })

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
