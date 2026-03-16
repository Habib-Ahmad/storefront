export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="relative flex items-center justify-center min-h-screen px-4 bg-dots">
      <div className="fixed inset-0 pointer-events-none" aria-hidden="true">
        <div className="absolute -top-32 left-1/2 -translate-x-1/2 h-80 w-80 rounded-full bg-primary/10 blur-3xl" />
        <div className="absolute bottom-0 -right-20 h-72 w-72 rounded-full bg-primary/8 blur-3xl" />
        <div className="absolute bottom-1/3 -left-16 h-48 w-48 rounded-full bg-primary/6 blur-3xl" />
      </div>
      <div className="relative w-full max-w-sm">{children}</div>
    </div>
  );
}
