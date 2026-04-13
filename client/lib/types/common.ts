import * as z from "zod";

export const PaginationParamsSchema = z.object({
  page: z.number().int().positive().optional(),
  per_page: z.number().int().positive().optional(),
});

export const PaginationMetaSchema = z.object({
  total: z.number().int(),
  page: z.number().int(),
  per_page: z.number().int(),
});

export function createPaginatedResponseSchema<TItem extends z.ZodType>(itemSchema: TItem) {
  return PaginationMetaSchema.extend({
    data: z.array(itemSchema),
  });
}

export type PaginationParams = z.infer<typeof PaginationParamsSchema>;
export type PaginationMeta = z.infer<typeof PaginationMetaSchema>;
export type PaginatedResponse<TItem> = PaginationMeta & {
  data: TItem[];
};
