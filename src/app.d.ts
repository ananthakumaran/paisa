/// <reference types="@sveltejs/kit" />

interface UserConfig {
  default_currency: string;
  locale: string;
  journal_path: string;
}

// eslint-disable-next-line no-var
declare var USER_CONFIG: UserConfig;

declare namespace App {
  // interface Error {}
  // interface Locals {}
  // interface PageData {}
  // interface Platform {}
}

declare module "arima/async" {
  export class Arima {
    constructor(options: object);
    train(points: number[]): Arima;
    predict(count: number): [number[], number[]];
  }
  const P: Promise<typeof Arima>;
  export default P;
}

declare module "d3-sankey-circular" {
  export function sankeyCircular(): any;
  export function sankeyJustify(): any;
}

declare module "d3-path-arrows" {
  export function pathArrows(): any;
}

declare module "compute-cosine-similarity" {
  export default function similarity(a: number[], b: number[]): number;
}
