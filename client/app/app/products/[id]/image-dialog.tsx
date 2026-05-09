"use client";

import { SpinnerGapIcon, UploadSimpleIcon } from "@phosphor-icons/react";
import { useEffect, useState } from "react";
import { ApiError } from "@/lib/api";
import { productImageSizeMessage } from "@/lib/product-image";
import { uploadImageFile } from "@/lib/media-upload";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { ImageCropDialog } from "../image-crop-dialog";

interface ImageDialogProps {
  open: boolean;
  onClose: () => void;
  productId: string;
  nextSortOrder: number;
  onSubmit: (data: { url: string; sort_order: number; is_primary: boolean }) => Promise<void>;
}

export function ImageDialog({
  open,
  onClose,
  productId,
  nextSortOrder,
  onSubmit,
}: ImageDialogProps) {
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const [isPrimary, setIsPrimary] = useState(nextSortOrder === 0);
  const [uploading, setUploading] = useState(false);
  const [dialogError, setDialogError] = useState<string | null>(null);
  const [fileToCrop, setFileToCrop] = useState<File | null>(null);

  useEffect(() => {
    if (!open) {
      setFile(null);
      if (preview) URL.revokeObjectURL(preview);
      setPreview(null);
      setDialogError(null);
      setFileToCrop(null);
    }
    setIsPrimary(nextSortOrder === 0);
  }, [open, nextSortOrder, preview]);

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const nextFile = event.target.files?.[0];
    if (!nextFile) return;
    setDialogError(null);
    setFileToCrop(nextFile);
    event.target.value = "";
  };

  function handleCropConfirm(nextFile: File) {
    if (preview) URL.revokeObjectURL(preview);
    setFile(nextFile);
    setPreview(URL.createObjectURL(nextFile));
    setFileToCrop(null);
  }

  const handleSubmit = async () => {
    if (!file) return;
    setUploading(true);
    setDialogError(null);
    try {
      const url = await uploadImageFile(file, productId);
      await onSubmit({ url, sort_order: nextSortOrder, is_primary: isPrimary });
    } catch (err) {
      if (err instanceof ApiError) setDialogError(err.message);
      else if (err instanceof Error) setDialogError(err.message);
    } finally {
      setUploading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={(nextOpen: boolean) => !nextOpen && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add image</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          {dialogError && (
            <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {dialogError}
            </p>
          )}
          <label className="block cursor-pointer">
            <div
              className={cn(
                "flex aspect-video flex-col items-center justify-center gap-2 rounded-xl border-2 border-dashed text-muted-foreground transition-colors",
                preview
                  ? "overflow-hidden border-transparent p-0"
                  : "border-border p-6 hover:border-primary/50",
              )}
            >
              {preview ? (
                <img src={preview} alt="" className="size-full object-cover" />
              ) : (
                <>
                  <UploadSimpleIcon className="size-8 opacity-40" />
                  <p className="text-sm">Tap to choose a photo</p>
                </>
              )}
            </div>
            <input type="file" accept="image/*" className="sr-only" onChange={handleFileChange} />
          </label>
          <p className="text-xs text-muted-foreground">
            We crop uploads to match the product card and {productImageSizeMessage().toLowerCase()}
          </p>
          {preview && (
            <button
              type="button"
              className="text-xs text-muted-foreground underline"
              onClick={() => {
                URL.revokeObjectURL(preview);
                setFile(null);
                setPreview(null);
              }}
            >
              Choose a different image
            </button>
          )}
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="img-primary"
              className="size-4 accent-primary"
              checked={isPrimary}
              onChange={(event) => setIsPrimary(event.target.checked)}
            />
            <Label htmlFor="img-primary" className="text-sm font-normal">
              Set as primary image
            </Label>
          </div>
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={!file || uploading}>
            {uploading && <SpinnerGapIcon className="size-4 animate-spin" />}
            {uploading ? "Uploading…" : "Add image"}
          </Button>
        </DialogFooter>

        <ImageCropDialog
          open={!!fileToCrop}
          file={fileToCrop}
          onClose={() => setFileToCrop(null)}
          onConfirm={handleCropConfirm}
        />
      </DialogContent>
    </Dialog>
  );
}
