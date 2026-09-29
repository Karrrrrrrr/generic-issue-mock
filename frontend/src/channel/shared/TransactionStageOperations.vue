<script setup lang="ts">
import TransactionStageForm from "@/channel/shared/TransactionStageForm.vue";
import { simulationStages, type SimulationStage } from "@/channel/shared/simulation";

defineProps<{
  currency: string;
  disabled?: boolean;
  initialAmount?: number;
  operationKey: string | number;
  simulate: (stage: SimulationStage, amount: number) => Promise<unknown>;
}>();

const emit = defineEmits<{
  completed: [];
  busy: [value: boolean];
}>();
</script>

<template>
  <section class="authorization-section">
    <div class="authorization-section-heading">
      <h2>创建后续交易</h2>
    </div>
    <p class="authorization-hint">每次操作生成一笔独立交易，金额必须为正数。创建后自动刷新汇总与交易记录。</p>
    <div class="authorization-operations">
      <div
        v-for="operation in simulationStages"
        :key="operation.key"
        class="authorization-operation"
      >
        <h3>{{ operation.title }}</h3>
        <TransactionStageForm
          :key="`${operationKey}-${operation.key}`"
          :stage="operation.key"
          :currency="currency"
          :initial-amount="initialAmount"
          :disabled="disabled"
          :simulate="(amount) => simulate(operation.key, amount)"
          @busy="emit('busy', $event)"
          @completed="emit('completed')"
        />
      </div>
    </div>
  </section>
</template>
