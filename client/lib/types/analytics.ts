import * as z from "zod";
import { PaymentMethodSchema, TrackingResponseSchema } from "./orders";

export const AnalyticsPaymentMethodBreakdownSchema = z.object({
  method: PaymentMethodSchema,
  revenue: z.string(),
  count: z.number().int(),
});

export const AnalyticsTopProductSchema = z.object({
  product_name: z.string(),
  quantity_sold: z.number().int(),
  revenue: z.string(),
});

export const AnalyticsPeriodSchema = z.object({
  from: z.string(),
  to: z.string(),
});

export const AnalyticsSummarySchema = z.object({
  total_revenue: z.string(),
  total_cost: z.string(),
  total_profit: z.string(),
  order_count: z.number().int(),
  avg_order_value: z.string(),
  by_payment_method: z.array(AnalyticsPaymentMethodBreakdownSchema),
  top_products: z.array(AnalyticsTopProductSchema),
  period: AnalyticsPeriodSchema,
});

export type TrackingResponse = z.infer<typeof TrackingResponseSchema>;
export type AnalyticsPaymentMethodBreakdown = z.infer<typeof AnalyticsPaymentMethodBreakdownSchema>;
export type AnalyticsTopProduct = z.infer<typeof AnalyticsTopProductSchema>;
export type AnalyticsPeriod = z.infer<typeof AnalyticsPeriodSchema>;
export type AnalyticsSummary = z.infer<typeof AnalyticsSummarySchema>;
