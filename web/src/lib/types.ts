export interface Hero {
  id: number;
  name: string;
  class_name: string;
  image_url: string;
}

export interface Item {
  id: number;
  name: string;
  class_name: string;
  item_slot_type: string;
  item_tier: number;
  cost: number;
  image_url: string;
}

export interface WPAResult {
  hero_id: number;
  item_id: number;
  context_key: string;
  mean_delta_w: number;
  mean_initial_w: number;
  win_rate: number;
  sample_size: number;
  std_delta_w: number;
}

export interface ModelStats {
  trained: boolean;
  trained_at?: string;
  accuracy?: number;
  ece?: number;
  num_matches?: number;
  message?: string;
}

export interface StatusInfo {
  match_count: number;
  latest_match?: string;
  model_accuracy: number;
}
