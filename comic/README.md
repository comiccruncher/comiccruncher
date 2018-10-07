# 🎭 Comic

The comic package contains the models and repositories for publishers, issues, and characters and their issues, sources, and sync logs.

## Helpful queries

### Most popular characters

#### All
```postgresql
SELECT count(*) AS issue_count, c.id, concat_ws(',', c.name, c.other_name) AS name, c.other_name, c.slug
FROM characters c
  INNER JOIN character_issues ci ON ci.character_id = c.id
WHERE c.is_disabled = FALSE
GROUP BY c.slug, c.id, c.name, c.other_name
ORDER BY issue_count DESC;
```

#### Marvel
```postgresql
SELECT count(*) AS issue_count, c.id, concat_ws(',', c.name, c.other_name) AS name, c.other_name, c.slug
FROM characters c
  INNER JOIN character_issues ci ON ci.character_id = c.id
WHERE c.publisher_id = 1
  AND c.is_disabled = FALSE
GROUP BY c.slug, c.id, c.name, c.other_name
ORDER BY issue_count DESC;
```

#### DC
```postgresql
SELECT count(*) AS issue_count, c.id, concat_ws(',', c.name, c.other_name) AS name, c.other_name, c.slug
FROM characters c
  INNER JOIN character_issues ci ON ci.character_id = c.id
WHERE c.publisher_id = 2
  AND c.is_disabled = FALSE
GROUP BY c.slug, c.id, c.name, c.other_name
ORDER BY issue_count DESC;
```

### Most popular characters by main appearances only

#### All
```postgresql
SELECT count(*) AS issue_count, c.id, concat_ws(',', c.name, c.other_name) AS name, c.other_name, c.slug
FROM characters c
  INNER JOIN character_issues ci ON ci.character_id = c.id
WHERE c.is_disabled = FALSE
  AND ci.appearance_type & B'00000001' > 0::BIT(8)
GROUP BY c.slug, c.id, c.name, c.other_name
ORDER BY issue_count DESC;
```

#### Marvel
```postgresql
SELECT count(*) AS issue_count, c.id, concat_ws(',', c.name, c.other_name) AS name, c.slug
FROM characters c
  INNER JOIN character_issues ci ON ci.character_id = c.id
WHERE c.publisher_id = 1
  AND c.is_disabled = FALSE
  AND ci.appearance_type & B'00000001' > 0::BIT(8)
GROUP BY c.slug, c.id, c.name, c.other_name
ORDER BY issue_count DESC;
```

#### DC
```postgresql
SELECT count(*) AS issue_count, c.id, concat_ws(',', c.name, c.other_name) AS name, c.other_name, c.slug
FROM characters c
  INNER JOIN character_issues ci ON ci.character_id = c.id
WHERE c.publisher_id = 2
  AND c.is_disabled = FALSE
  AND ci.appearance_type & B'00000001' > 0::BIT(8)
GROUP BY c.slug, c.id, c.name, c.other_name
ORDER BY issue_count DESC;
```

### Most popular characters X amount of years
```postgresql
SELECT count(*) AS issue_count, c.id, c.name, c.other_name, c.slug
FROM characters c
  INNER JOIN character_issues ci ON ci.character_id = c.id
  INNER JOIN issues i on i.id = ci.issue_id
WHERE c.publisher_id = 1
      AND c.is_disabled = FALSE
      AND date_part('year', i.sale_date) > 2010
GROUP BY c.slug, c.id, c.name, c.other_name
ORDER BY issue_count DESC;
```
