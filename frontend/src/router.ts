import { createRouter, createWebHistory } from "vue-router";

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/slash", redirect: "/slash/cardholders" },
    {
      path: "/slash/cardholders",
      component: () => import("@/channel/slash/Console.vue"),
    },
    {
      path: "/slash/cards",
      component: () => import("@/channel/slash/Console.vue"),
    },
    {
      path: "/slash/transactions",
      component: () => import("@/channel/slash/Console.vue"),
    },
    {
      path: "/slash/authorization",
      component: () => import("@/channel/slash/Console.vue"),
    },
    { path: "/photonpay", redirect: "/photonpay/cardholders" },
    {
      path: "/photonpay/cardholders",
      component: () => import("@/channel/photonpay/Console.vue"),
    },
    {
      path: "/photonpay/cards",
      component: () => import("@/channel/photonpay/Console.vue"),
    },
    {
      path: "/photonpay/transactions",
      component: () => import("@/channel/photonpay/Console.vue"),
    },
    {
      path: "/photonpay/authorization",
      component: () => import("@/channel/photonpay/Console.vue"),
    },
    { path: "/paynda", redirect: "/paynda/cardholders" },
    {
      path: "/paynda/cardholders",
      component: () => import("@/channel/paynda/Console.vue"),
    },
    {
      path: "/paynda/cards",
      component: () => import("@/channel/paynda/Console.vue"),
    },
    {
      path: "/paynda/transactions",
      component: () => import("@/channel/paynda/Console.vue"),
    },
    {
      path: "/paynda/authorization",
      component: () => import("@/channel/paynda/Console.vue"),
    },
    { path: "/", redirect: "/slash/cardholders" },
  ],
});
