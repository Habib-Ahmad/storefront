interface Props {
  params: Promise<{ slug: string }>;
}

export default async function TrackPage({ params }: Props) {
  const { slug } = await params;

  return (
    <div className="flex min-h-screen flex-col items-center justify-center px-4">
      <h1 className="text-2xl font-bold">Track Order</h1>
      <p className="mt-2 text-muted-foreground">Tracking: {slug}</p>
    </div>
  );
}
