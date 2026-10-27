/** Chronomètre avec pause : seul le temps de frappe effectif compte. */
export class Clock {
  private accumulated = 0;
  private since: number | null = null;
  started = false;

  constructor(private readonly now: () => number = () => performance.now()) {}

  get running(): boolean {
    return this.since !== null;
  }

  start(): void {
    if (this.started) return;
    this.started = true;
    this.since = this.now();
  }

  pause(): void {
    if (this.since === null) return;
    this.accumulated += this.now() - this.since;
    this.since = null;
  }

  resume(): void {
    if (this.started && this.since === null) this.since = this.now();
  }

  elapsed(): number {
    return this.accumulated + (this.since === null ? 0 : this.now() - this.since);
  }
}
