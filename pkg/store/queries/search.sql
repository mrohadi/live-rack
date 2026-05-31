-- name: SearchEntities :many
-- Fuzzy ⌘K search across items (sku + name) and zones (name) for one org.
-- Trigram similarity ranks fuzzy hits; ILIKE catches short substrings GIN trigram
-- can still serve. Results ordered by best score, capped by max_results.
SELECT kind, id, label, sublabel, score
FROM (
    SELECT
        'item'::text AS kind,
        i.id         AS id,
        i.name       AS label,
        i.sku        AS sublabel,
        GREATEST(
            similarity(i.name, @query::text),
            similarity(i.sku, @query::text)
        )::real      AS score
    FROM items i
    WHERE i.org_id = @org_id
      AND (
          i.name % @query::text
          OR i.sku % @query::text
          OR i.name ILIKE '%' || @query::text || '%'
          OR i.sku ILIKE '%' || @query::text || '%'
      )

    UNION ALL

    SELECT
        'zone'::text                           AS kind,
        z.id                                   AS id,
        z.name                                 AS label,
        z.type::text                           AS sublabel,
        similarity(z.name, @query::text)::real AS score
    FROM zones z
    WHERE z.org_id = @org_id
      AND (
          z.name % @query::text
          OR z.name ILIKE '%' || @query::text || '%'
      )
) results
ORDER BY score DESC, label ASC
LIMIT @max_results::int;
