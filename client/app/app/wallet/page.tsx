"use client";

import { WalletCoinsSvg } from "@/components/illustrations";

export default function WalletPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Wallet</h1>
      </div>

      <div className="card-3d space-y-1 rounded-2xl p-5">
        <p className="text-sm text-muted-foreground">Available Balance</p>
        <p className="text-3xl font-bold">₦0.00</p>
      </div>

      <div className="card-3d flex flex-col items-center justify-center rounded-2xl p-8 text-center">
        <WalletCoinsSvg className="size-36" />
        <p className="mt-3 text-sm text-muted-foreground">
          Transactions will appear here as you make sales
        </p>
      </div>
    </div>
  );
}
