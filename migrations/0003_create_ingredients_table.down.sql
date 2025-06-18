/* Откат выполняем в порядке обратном зависимостям */

/* 1. Таблица взаимозаменяемых ингредиентов (FK → ingredients.id) */
DROP TABLE IF EXISTS ingredient_replacements;

/* 2. Таблица ингредиентов (FK → cocktails.id, ingredients_nomenclature.id) */
DROP TABLE IF EXISTS ingredients;

/* 3. Справочник номенклатуры ингредиентов */
DROP TABLE IF EXISTS ingredients_nomenclature;

/* 4. Удаляем расширение, если больше не нужно */
DROP EXTENSION IF EXISTS "uuid-ossp";
