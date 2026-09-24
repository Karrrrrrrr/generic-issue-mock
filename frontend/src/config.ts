export type Channel = "slash" | "photonpay" | "paynda" | "pingpong";

const channelFromPath = window.location.pathname.split("/").filter(Boolean)[0];

export const channel: Channel =
  channelFromPath === "photonpay" ||
  channelFromPath === "paynda" ||
  channelFromPath === "pingpong"
    ? channelFromPath
    : "slash";

export const channelConfig = {
  pingpong: {
    label: "PingPong",
    supportsAuthorizationSimulation: true,
    supportsCardStatus: true,
    supportsCardholder: false,
  },
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
