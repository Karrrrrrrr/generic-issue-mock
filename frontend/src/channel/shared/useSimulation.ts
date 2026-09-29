import { ref } from "vue";
import { useMessage } from "naive-ui";

export function useSimulation() {
  const message = useMessage();
  const submitting = ref(false);
  const result = ref("");

  async function submitData<T>(
    action: () => Promise<T>,
    successMessage: string,
    options: {
      isSuccess?: (data: T) => boolean;
      failureMessage?: (data: T) => string;
    } = {},
  ) {
    if (submitting.value) {
      return undefined;
    }
    submitting.value = true;
    result.value = "";
    try {
      const data = await action();
      const succeeded = options.isSuccess ? options.isSuccess(data) : true;
      if (!succeeded) {
        message.error(options.failureMessage ? options.failureMessage(data) : "模拟交易失败");
        return data;
      }
      result.value = successMessage;
      message.success(successMessage);
      return data;
    } catch (error) {
      message.error(error instanceof Error ? error.message : "模拟交易失败");
      return undefined;
    } finally {
      submitting.value = false;
    }
  }

  async function submit(action: () => Promise<unknown>, successMessage: string) {
    return Boolean(await submitData(
      async () => {
        await action();
        return true;
      },
      successMessage,
    ));
  }

  return {
    submitting,
    result,
    submit,
    submitData,
  };
}
