SELECT p.id,
    p.user_id,
    p.title,
    p.content,
    p.tags,
    p.created_at,
    p.version,
    u.username,
    COUNT(c.id) AS comments_count
FROM posts p
    LEFT JOIN comments c ON p.id = c.post_id
    LEFT JOIN users u on u.id = p.user_id
    JOIN followers f ON f.follower_id = p.user_id
    OR p.user_id = $1
WHERE f.user_id = $1
    OR p.user_id = $1
GROUP BY p.id,
    u.username
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;