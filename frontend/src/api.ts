import axios from 'axios'

import { channel } from '@/config'

export interface Cardholder {
  id: string
  first_name: string
  last_name: string
  email: string
  phone_number: string
  status: string
  created_at: string
}

export interface Card {
  id: string
  cardholder_id: string
  card_number: string
  card_bin: string
  card_currency: string
  card_status: string
  cvv: string
  expires_at: string
  created_at: string
}

export interface Transaction {
  id: string
  card_id: string
  authorization_id: string
  transaction_type: string
  status: string
  amount: string
  currency: string
  merchant_name: string
  merchant_category_code: string
  transacted_at: string
}

interface SlashList<T> {
  total_items: number
  data: T[]
}

const request = axios.create({ timeout: 10_000 })

export const api = {
  async listCardholders(): Promise<SlashList<Cardholder>> {
    if (channel === 'slash') {
      return (await request.get<SlashList<Cardholder>>('/slash/ui/cardholders')).data
    }
    return (await request.get<SlashList<Cardholder>>('/photonpay/ui/cardholders')).data
  },
  async createCardholder(payload: { firstName: string; lastName: string; email: string; mobile: string }): Promise<Cardholder> {
    if (channel === 'slash') {
      return (await request.post<Cardholder>('/slash/ui/cardholders', {
        first_name: payload.firstName,
        last_name: payload.lastName,
        email: payload.email,
        phone_number: payload.mobile,
      })).data
    }
    return (await request.post<Cardholder>('/photonpay/ui/cardholders', {
      first_name: payload.firstName,
      last_name: payload.lastName,
      email: payload.email,
      phone_number: payload.mobile,
    })).data
  },
  async listCards(): Promise<SlashList<Card>> {
    return (await request.get<SlashList<Card>>(`/${channel}/ui/cards`)).data
  },
  async createCard(cardholderID: string, currency: string): Promise<Card> {
    if (channel === 'slash') {
      return (await request.post<Card>('/slash/ui/cards', {
        cardholder_id: cardholderID,
        card_currency: currency,
      })).data
    }
    return (await request.post<Card>('/photonpay/ui/cards', {
      cardholder_id: cardholderID,
      card_currency: currency,
      request_id: crypto.randomUUID(),
    })).data
  },
  async updateCardStatus(id: string, status: string): Promise<Card> {
    return (await request.put<Card>(`/${channel}/ui/cards/${id}/status`, { card_status: status })).data
  },
  async listTransactions(): Promise<SlashList<Transaction>> {
    if (channel === 'slash') {
      return (await request.get<SlashList<Transaction>>('/slash/ui/transactions')).data
    }
    return (await request.get<SlashList<Transaction>>('/photonpay/ui/transactions')).data
  },
  async simulateAuthorization(payload: { cardID: string; amount: number; currency: string; merchantName: string; merchantMCC: string; merchantCountry: string }): Promise<void> {
    await request.post(`/${channel}/ui/simulate/authorizations`, {
      card_id: payload.cardID,
      transaction_amount: payload.amount,
      transaction_currency: payload.currency,
      merchant_name: payload.merchantName,
      merchant_category_code: payload.merchantMCC,
      merchant_country: payload.merchantCountry,
    })
  },
  async applyTransactionStep(id: string, action: 'clear' | 'reverse' | 'refund'): Promise<void> {
    await request.post(`/${channel}/ui/transactions/${id}/${action}`, {})
  },
}
