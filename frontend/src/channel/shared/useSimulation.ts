import { ref } from "vue";
import { useMessage } from "naive-ui";

export function useSimulation() {
  const message = useMessage();
  const submitting = ref(false);
  const result = ref("");

  async function submit(action: () => Promise<unknown>, successMessage: string) {
    if (submitting.value) {
      return false;
    }
    submitting.value = true;
    result.value = "";
    try {
      await action();
      result.value = successMessage;
      message.success(successMessage);
      return true;
    } catch (error) {
      message.error(error instanceof Error ? error.message : "模拟交易失败");
      return false;
    } finally {
      submitting.value = false;
    }
  }

  return { submitting, result, submit };
}
