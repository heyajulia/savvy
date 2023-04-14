import formatRanges, {
  groupIntoRanges,
  sortAscending,
} from "./pretty-range.ts";

import { assertEquals } from "asserts";

// First, let's test that we can sort the values in ascending order.
Deno.test("sorts values in ascending order", () => {
  assertEquals(sortAscending([2, 1, 3]), [1, 2, 3]);
});

// Next, let's test that we can group the values into ranges.
Deno.test("groups values into ranges: 1, 3", () => {
  assertEquals(groupIntoRanges([1, 3]), [1, 3]);
});

Deno.test("groups values into ranges: 1, 2, 3", () => {
  assertEquals(groupIntoRanges([1, 2, 3]), [[1, 3]]);
});

Deno.test("groups values into ranges: 1, 2, 3, 5, 6, 7", () => {
  assertEquals(groupIntoRanges([1, 2, 3, 5, 6, 7]), [[1, 3], [5, 7]]);
});

Deno.test("groups values into ranges: 1, 2, 3, 5, 6, 7, 9, 10, 11", () => {
  assertEquals(
    groupIntoRanges([1, 2, 3, 5, 6, 7, 9, 10, 11]),
    [[1, 3], [5, 7], [9, 11]],
  );
});

Deno.test("groups values into ranges: 1, 2, 4, 6", () => {
  assertEquals(groupIntoRanges([1, 2, 4, 6]), [[1, 2], 4, 6]);
});

Deno.test("groups values into ranges: empty array", () => {
  assertEquals(groupIntoRanges([]), []);
});

Deno.test("groups values into ranges: single element array", () => {
  assertEquals(groupIntoRanges([1]), [1]);
});

Deno.test("groups values into ranges: negative numbers", () => {
  assertEquals(groupIntoRanges([-3, -2, -1]), [[-3, -1]]);
});

Deno.test("groups values into ranges: non-consecutive numbers", () => {
  assertEquals(groupIntoRanges([1, 3, 5]), [1, 3, 5]);
});

Deno.test("groups values into ranges: large gap", () => {
  assertEquals(groupIntoRanges([1, 2, 10, 11]), [[1, 2], [10, 11]]);
});

// Finally, let's test that we can format the ranges into a string.
Deno.test("formats ranges into a string", () => {
  assertEquals(formatRanges([[1, 3]]), "van 01:00 tot 03:59");
});

Deno.test("formats ranges into a string: 5, 7", () => {
  assertEquals(formatRanges([[5, 7]]), "van 05:00 tot 07:59");
});

Deno.test("formats ranges into a string: 1, 3 and 5, 7", () => {
  assertEquals(
    formatRanges([[1, 3], [5, 7]]),
    "van 01:00 tot 03:59 en van 05:00 tot 07:59",
  );
});

Deno.test("formats ranges into a string: 1, 3, 5, 7, 9, 11", () => {
  assertEquals(
    formatRanges([[1, 3], [5, 7], [9, 11]]),
    "van 01:00 tot 03:59, van 05:00 tot 07:59 en van 09:00 tot 11:59",
  );
});

Deno.test("formats ranges into a string: empty array", () => {
  assertEquals(formatRanges([]), "");
});

Deno.test("formats ranges into a string: single element array", () => {
  assertEquals(formatRanges([1]), "van 01:00 tot 01:59");
});

Deno.test("formats ranges into a string: 1, 3", () => {
  assertEquals(
    formatRanges([1, 3]),
    "van 01:00 tot 01:59 en van 03:00 tot 03:59",
  );
});
