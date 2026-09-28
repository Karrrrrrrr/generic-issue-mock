import { createManagementAPI } from "@/channel/shared/api";

export type { Account, CardProduct, VirtualAccount, Wallet, Webhook, WebhookEvent, WebhookRecord } from "@/channel/shared/api";

export const { fundCard, virtualAccountApi, authorizationApi, accountApi, webhookApi, managementApi, refundApi, api, fundsApi, webhookRecordApi } =
  createManagementAPI("/paynda/ui");
