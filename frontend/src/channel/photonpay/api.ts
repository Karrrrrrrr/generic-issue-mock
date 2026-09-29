import { createManagementAPI } from "@/channel/shared/api";

export type {
  Account,
  AuthorizationConfig,
  CardProduct,
  VirtualAccount,
  Wallet,
  Webhook,
  WebhookEvent,
  WebhookRecord,
} from "@/channel/shared/api";

export const {
  fundCard,
  virtualAccountApi,
  authorizationApi,
  accountApi,
  webhookApi,
  authorizationConfigApi,
  managementApi,
  refundApi,
  api,
  fundsApi,
  webhookRecordApi,
} =
  createManagementAPI("/photonpay/ui");
