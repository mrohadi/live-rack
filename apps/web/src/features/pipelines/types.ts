/** Card priority — mirrors the pipeline_cards.priority CHECK constraint. */
export type CardPriority = "low" | "medium" | "high";

export interface Pipeline {
  id: string;
  key: string;
  name: string;
}

export interface Stage {
  position: number;
  name: string;
  sla_seconds: number;
}

export interface Card {
  id: string;
  stage_position: number;
  title: string;
  sku?: string;
  priority: CardPriority;
  owner_id?: string;
  entered_stage_at: string;
  /** Seconds the card has dwelt in its current stage (server-computed). */
  age_seconds: number;
  /** True when age exceeds the stage SLA. */
  ageing: boolean;
}

export interface Board {
  pipeline: Pipeline;
  stages: Stage[];
  cards: Card[];
}
