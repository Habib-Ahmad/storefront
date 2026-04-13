"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeftIcon } from "@phosphor-icons/react";
import { Button } from "@/components/ui/button";
import { useCancelOrder, useOrder, useOrderItems, useResumeOrderPayment } from "@/hooks/use-orders";
import { ApiError } from "@/lib/api";
import { OrderCancelDialog } from "./order-cancel-dialog";
import {
  ActionCard,
  DeliveryCard,
  ItemsCard,
  OrderDetailSkeleton,
  OrderSummaryCard,
} from "./order-detail-sections";

export default function OrderDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();

  const { data: order, isLoading: orderLoading } = useOrder(id);
  const { data: items, isLoading: itemsLoading } = useOrderItems(id);
  const cancelOrder = useCancelOrder();
  const resumeOrderPayment = useResumeOrderPayment();

  const [cancelDialogOpen, setCancelDialogOpen] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  const loading = orderLoading || itemsLoading;
  const orderItems = useMemo(() => items ?? [], [items]);

  if (loading) {
    return <OrderDetailSkeleton />;
  }

  if (!order) {
    return (
      <div className="mx-auto max-w-3xl space-y-4">
        <Link href="/app/orders">
          <Button variant="ghost" size="sm" className="-ml-2 gap-1">
            <ArrowLeftIcon className="size-4" />
            Back
          </Button>
        </Link>
        <div className="card-3d rounded-2xl p-6">
          <p className="text-sm text-muted-foreground">Order not found.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl space-y-4">
      <div>
        <Link href="/app/orders">
          <Button variant="ghost" size="sm" className="-ml-2 gap-1">
            <ArrowLeftIcon className="size-4" />
            Back
          </Button>
        </Link>
        <h1 className="mt-1 text-2xl font-bold">Order details</h1>
        <p className="text-sm text-muted-foreground">Review this order and its current status.</p>
      </div>

      <OrderSummaryCard order={order} />
      <DeliveryCard order={order} />
      <ItemsCard items={orderItems} order={order} />
      <ActionCard
        order={order}
        actionError={actionError}
        isResumingPayment={resumeOrderPayment.isPending}
        onCancel={() => {
          setActionError(null);
          setCancelDialogOpen(true);
        }}
        onResumePayment={async () => {
          try {
            setActionError(null);
            const response = await resumeOrderPayment.mutateAsync(order.id);
            window.location.href = response.authorization_url;
          } catch (err) {
            if (err instanceof ApiError) {
              setActionError(err.message);
            } else {
              setActionError("Unable to continue payment");
            }
          }
        }}
      />

      <OrderCancelDialog
        open={cancelDialogOpen}
        onOpenChange={setCancelDialogOpen}
        isPending={cancelOrder.isPending}
        onConfirm={async () => {
          try {
            await cancelOrder.mutateAsync(order.id);
            setCancelDialogOpen(false);
            router.refresh();
          } catch (err) {
            if (err instanceof ApiError) {
              setActionError(err.message);
            } else {
              setActionError("Unable to cancel order");
            }
            setCancelDialogOpen(false);
          }
        }}
      />
    </div>
  );
}
