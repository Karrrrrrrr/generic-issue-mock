export type Channel = "slash" | "photonpay" | "paynda";

const channelFromPath = window.location.pathname.split("/").filter(Boolean)[0];

export const channel: Channel =
  channelFromPath === "photonpay" || channelFromPath === "paynda"
    ? channelFromPath
    : "slash";

export const channelConfig = {
  slash: {
    label: "Slash",
    supportsAuthorizationSimulation: true,
    supportsCardStatus: true,
    supportsCardholder: true,
  },
  photonpay: {
    label: "PhotonPay",
    supportsAuthorizationSimulation: true,
    supportsCardStatus: true,
    supportsCardholder: true,
  },
  paynda: {
    label: "Paynda",
    supportsAuthorizationSimulation: true,
    supportsCardStatus: true,
    supportsCardholder: true,
  },
}[channel];
