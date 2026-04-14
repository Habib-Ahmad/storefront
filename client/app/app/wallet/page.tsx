"use client";

import { useMemo, useState } from "react";
import { ArrowRightIcon, SpinnerGapIcon } from "@phosphor-icons/react";
import { WalletCoinsSvg } from "@/components/illustrations";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useTransactions, useWallet } from "@/hooks/use-wallet";

function formatCurrency(amount: string) {
  return new Intl.NumberFormat("en-NG", {
    style: "currency",
    currency: "NGN",
    minimumFractionDigits: 0,
  }).format(Number(amount || 0));
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("en-NG", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
    hour12: true,
    timeZone: "Africa/Lagos",
  }).format(new Date(value));
}

function transactionBadgeVariant(type: string): "default" | "secondary" | "destructive" {
  if (type === "debit" || type === "refund" || type === "commission") {
    return "destructive";
  }
  if (type === "release" || type === "credit") {
    return "default";
  }
  return "secondary";
}

export default function WalletPage() {
  const [page, setPage] = useState(1);
  const perPage = 10;
  const { data: wallet, isLoading: walletLoading } = useWallet();
  const { data: transactionsResponse, isLoading: transactionsLoading } = useTransactions({
    page,
    per_page: perPage,
  });

  const transactions = transactionsResponse?.data ?? [];
  const totalPages = Math.max(1, Math.ceil((transactionsResponse?.total ?? 0) / perPage));
  const totalBalance = useMemo(() => {
    const available = Number(wallet?.available_balance ?? 0);
    const pending = Number(wallet?.pending_balance ?? 0);
    return (available + pending).toString();
  }, [wallet?.available_balance, wallet?.pending_balance]);

  if (walletLoading || transactionsLoading) {
    return (
      <div className="card-3d flex min-h-80 flex-col items-center justify-center gap-3 rounded-2xl p-8 text-center">
        <SpinnerGapIcon className="size-5 animate-spin text-primary" />
        <p className="text-sm text-muted-foreground">Loading wallet</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Wallet</h1>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <div className="card-3d space-y-1 rounded-2xl p-5">
          <p className="text-sm text-muted-foreground">Available balance</p>
          <p className="text-3xl font-bold">{formatCurrency(wallet?.available_balance ?? "0")}</p>
        </div>
        <div className="card-3d space-y-1 rounded-2xl p-5">
          <p className="text-sm text-muted-foreground">Pending balance</p>
          <p className="text-3xl font-bold">{formatCurrency(wallet?.pending_balance ?? "0")}</p>
        </div>
        <div className="card-3d space-y-1 rounded-2xl p-5">
          <p className="text-sm text-muted-foreground">Total tracked balance</p>
          <p className="text-3xl font-bold">{formatCurrency(totalBalance)}</p>
        </div>
      </div>

      {transactions.length === 0 ? (
        <div className="card-3d flex flex-col items-center justify-center rounded-2xl p-8 text-center">
          <WalletCoinsSvg className="size-36" />
          <p className="mt-3 text-sm text-muted-foreground">
            Transactions will appear here as sales, releases, and refunds happen.
          </p>
        </div>
      ) : (
        <div className="card-3d space-y-4 rounded-2xl p-5">
          <div className="flex items-center justify-between gap-3">
            <div>
              <h2 className="text-base font-semibold">Recent transactions</h2>
              <p className="text-sm text-muted-foreground">
                Ledger activity for credits, releases, commissions, and refunds.
              </p>
            </div>
            <Badge variant="secondary" className="text-xs">
              {transactionsResponse?.total ?? transactions.length} total
            </Badge>
          </div>

          <div className="space-y-3">
            {transactions.map((transaction) => (
              <div
                key={transaction.id}
                className="rounded-xl border border-border/60 bg-background/50 p-4"
              >
                <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div className="space-y-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <Badge
                        variant={transactionBadgeVariant(transaction.type)}
                        className="text-xs capitalize"
                      >
                        {transaction.type.replaceAll("_", " ")}
                      </Badge>
                      {transaction.order_id ? (
                        <span className="text-xs text-muted-foreground">
                          Order {transaction.order_id}
                        </span>
                      ) : null}
                    </div>
                    <p className="text-sm text-muted-foreground">
                      {formatDateTime(transaction.created_at)}
                    </p>
                  </div>

                  <div className="text-left sm:text-right">
                    <p className="text-base font-semibold text-primary">
                      {formatCurrency(transaction.amount)}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      Running balance {formatCurrency(transaction.running_balance)}
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </div>

          {totalPages > 1 ? (
            <div className="flex items-center justify-between gap-3 border-t border-border/60 pt-4">
              <p className="text-sm text-muted-foreground">
                Page {page} of {totalPages}
              </p>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage((value) => value - 1)}
                  disabled={page === 1}
                >
                  Previous
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage((value) => value + 1)}
                  disabled={page >= totalPages}
                >
                  Next
                  <ArrowRightIcon className="size-4" />
                </Button>
              </div>
            </div>
          ) : null}
        </div>
      )}
    </div>
  );
}
