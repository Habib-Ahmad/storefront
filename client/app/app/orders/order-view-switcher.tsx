import type { OrderListView } from "@/lib/types/orders";

type OrderViewOption = {
  value: OrderListView;
  label: string;
  description: string;
};

interface OrderViewSwitcherProps {
  value: OrderListView;
  onChange: (value: OrderListView) => void;
  options: readonly OrderViewOption[];
}

export function OrderViewSwitcher({ value, onChange, options }: OrderViewSwitcherProps) {
  return (
    <div className="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
      {options.map((option) => {
        const active = option.value === value;

        return (
          <button
            key={option.value}
            type="button"
            onClick={() => onChange(option.value)}
            className={`rounded-2xl border px-4 py-3 text-left transition-colors ${
              active
                ? "border-foreground bg-foreground text-background"
                : "border-border/60 bg-card text-foreground hover:border-foreground/20"
            }`}
          >
            <p className="text-sm font-semibold">{option.label}</p>
            <p
              className={`mt-1 text-xs leading-5 ${active ? "text-background/75" : "text-muted-foreground"}`}
            >
              {option.description}
            </p>
          </button>
        );
      })}
    </div>
  );
}
