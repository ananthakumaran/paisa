declare module "bun" {
  export function spawn(args: string[], options?: any): any;
}
declare module "bun:test" {
  export function describe(name: string, fn: () => void): void;
  export function test(name: string, fn: () => void | Promise<void>): void;
  export function expect(actual: any): any;
}
