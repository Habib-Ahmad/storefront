interface Props {
  params: Promise<{ slug: string }>;
}

export default async function StorefrontPage({ params }: Props) {
  const { slug } = await params;

  return (
    <div className="flex min-h-screen flex-col items-center justify-center px-4">
      <h1 className="text-3xl font-bold">{slug}</h1>
      <p className="mt-2 text-muted-foreground">Public storefront — coming soon</p>
    </div>
  );
}
