/* Удаляем таблицу пользователей */
DROP TABLE IF EXISTS users;

/* При необходимости удаляем расширение uuid-ossp
   (без ошибки, даже если его уже нет) */
DROP EXTENSION IF EXISTS "uuid-ossp";