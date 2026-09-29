import { createManagementAPI } from "@/channel/shared/api";
import type { AuthorizationAPI, CardFundingRequest } from "@/channel/shared/contracts";
import type { AuthorizationSimulationRequest, RefundSimulationRequest } from "@/channel/shared/simulation";
import type { Card, ChannelAPI } from "@/channel/types";
import { api, requestId } from "./api";

const management = createManagementAPI("/pingpong/ui");
export const accountApi = management.accountApi;
export const adjustAccountBalance = management.fundsApi.adjustAccount;
export const cardApi = management.api;
export const virtualAccountApi = management.virtualAccountApi;
export const loadProducts = management.managementApi.cardProducts;

export async function fundCard(input: CardFundingRequest) {
  await management.fundCard(input);
}

export async function simulateAuthorization(input: AuthorizationSimulationRequest) {
	await management.api.simulateAuthorization(input);
}

export async function simulateRefund(input: RefundSimulationRequest & { requestId?: string }) {
	await management.refundApi.simulate({
		...input,
		request_id: input.request_id ?? input.requestId,
	});
}

export async function loadSimulationCards(): Promise<Card[]> {
	const result = await cardApi.listCards();
	return result.data;
}

export const authorizationApi: AuthorizationAPI = management.authorizationApi;
export const webhookApi = management.webhookApi;
export const webhookRecordApi = management.webhookRecordApi;

const pendingTransactionRequests = new Map<string, string>();
export const transactionApi: Pick<ChannelAPI, "listTransactions" | "applyTransactionStep"> = {
  listTransactions: management.api.listTransactions,
  async applyTransactionStep(id, action, amount) {
    const key = JSON.stringify([id, action, amount]);
    const orderId = pendingTransactionRequests.get(key) ?? requestId();
    pendingTransactionRequests.set(key, orderId);
    await api.post("transactions/stages", {
      id, type: action === "reverse" ? "void" : action, amount, request_id: orderId,
    });
    pendingTransactionRequests.delete(key);
  },
};
