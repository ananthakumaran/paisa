/// <reference types="@sveltejs/kit" />

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
