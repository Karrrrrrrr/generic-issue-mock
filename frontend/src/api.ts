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
  expiry_month: string
  expiry_year: string
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

interface PhotonResponse<T> {
  code: string
  msg: string
  data: T
}

const request = axios.create({ timeout: 10_000 })

function photonData<T>(response: PhotonResponse<T>): T {
  if (response.code !== '000000') {
    throw new Error(response.msg)
  }
  return response.data
}

export const api = {
  async listCardholders(): Promise<SlashList<Cardholder>> {
    if (channel === 'slash') {
      return (await request.get<SlashList<Cardholder>>('/slash/ui/cardholders')).data
    }
    const response = await request.get<PhotonResponse<Array<{
      cardholderId: string
      firstName: string
      lastName: string
      email: string
      mobile: string
      status: string
      createdAt: string
    }>>>('/photonpay/vcc/openApi/v4/pagingVccCardholder')
    const items = photonData(response.data).map((item) => ({
      id: item.cardholderId,
      first_name: item.firstName,
      last_name: item.lastName,
      email: item.email,
      phone_number: item.mobile,
      status: item.status,
      created_at: item.createdAt,
    }))
    return { total_items: items.length, data: items }
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
    const response = await request.post<PhotonResponse<{
      cardholderId: string
      status: string
    }>>('/photonpay/vcc/openApi/v4/addCardholder', {
      firstName: payload.firstName,
      lastName: payload.lastName,
      email: payload.email,
      mobile: payload.mobile,
      mobilePrefix: '+1',
      dateOfBirth: '1990-01-01',
      nationalityCountryCode: 'US',
    })
    const item = photonData(response.data)
    return {
      id: item.cardholderId,
      first_name: payload.firstName,
      last_name: payload.lastName,
      email: payload.email,
      phone_number: payload.mobile,
      status: item.status,
      created_at: new Date().toISOString(),
    }
  },
  async listCards(): Promise<SlashList<Card>> {
    if (channel !== 'slash') {
      return { total_items: 0, data: [] }
    }
    return (await request.get<SlashList<Card>>('/slash/ui/cards')).data
  },
  async createCard(cardholderID: string, currency: string): Promise<Card> {
    if (channel === 'slash') {
      return (await request.post<Card>('/slash/ui/cards', {
        cardholder_id: cardholderID,
        card_currency: currency,
      })).data
    }
    const response = await request.post<PhotonResponse<{
      cardDetail: {
        cardId: string
        cardNo: string
        cardCurrency: string
        cardStatus: string
        cvv: string
        expirationDate: string
      }
    }>>('/photonpay/vcc/openApi/v4/openCard', {
      cardBin: '424242',
      cardCurrency: currency,
      cardType: 'single',
      cardholderId: cardholderID,
      requestId: crypto.randomUUID(),
    })
    const item = photonData(response.data).cardDetail
    return {
      id: item.cardId,
      cardholder_id: cardholderID,
      card_number: item.cardNo,
      card_bin: item.cardNo.slice(0, 6),
      card_currency: item.cardCurrency,
      card_status: item.cardStatus,
      cvv: item.cvv,
      expiry_month: item.expirationDate.slice(0, 2),
      expiry_year: item.expirationDate.slice(-4),
      created_at: new Date().toISOString(),
    }
  },
  async updateCardStatus(id: string, status: string): Promise<Card> {
    return (await request.put<Card>(`/slash/ui/cards/${id}/status`, { card_status: status })).data
  },
  async listTransactions(): Promise<SlashList<Transaction>> {
    if (channel === 'slash') {
      return (await request.get<SlashList<Transaction>>('/slash/ui/transactions')).data
    }
    const response = await request.get<PhotonResponse<Array<{
      transactionId: string
      cardId: string
      transactionAmount: number
      transactionCurrency: string
      merchantName: string
      status: string
    }>>>('/photonpay/vcc/openApi/v4/pagingVccTradeOrder')
    const items = photonData(response.data).map((item) => ({
      id: item.transactionId,
      card_id: item.cardId,
      authorization_id: '',
      transaction_type: 'auth',
      status: item.status,
      amount: String(item.transactionAmount),
      currency: item.transactionCurrency,
      merchant_name: item.merchantName,
      merchant_category_code: '',
      transacted_at: '',
    }))
    return { total_items: items.length, data: items }
  },
  async simulateAuthorization(payload: { cardID: string; amount: number; currency: string; merchantName: string; merchantMCC: string; merchantCountry: string }): Promise<void> {
    await request.post('/slash/ui/simulate/authorizations', {
      card_id: payload.cardID,
      transaction_amount: payload.amount,
      transaction_currency: payload.currency,
      merchant_name: payload.merchantName,
      merchant_category_code: payload.merchantMCC,
      merchant_country: payload.merchantCountry,
    })
  },
  async applyTransactionStep(id: string, action: 'clear' | 'reverse' | 'refund'): Promise<void> {
    await request.post(`/slash/ui/transactions/${id}/${action}`, {})
  },
}
