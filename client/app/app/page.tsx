export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Overview</h1>
      <div className="grid gap-4 grid-cols-2 lg:grid-cols-4">
        {["Revenue", "Orders", "Profit", "Avg. Order"].map((label) => (
          <div
            key={label}
            className="rounded-lg border p-4 space-y-1"
          >
            <p className="text-sm text-muted-foreground">{label}</p>
            <p className="text-2xl font-bold">—</p>
          </div>
        ))}
      </div>
    </div>
  );
}
