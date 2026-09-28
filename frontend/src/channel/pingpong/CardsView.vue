<script setup lang="ts">
import { computed, ref } from "vue";
import { NAlert, NModal } from "naive-ui";
import SharedView from "@/channel/shared/CardsView.vue";
import CardAuthorizationForm from "@/channel/shared/CardAuthorizationForm.vue";
import type { AuthorizationSimulationRequest, SimulationCard } from "@/channel/shared/simulation";
import type { Card } from "@/channel/types";
import { accountApi, cardApi, fundCard, simulateAuthorization } from "./management";
import { requestID } from "./api";

const view = ref<InstanceType<typeof SharedView>>();
const selected = ref<Card>();
const busy = ref(false);
const orderID = ref("");
const simulationCards = computed<SimulationCard[]>(() => selected.value ? [{
  id: selected.value.id,
  label: `${selected.value.account_name} · ${selected.value.card_number}`,
  currency: selected.value.card_currency,
  disabled: selected.value.card_status !== "active",
}] : []);

function openAuthorization(card: Card) {
  selected.value = card;
  orderID.value = requestID();
}

async function authorize(request: AuthorizationSimulationRequest) {
  await simulateAuthorization({ ...request, requestID: orderID.value });
}

function completeAuthorization() {
  selected.value = undefined;
  void view.value?.reload();
}

function closeAuthorization(shown: boolean) {
  if (!shown && !busy.value) {
    selected.value = undefined;
  }
}
</script>

<template>
  <SharedView
    ref="view"
    :api="cardApi"
    :account-api="accountApi"
    :fund-card="fundCard"
    :authorize-card="openAuthorization"
    :show-expiry="false"
    show-reserved
    show-virtual-account
    :detailed-filters="false"
    :new-request-i-d="requestID"
    funding-source-label="所属虚拟账户"
  >
    <template #description>
      <n-alert type="info" :show-icon="false">
        卡拥有独立钱包。充值从所属虚拟账户扣款，转出返回同一虚拟账户；开卡通过 OpenAPI 完成。
      </n-alert>
    </template>
  </SharedView>
  <n-modal
    :show="Boolean(selected)"
    preset="card"
    title="本地模拟授权"
    style="width: min(560px, 90vw)"
    :closable="!busy"
    :mask-closable="!busy"
    :close-on-esc="!busy"
    @update:show="closeAuthorization"
  >
    <n-alert type="warning" :show-icon="false">
      余额足够即可本地授权，不等待下游同意。Webhook 协议尚未补齐，本轮不会发送通知。
    </n-alert>
    <CardAuthorizationForm
      v-if="selected"
      :key="orderID"
      :cards="simulationCards"
      :authorize="authorize"
      :select-card="false"
      :merchant-details="false"
      @busy="busy = $event"
      @completed="completeAuthorization"
    />
  </n-modal>
</template>
