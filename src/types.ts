import { z } from "zod";

const ZApiResponse = z.object({
  Prices: z.object({
    price: z.number(),
    readingDate: z.string(),
  }).array(),
  intervalType: z.number(),
  average: z.number(),
  fromDate: z.string(),
  tillDate: z.string(),
});

const ZPrices = z.object({
  prices: z.object({
    hour: z.number(),
    price: z.number(),
  }).array(),
  average: z.number(),
  date: z.string(),
});

export type ApiResponse = z.infer<typeof ZApiResponse>;
export type Prices = z.infer<typeof ZPrices>;

export { ZApiResponse, ZPrices };
