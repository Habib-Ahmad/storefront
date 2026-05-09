"use client";

import Cropper from "react-easy-crop";
import "react-easy-crop/react-easy-crop.css";
import { useEffect, useState } from "react";
import type { Area } from "react-easy-crop";
import { SpinnerGapIcon } from "@phosphor-icons/react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { createProductUploadImage, productImageSizeMessage } from "@/lib/product-image";

interface ImageCropDialogProps {
  open: boolean;
  file: File | null;
  onClose: () => void;
  onConfirm: (file: File) => void;
}

export function ImageCropDialog({ open, file, onClose, onConfirm }: ImageCropDialogProps) {
  const [crop, setCrop] = useState({ x: 0, y: 0 });
  const [zoom, setZoom] = useState(1);
  const [croppedAreaPixels, setCroppedAreaPixels] = useState<Area | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [dialogError, setDialogError] = useState<string | null>(null);
  const [previewURL, setPreviewURL] = useState<string | null>(null);

  useEffect(() => {
    if (!file) {
      setPreviewURL(null);
      return;
    }

    const objectURL = URL.createObjectURL(file);
    setPreviewURL(objectURL);

    return () => URL.revokeObjectURL(objectURL);
  }, [file]);

  function handleOpenChange(nextOpen: boolean) {
    if (!nextOpen) {
      setCrop({ x: 0, y: 0 });
      setZoom(1);
      setCroppedAreaPixels(null);
      setDialogError(null);
      onClose();
    }
  }

  async function handleConfirm() {
    if (!file || !croppedAreaPixels) {
      return;
    }

    setSubmitting(true);
    setDialogError(null);
    try {
      const processedFile = await createProductUploadImage(file, croppedAreaPixels);
      onConfirm(processedFile);
    } catch (error) {
      setDialogError(error instanceof Error ? error.message : "We couldn't prepare that image.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-3xl">
        <DialogHeader>
          <DialogTitle>Crop product image</DialogTitle>
          <DialogDescription>
            Reposition the image so the product card stays sharp and intentional.{" "}
            {productImageSizeMessage()}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {dialogError ? (
            <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {dialogError}
            </p>
          ) : null}

          <div className="relative h-105 overflow-hidden rounded-2xl bg-black/90">
            {previewURL ? (
              <Cropper
                image={previewURL}
                crop={crop}
                zoom={zoom}
                aspect={4 / 5}
                cropShape="rect"
                showGrid={false}
                objectFit="contain"
                onCropChange={setCrop}
                onZoomChange={setZoom}
                onCropComplete={(_, nextCroppedAreaPixels) => {
                  setCroppedAreaPixels(nextCroppedAreaPixels);
                }}
              />
            ) : null}
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="font-medium">Zoom</span>
              <span className="text-muted-foreground">{zoom.toFixed(1)}x</span>
            </div>
            <input
              type="range"
              min={1}
              max={3}
              step={0.1}
              value={zoom}
              onChange={(event) => setZoom(Number(event.target.value))}
              className="w-full accent-primary"
            />
            <p className="text-xs text-muted-foreground">
              Drag to reposition. We export an optimized image to reduce upload friction for most
              people.
            </p>
          </div>

          <p className="text-xs text-muted-foreground">{productImageSizeMessage()}</p>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>
            Cancel
          </Button>
          <Button type="button" onClick={() => void handleConfirm()} disabled={!file || submitting}>
            {submitting ? <SpinnerGapIcon className="size-4 animate-spin" /> : null}
            {submitting ? "Preparing…" : "Use cropped image"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
