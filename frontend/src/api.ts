import { channel } from '@/config'
import { api as payndaAPI } from '@/channel/paynda/api'
import { api as photonPayAPI } from '@/channel/photonpay/api'
import { api as slashAPI } from '@/channel/slash/api'

export type { Card, Cardholder, Transaction } from '@/channel/types'

export const api = {
  slash: slashAPI,
  photonpay: photonPayAPI,
  paynda: payndaAPI,
}[channel]
