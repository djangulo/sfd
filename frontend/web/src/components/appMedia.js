/** @jsx jsx */
import { jsx, css } from "@emotion/core";

import facepaint from "facepaint";

const breakpoints = [0, 768, 992, 1200];

export const MediaQueries = facepaint(
  breakpoints.map((bp) => `@media (min-width: ${bp}px)`)
);

// export const AppMedia = createMedia({
//   breakpoints: {
//     sm: 0,
//     md: 768,
//     lg: 1024,
//     xl: 1192,
//   },
// });
