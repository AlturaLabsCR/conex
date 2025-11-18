import {
  loadScript,
  PayPalNamespace,
  PayPalButtonsComponentOptions
} from "@paypal/paypal-js";

let paypal: PayPalNamespace | null = null;

export async function initPayPalButtonsPurchase(
  clientId: string,
  selector: string = "#paypal-buttons",
  createOrderUrl = "/checkout/create",
  completeOrderUrl = "/checkout/complete",
  redirectOnApproveUrl = "/dashboard",

) {
  const paypalButtonsOpts: PayPalButtonsComponentOptions = {
    style: {
      layout: "vertical",
      color: "black",
      shape: "pill",
      label: "pay",
    },

    //
    // https://developer.paypal.com/docs/api/orders/v2/#orders_create
    //
    createOrder: async () => {
      const csrfToken =
        document.cookie
          .split("; ")
          .find((c) => c.startsWith("csrf="))
          ?.split("=")[1] || "";

      const response = await fetch(createOrderUrl, {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": csrfToken ? csrfToken : "",
        },
      });

      const body = await response.json();
      return body.id;
    },

    //
    // https://developer.paypal.com/docs/api/orders/v2/#orders_capture
    //
    onApprove: async (data) => {
      try {
        const el = document.querySelector(selector);
        if (el) el.innerHTML = '<h1 class="text-center" aria-busy="true"></h1>';
        const csrfToken =
          document.cookie
            .split("; ")
            .find((c) => c.startsWith("csrf="))
            ?.split("=")[1] || "";

        await fetch(completeOrderUrl, {
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
            "X-CSRF-Token": csrfToken ? csrfToken : "",
          },
          body: JSON.stringify({
            order_id: data.orderID,
          }),
        });

        window.location.href = redirectOnApproveUrl;
      } catch (error) {
        console.error("Error completing order:", error);
      }
    },

    onCancel: () => {
      alert("Se ha cancelado el pago");
    },

    onError: (err) => {
      console.error("PayPal Button Error:", err);
    },
  };

  try {
    const intent = "capture";

    paypal = await loadScript({
      clientId,
      currency: "USD",
      components: ["buttons"],
      intent,
      vault: false,
    });
  } catch (err) {
    console.error("Failed to load PayPal JS SDK:", err);
    return;
  }

  if (!paypal?.Buttons) {
    console.error("PayPal Buttons not available.");
    return;
  }

  try {
    await paypal.Buttons(paypalButtonsOpts).render(selector);
  } catch (err) {
    console.error("Failed to render PayPal Buttons:", err);
  }
}
(window as any).initPayPalButtonsPurchase = initPayPalButtonsPurchase;

export async function initPayPalButtonsSubscription(
  clientId: string,
  selector: string = "#paypal-buttons",
  createSubscriptionUrl = "/checkout/create",
  completeSubscriptionUrl = "/checkout/complete",
) {
  const paypalButtonsOpts: PayPalButtonsComponentOptions = {
    style: {
      layout: "vertical",
      color: "black",
      shape: "pill",
      label: "pay",
    },

    //
    // https://developer.paypal.com/docs/api/orders/v2/#orders_create
    //
    createSubscription: async () => {
      const csrfToken =
        document.cookie
          .split("; ")
          .find((c) => c.startsWith("csrf="))
          ?.split("=")[1] || "";

      const response = await fetch(createSubscriptionUrl, {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": csrfToken ? csrfToken : "",
        },
      });

      const body = await response.json();
      return body.id;
    },

    //
    // https://developer.paypal.com/docs/api/orders/v2/#orders_capture
    //
    onApprove: async (data) => {
      try {
        const csrfToken =
          document.cookie
            .split("; ")
            .find((c) => c.startsWith("csrf="))
            ?.split("=")[1] || "";

        const response = await fetch(completeSubscriptionUrl, {
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
            "X-CSRF-Token": csrfToken ? csrfToken : "",
          },
          body: JSON.stringify({
            subscription_id: data.subscriptionID,
          }),
        });

        const order = await response.json();
        console.log("Payment completed:", order);
      } catch (error) {
        console.error("Error completing order:", error);
        alert("An error occurred while capturing your payment.");
      }
    },

    onCancel: () => {
      alert("Order cancelled.");
    },

    onError: (err) => {
      console.error("PayPal Button Error:", err);
    },
  };

  try {
    const intent = "subscription";

    paypal = await loadScript({
      clientId,
      currency: "USD",
      components: ["buttons"],
      intent,
      vault: true,
    });
  } catch (err) {
    console.error("Failed to load PayPal JS SDK:", err);
    return;
  }

  if (!paypal?.Buttons) {
    console.error("PayPal Buttons not available.");
    return;
  }

  try {
    await paypal.Buttons(paypalButtonsOpts).render(selector);
  } catch (err) {
    console.error("Failed to render PayPal Buttons:", err);
  }
}
(window as any).initPayPalButtonsSubscription = initPayPalButtonsSubscription;
