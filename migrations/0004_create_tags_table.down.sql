/* 1. Таблица связей коктейль-тег (FK → cocktails.id, tags.id) */
DROP TABLE IF EXISTS cocktail_tags;

/* 2. Таблица самих тегов */
DROP TABLE IF EXISTS tags;

/* 3. Удаляем расширение uuid-ossp, если оно лишнее */
DROP EXTENSION IF EXISTS "uuid-ossp";
