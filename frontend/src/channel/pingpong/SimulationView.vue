<script setup lang="ts">
import { NAlert } from "naive-ui";
import CardTransactionSimulator from "@/channel/shared/CardTransactionSimulator.vue";
import type { AuthorizationSimulationRequest, RefundSimulationRequest } from "@/channel/shared/simulation";
import { loadSimulationCards, simulateAuthorization, simulateRefund } from "./management";
import { requestID } from "./api";

let authorizationRequestID = requestID();
let refundRequestID = requestID();

async function authorize(request: AuthorizationSimulationRequest) {
  await simulateAuthorization({ ...request, requestID: authorizationRequestID });
  authorizationRequestID = requestID();
}

async function refund(request: RefundSimulationRequest) {
  await simulateRefund({ ...request, requestID: refundRequestID });
  refundRequestID = requestID();
}
</script>

<template>
  <div class="ping-page">
    <div class="page-heading">
      <div>
        <h1>交易模拟</h1>
        <p>模拟授权、独立退款或关联退款；清算和撤销请到授权管理。</p>
      </div>
    </div>
    <n-alert type="warning" :show-icon="false">
      操作会更新本地钱包和交易记录，并按启用的 Webhook 配置异步通知。投递失败不回滚交易；当前暂不签名，HTTP 200 暂视为成功。
    </n-alert>
    <CardTransactionSimulator
      :load-cards="loadSimulationCards"
      :authorize="authorize"
      :refund="refund"
    />
  </div>
</template>
