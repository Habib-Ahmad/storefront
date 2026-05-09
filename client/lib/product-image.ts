const PRODUCT_IMAGE_UPLOAD_LIMIT_BYTES = 20 * 1024 * 1024;
const PRODUCT_IMAGE_MAX_DIMENSION = 1800;

export const PRODUCT_IMAGE_UPLOAD_LIMIT_MB = 20;

export function productImageSizeMessage() {
  return `Use an image under ${PRODUCT_IMAGE_UPLOAD_LIMIT_MB} MB.`;
}

function replaceExtension(filename: string, nextExtension: string) {
  return filename.replace(/\.[^.]+$/, "") + nextExtension;
}

function loadImage(file: File) {
  return new Promise<HTMLImageElement>((resolve, reject) => {
    const objectURL = URL.createObjectURL(file);
    const image = new Image();

    image.onload = () => {
      URL.revokeObjectURL(objectURL);
      resolve(image);
    };
    image.onerror = () => {
      URL.revokeObjectURL(objectURL);
      reject(new Error("We couldn't read that image. Try a different file."));
    };

    image.src = objectURL;
  });
}

export async function createProductUploadImage(
  sourceFile: File,
  crop: { x: number; y: number; width: number; height: number },
) {
  const image = await loadImage(sourceFile);
  const largestSide = Math.max(crop.width, crop.height);
  const scale =
    largestSide > PRODUCT_IMAGE_MAX_DIMENSION ? PRODUCT_IMAGE_MAX_DIMENSION / largestSide : 1;

  const canvas = document.createElement("canvas");
  canvas.width = Math.max(1, Math.round(crop.width * scale));
  canvas.height = Math.max(1, Math.round(crop.height * scale));

  const context = canvas.getContext("2d");
  if (!context) {
    throw new Error("Image editing is unavailable in this browser.");
  }

  context.drawImage(
    image,
    crop.x,
    crop.y,
    crop.width,
    crop.height,
    0,
    0,
    canvas.width,
    canvas.height,
  );

  const blob = await new Promise<Blob>((resolve, reject) => {
    canvas.toBlob(
      (nextBlob) => {
        if (!nextBlob) {
          reject(new Error("We couldn't prepare that image for upload."));
          return;
        }
        resolve(nextBlob);
      },
      "image/webp",
      0.9,
    );
  });

  const processedFile = new File([blob], replaceExtension(sourceFile.name, ".webp"), {
    type: "image/webp",
  });

  if (processedFile.size > PRODUCT_IMAGE_UPLOAD_LIMIT_BYTES) {
    throw new Error(`This image is still too large after editing. ${productImageSizeMessage()}`);
  }

  return processedFile;
}
