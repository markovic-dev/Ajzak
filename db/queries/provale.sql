-- name: GetLeastUsedProvale :many
SELECT id, tekst, usage_count
FROM provale
ORDER BY usage_count ASC
LIMIT 5;

-- name: IncrementProvalaUsage :exec
UPDATE provale
SET usage_count = usage_count + 1
WHERE id = ?;