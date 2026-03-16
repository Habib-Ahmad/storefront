"use client";

import { WalletCoinsSvg } from "@/components/illustrations";

export default function WalletPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Wallet</h1>
      </div>

      <div className="card-3d rounded-2xl p-5 space-y-1">
        <p className="text-sm text-muted-foreground">Available Balance</p>
        <p className="text-3xl font-bold">₦0.00</p>
      </div>

      <div className="card-3d rounded-2xl p-8 flex flex-col items-center justify-center text-center">
        <WalletCoinsSvg className="size-36" />
        <p className="text-sm text-muted-foreground mt-3">
          Transactions will appear here as you make sales
        </p>
      </div>
    </div>
  );
}
