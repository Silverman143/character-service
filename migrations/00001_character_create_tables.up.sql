-- Создание таблицы уровней (должна быть первой из-за foreign key зависимостей)
CREATE TABLE character_levels (
    level_id SERIAL PRIMARY KEY,
    level_number INTEGER NOT NULL UNIQUE,
    price INTEGER NOT NULL CHECK (price >= 0),
    referrals INTEGER NOT NULL CHECK (referrals >= 0),  -- referrals to buy
    referral_to_open INTEGER NOT NULL CHECK (referral_to_open >= 0),    -- referrals to open
    mining_force INTEGER NOT NULL CHECK (mining_force >= 0),
    mining_duration_minuts INTEGER NOT NULL CHECK (mining_duration_minuts >= 0),
    game_multiplayer INTEGER NOT NULL CHECK (game_multiplayer >= 0)
);

-- Создание таблицы скинов
CREATE TABLE character_skins (
    skin_id SERIAL PRIMARY KEY,
    character_name VARCHAR(255) NOT NULL,
    character_lore TEXT,
    character_image_url TEXT,
    unlock_level INTEGER NOT NULL,
    FOREIGN KEY (unlock_level) REFERENCES character_levels(level_number)
);

-- Создание таблицы персонажей
CREATE TABLE characters (
    character_id SERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL,
    current_level INTEGER NOT NULL DEFAULT 1,
    current_skin_id INTEGER NOT NULL DEFAULT 1,
    CONSTRAINT fk_current_level FOREIGN KEY (current_level) REFERENCES character_levels(level_number),
    CONSTRAINT fk_current_skin FOREIGN KEY (current_skin_id) REFERENCES character_skins(skin_id)
);

-- Создание таблицы для логирования изменений
CREATE TABLE character_change_log (
    log_id SERIAL PRIMARY KEY,
    character_id INTEGER NOT NULL,
    user_id BIGINT NOT NULL,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    operation VARCHAR(10) NOT NULL,
    old_data JSONB,
    new_data JSONB
);

-- Создание функции для триггера
CREATE OR REPLACE FUNCTION log_character_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO character_change_log (character_id, user_id, operation, new_data)
        VALUES (NEW.character_id, NEW.user_id, TG_OP, row_to_json(NEW));
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO character_change_log (character_id, user_id, operation, old_data, new_data)
        VALUES (NEW.character_id, NEW.user_id, TG_OP, row_to_json(OLD), row_to_json(NEW));
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO character_change_log (character_id, user_id, operation, old_data)
        VALUES (OLD.character_id, OLD.user_id, TG_OP, row_to_json(OLD));
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Создание триггера
CREATE TRIGGER character_changes_trigger
AFTER INSERT OR UPDATE OR DELETE ON characters
FOR EACH ROW EXECUTE FUNCTION log_character_changes();

-- Индексы для оптимизации запросов
CREATE INDEX idx_characters_user_id ON characters(user_id);
CREATE INDEX idx_character_skins_unlock_level ON character_skins(unlock_level);
CREATE INDEX idx_character_change_log_character_id ON character_change_log(character_id);
CREATE INDEX idx_character_change_log_user_id ON character_change_log(user_id);
CREATE INDEX idx_character_change_log_changed_at ON character_change_log(changed_at);
