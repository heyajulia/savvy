import { z } from "zod";

const ZodApiResponse = z.object({
  Prices: z.object({
    price: z.number(),
    readingDate: z.string(),
  }).array(),
  intervalType: z.number(),
  average: z.number(),
  fromDate: z.string(),
  tillDate: z.string(),
});

const ZodPrices = z.object({
  prices: z.tuple([z.number(), z.number()]).array(),
  average: z.number(),
  date: z.string(),
});

export type ApiResponse = z.infer<typeof ZodApiResponse>;
export type Prices = z.infer<typeof ZodPrices>;

export { ZodApiResponse, ZodPrices };
