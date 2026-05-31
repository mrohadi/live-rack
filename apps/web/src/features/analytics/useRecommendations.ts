import { useCallback, useEffect, useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { useAuth } from "react-oidc-context";
import { useApi } from "../../lib/api";
import { openRecommendationSocket, type Recommendation } from "../../lib/ws";

const MAX_CARDS = 20;

/** Prepend a recommendation, dedupe by id, cap the list length. Pure. */
export function addRecommendation(list: Recommendation[], rec: Recommendation): Recommendation[] {
  const deduped = list.filter((r) => r.id !== rec.id);
  return [rec, ...deduped].slice(0, MAX_CARDS);
}

/** Live recommendation feed from the insight engine over WebSocket. */
export function useRecommendations(): Recommendation[] {
  const auth = useAuth();
  const token = auth.user?.access_token ?? null;
  const [recs, setRecs] = useState<Recommendation[]>([]);

  useEffect(() => {
    return openRecommendationSocket(
      async () => token,
      (rec) => setRecs((prev) => addRecommendation(prev, rec)),
    );
  }, [token]);

  return recs;
}

interface ApplyResponse {
  task_id: string;
  status: string;
}

/** Apply a recommendation: create a task from its suggested action. */
export function useApplyRecommendation() {
  const { post } = useApi();
  return useMutation({
    mutationFn: (rec: Recommendation) =>
      post<ApplyResponse>("/api/v1/recommendations/apply", {
        store_id: rec.store_id,
        suggested_task: rec.suggested_task,
      }),
  });
}

/** Stable callback wrapper for triggering an apply. */
export function useApplyHandler() {
  const apply = useApplyRecommendation();
  return useCallback((rec: Recommendation) => apply.mutate(rec), [apply]);
}
