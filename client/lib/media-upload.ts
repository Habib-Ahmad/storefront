import { api } from "@/lib/api";

export async function uploadImageFile(file: File, productId?: string) {
  const contentType = file.type || "application/octet-stream";
  const { upload_url, public_url } = await api.getUploadUrl({
    filename: file.name,
    content_type: contentType,
    product_id: productId,
  });

  let uploadResponse: Response;
  try {
    uploadResponse = await fetch(upload_url, {
      method: "PUT",
      headers: {
        "Content-Type": contentType,
      },
      body: file,
    });
  } catch {
    const origin = typeof window !== "undefined" ? window.location.origin : "your app origin";
    throw new Error(
      `Upload request was blocked before reaching R2. Configure bucket CORS to allow ${origin} with PUT and the Content-Type header.`,
    );
  }

  if (!uploadResponse.ok) {
    throw new Error(`R2 rejected the upload with status ${uploadResponse.status}`);
  }

  return public_url;
}
