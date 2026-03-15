import Link from "next/link";
import { Wallet, Settings } from "lucide-react";

const items = [
  { href: "/app/wallet", label: "Wallet", icon: Wallet },
  { href: "/app/settings", label: "Settings", icon: Settings },
];

export default function MorePage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">More</h1>
      <div className="space-y-2">
        {items.map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className="flex items-center gap-3 rounded-md border p-4"
          >
            <item.icon className="w-5 h-5 text-muted-foreground" />
            <span className="text-sm font-medium">{item.label}</span>
          </Link>
        ))}
      </div>
    </div>
  );
}
