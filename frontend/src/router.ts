import { createRouter, createWebHistory } from "vue-router";

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/slash",
      component: () => import("@/channel/slash/Shell.vue"),
      children: [
        { path: "", redirect: "/slash/cardholders" },
        { path: "accounts", component: () => import("@/channel/slash/AccountsView.vue") },
        {
          path: "card-products",
          component: () => import("@/channel/slash/CardProductsView.vue"),
        },
        {
          path: "virtual-accounts",
          component: () => import("@/channel/slash/VirtualAccountsView.vue"),
        },
        {
          path: "cardholders",
          component: () => import("@/channel/slash/CardholdersView.vue"),
        },
        {
          path: "cards",
          component: () => import("@/channel/slash/CardsView.vue"),
        },
        {
          path: "transactions",
          component: () => import("@/channel/slash/TransactionsView.vue"),
        },
        {
          path: "authorizations",
          component: () => import("@/channel/slash/AuthorizationsView.vue"),
        },
        {
          path: "authorization",
          component: () => import("@/channel/slash/AuthorizationView.vue"),
        },
        {
          path: "webhooks",
          component: () => import("@/channel/slash/WebhooksView.vue"),
        },
      ],
    },
    {
      path: "/photonpay",
      component: () => import("@/channel/photonpay/Shell.vue"),
      children: [
        { path: "", redirect: "/photonpay/cardholders" },
        {
          path: "virtual-accounts",
          component: () => import("@/channel/photonpay/VirtualAccountsView.vue"),
        },
        { path: "accounts", component: () => import("@/channel/photonpay/AccountsView.vue") },
        {
          path: "cardholders",
          component: () => import("@/channel/photonpay/CardholdersView.vue"),
        },
        {
          path: "cards",
          component: () => import("@/channel/photonpay/CardsView.vue"),
        },
        {
          path: "transactions",
          component: () => import("@/channel/photonpay/TransactionsView.vue"),
        },
        {
          path: "authorizations",
          component: () => import("@/channel/photonpay/AuthorizationsView.vue"),
        },
        {
          path: "authorization",
          component: () => import("@/channel/photonpay/AuthorizationView.vue"),
        },
        { path: "webhooks", component: () => import("@/channel/photonpay/WebhooksView.vue") },
        {
          path: "webhook-records",
          component: () => import("@/channel/photonpay/WebhookRecordsView.vue"),
        },
      ],
    },
    {
      path: "/paynda",
      component: () => import("@/channel/paynda/Shell.vue"),
      children: [
        { path: "", redirect: "/paynda/cardholders" },
        {
          path: "cardholders",
          component: () => import("@/channel/paynda/CardholdersView.vue"),
        },
        {
          path: "cards",
          component: () => import("@/channel/paynda/CardsView.vue"),
        },
        {
          path: "transactions",
          component: () => import("@/channel/paynda/TransactionsView.vue"),
        },
        {
          path: "authorizations",
          component: () => import("@/channel/paynda/AuthorizationsView.vue"),
        },
        {
          path: "authorization",
          component: () => import("@/channel/paynda/AuthorizationView.vue"),
        },
        { path: "accounts", component: () => import("@/channel/paynda/AccountsView.vue") },
        { path: "webhooks", component: () => import("@/channel/paynda/WebhooksView.vue") },
        {
          path: "webhook-records",
          component: () => import("@/channel/paynda/WebhookRecordsView.vue"),
        },
      ],
    },
    { path: "/", redirect: "/slash/cardholders" },
  ],
});
