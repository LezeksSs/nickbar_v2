/* Сначала убираем таблицу коктейлей, завязанную на users.id */
DROP TABLE IF EXISTS cocktails;

/* Расширение тоже создавалось в up-скрипте — удалим симметрично */
DROP EXTENSION IF EXISTS "uuid-ossp";