export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="bg-dots relative flex min-h-screen items-center justify-center px-4 py-8 md:px-6">
      <div className="pointer-events-none fixed inset-0" aria-hidden="true">
        <div className="absolute -top-32 left-1/2 h-80 w-80 -translate-x-1/2 rounded-full bg-primary/10 blur-3xl" />
        <div className="absolute -right-20 bottom-0 h-72 w-72 rounded-full bg-primary/8 blur-3xl" />
        <div className="absolute bottom-1/3 -left-16 h-48 w-48 rounded-full bg-primary/6 blur-3xl" />
      </div>
      <div className="relative w-full max-w-sm md:max-w-md lg:max-w-lg">{children}</div>
    </div>
  );
}
