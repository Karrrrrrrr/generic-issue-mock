export interface SimulationCard {
  id: number;
  label: string;
  currency: string;
  disabled: boolean;
}

export interface AuthorizationSimulationRequest {
  cardID: number;
  amount: number;
  currency: string;
  merchantName: string;
  merchantMCC: string;
  merchantCountry: string;
  merchantCity: string;
}

export interface RefundSimulationRequest {
  authorization_id?: number;
  card_id?: number;
  amount: number;
  currency?: string;
  merchant_name: string;
  merchant_category_code: string;
  merchant_country: string;
  merchant_city: string;
}

export type SimulationStage = "clear" | "reverse" | "refund";

export const simulationStages = [
  {
    key: "clear",
    title: "清算",
    type: "primary",
    description: "释放对应冻结并扣款，允许超额及继续清算。",
  },
  {
    key: "reverse",
    title: "撤销",
    type: "warning",
    description: "释放对应冻结回可用余额，不产生钱包收支。",
  },
  {
    key: "refund",
    title: "退款",
    type: "info",
    description: "直接退回可用余额，无需先清算，不恢复授权额度。",
  },
] as const;

export function isPositiveSimulationAmount(amount: number | null): amount is number {
  return amount !== null && Number.isFinite(amount) && amount > 0;
}
