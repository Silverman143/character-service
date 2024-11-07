-- Удаление триггера
DROP TRIGGER IF EXISTS character_changes_trigger ON characters;

-- Удаление функции триггера
DROP FUNCTION IF EXISTS log_character_changes();

-- Удаление индексов
DROP INDEX IF EXISTS idx_character_change_log_changed_at;
DROP INDEX IF EXISTS idx_character_change_log_user_id;
DROP INDEX IF EXISTS idx_character_change_log_character_id;
DROP INDEX IF EXISTS idx_character_skins_unlock_level;
DROP INDEX IF EXISTS idx_characters_user_id;

-- Удаление таблиц
DROP TABLE IF EXISTS character_change_log;
DROP TABLE IF EXISTS characters;
DROP TABLE IF EXISTS character_skins;
DROP TABLE IF EXISTS character_levels;