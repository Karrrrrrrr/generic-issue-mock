<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { NButton, NForm, NFormItem, NInputNumber, useMessage } from "naive-ui";
import { isPositiveSimulationAmount, simulationStages, type SimulationStage } from "./simulation";
import { useSimulation } from "./useSimulation";

const props = defineProps<{
  stage: SimulationStage;
  currency: string;
  simulate: (amount: number) => Promise<unknown>;
  initialAmount?: number;
  disabled?: boolean;
}>();
const emit = defineEmits<{
  completed: [];
  busy: [value: boolean];
}>();
const message = useMessage();
const { submitting, submit } = useSimulation();
const amount = ref<number | null>(props.initialAmount ?? null);
const operation = computed(() => simulationStages.find((item) => item.key === props.stage)!);
watch(submitting, (busy) => emit("busy", busy), { flush: "sync" });

async function submitStage() {
  if (props.disabled || submitting.value) {
    return;
  }
  if (!isPositiveSimulationAmount(amount.value)) {
    message.warning("金额必须是有限的正数");
    return;
  }
  const submittedAmount = amount.value;
  if (await submit(() => props.simulate(submittedAmount), `已创建${operation.value.title}交易`)) {
    amount.value = null;
    emit("completed");
  }
}
</script>

<template>
  <n-form
    class="simulation-stage-form"
    label-placement="top"
    :disabled="disabled || submitting"
    @submit.prevent="submitStage"
  >
    <p class="authorization-hint">{{ operation.description }}</p>
    <n-form-item :label="`${operation.title}金额（${currency}）`" required>
      <n-input-number
        v-model:value="amount"
        :placeholder="`输入${operation.title}金额`"
        :show-button="false"
        style="width: 100%"
      />
    </n-form-item>
    <n-button
      attr-type="submit"
      :type="operation.type"
      :loading="submitting"
      :disabled="disabled || submitting || !isPositiveSimulationAmount(amount)"
      ghost
    >
      创建{{ operation.title }}
    </n-button>
  </n-form>
</template>

<style scoped>
.simulation-stage-form {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 12px;
}

.simulation-stage-form .authorization-hint {
  margin: 0;
}
</style>
