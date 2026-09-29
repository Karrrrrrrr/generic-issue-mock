<script setup lang="ts">
import { formatDateTime } from "@/channel/dateTime";
import { onMounted, ref } from "vue";
import {
  createDiscreteApi,
  NButton,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSelect,
  NSpace,
  NSwitch,
} from "naive-ui";
import {
  accountApi,
  authorizationConfigApi,
  type Account,
  type AuthorizationConfig,
} from "./api";

const { message } = createDiscreteApi(["message"]);
const accounts = ref<Account[]>([]);
const selectedAccountId = ref<number | null>(null);
const current = ref<AuthorizationConfig | null>(null);
const loading = ref(false);
const saving = ref(false);
const form = ref({
  target_url: "",
  enabled: true,
  timeout_millis: 500,
});

async function loadAccounts() {
  try {
    accounts.value = await accountApi.listAll();
    if (!selectedAccountId.value && accounts.value.length > 0) {
      selectedAccountId.value = accounts.value[0].id;
    }
  } catch (error) {
    message.error(error instanceof Error ? error.message : "加载账户失败");
  }
}

async function loadConfig() {
  if (!selectedAccountId.value) {
    current.value = null;
    return;
  }

  loading.value = true;
  try {
    const item = await authorizationConfigApi.get(selectedAccountId.value);
    current.value = item;
    form.value = {
      target_url: item.target_url,
      enabled: item.enabled,
      timeout_millis: item.timeout_millis,
    };
  } catch (error) {
    current.value = null;
    form.value = {
      target_url: "",
      enabled: true,
      timeout_millis: 500,
    };
    message.error(error instanceof Error ? error.message : "加载授权配置失败");
  } finally {
    loading.value = false;
  }
}

async function changeAccount(value: number | null) {
  selectedAccountId.value = value;
  await loadConfig();
}

async function save() {
  if (!selectedAccountId.value) {
    message.warning("请选择账户");
    return;
  }
  if (!form.value.target_url) {
    message.warning("请填写授权 URL");
    return;
  }

  saving.value = true;
  try {
    current.value = await authorizationConfigApi.update({
      account_id: selectedAccountId.value,
      target_url: form.value.target_url,
      enabled: form.value.enabled,
      timeout_millis: form.value.timeout_millis,
    });
    message.success("保存成功");
  } catch (error) {
    message.error(error instanceof Error ? error.message : "保存失败");
  } finally {
    saving.value = false;
  }
}

onMounted(async () => {
  await loadAccounts();
  await loadConfig();
});
</script>

<template>
  <section>
    <div class="page-heading">
      <div>
        <h1>授权配置</h1>
        <p>配置 PhotonPay 授权回调地址、启用状态和超时时间。</p>
      </div>
    </div>

    <div class="table-filters photonpay-authorization-config">
      <n-form label-placement="top">
        <div class="photonpay-authorization-config-grid">
          <n-form-item label="账户">
            <n-select
              v-model:value="selectedAccountId"
              :options="
                accounts.map((account) => ({
                  label: `${account.name} (${account.id})`,
                  value: account.id,
                }))
              "
              filterable
              placeholder="选择账户"
              :loading="loading"
              @update:value="changeAccount"
            />
          </n-form-item>
          <n-form-item label="启用">
            <n-switch v-model:value="form.enabled" />
          </n-form-item>
          <n-form-item label="超时毫秒">
            <n-input-number
              v-model:value="form.timeout_millis"
              :min="1"
              :show-button="false"
              placeholder="500"
            />
          </n-form-item>
          <n-form-item
            class="photonpay-authorization-config-url"
            label="授权 URL"
          >
            <n-input
              v-model:value="form.target_url"
              placeholder="http://127.0.0.1:18080/api/v1/notify/ds-authorization"
            />
          </n-form-item>
        </div>
      </n-form>

      <div class="photonpay-authorization-config-footer">
        <span>
          {{
            current?.updated_at
              ? `更新时间 ${formatDateTime(current.updated_at)}`
              : "未加载配置"
          }}
        </span>
        <n-space>
          <n-button
            :disabled="!selectedAccountId"
            :loading="loading"
            @click="loadConfig"
          >
            查询
          </n-button>
          <n-button
            type="primary"
            :disabled="!selectedAccountId"
            :loading="saving"
            @click="save"
          >
            保存
          </n-button>
        </n-space>
      </div>
    </div>
  </section>
</template>

<style scoped>
.photonpay-authorization-config {
  max-width: 920px;
}

.photonpay-authorization-config-grid {
  display: grid;
  grid-template-columns: minmax(220px, 1fr) 96px 160px;
  gap: 0 16px;
  align-items: start;
}

.photonpay-authorization-config-url {
  grid-column: 1 / -1;
}

.photonpay-authorization-config-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.photonpay-authorization-config-footer span {
  color: var(--app-text-muted);
  font-size: 13px;
  transition: var(--theme-color-transition);
}

@media (max-width: 720px) {
  .photonpay-authorization-config-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .photonpay-authorization-config-footer {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
