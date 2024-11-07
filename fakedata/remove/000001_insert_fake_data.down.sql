-- Удаление данных из таблицы character_skins
DELETE FROM character_skins WHERE character_name IN ('Rookie', 'Miner', 'Crypto Knight', 'Data Sorcerer', 'Blockchain Samurai', 'Quantum Hacker', 'Crypto God');

-- Удаление данных из таблицы character_levels
DELETE FROM character_levels WHERE level_number BETWEEN 1 AND 7;