export type Channel = 'slash' | 'photonpay'

const value = import.meta.env.VITE_CHANNEL?.trim().toLowerCase()

export const channel: Channel = value === 'photonpay' ? 'photonpay' : 'slash'

export const channelConfig = {
  slash: {
    label: 'Slash',
    supportsAuthorizationSimulation: true,
    supportsCardStatus: true,
    supportsCardholder: true,
  },
  photonpay: {
    label: 'PhotonPay',
    supportsAuthorizationSimulation: true,
    supportsCardStatus: true,
    supportsCardholder: true,
  },
}[channel]
