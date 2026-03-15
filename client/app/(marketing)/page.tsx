import Link from "next/link";

export default function HomePage() {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen px-4">
      <h1 className="text-4xl font-bold tracking-tight sm:text-6xl">
        Storefront
      </h1>
      <p className="mt-4 text-lg text-muted-foreground text-center max-w-md">
        The retail OS for Nigerian SMEs. Free to start — we only earn when you do.
      </p>
      <div className="mt-8 flex gap-4">
        <Link
          href="/login"
          className="rounded-md bg-primary px-6 py-3 text-sm font-medium text-primary-foreground"
        >
          Get Started
        </Link>
        <Link
          href="/about"
          className="rounded-md border px-6 py-3 text-sm font-medium"
        >
          Learn More
        </Link>
      </div>
    </div>
  );
}
