-- Вставка данных в таблицу character_levels
INSERT INTO character_levels (level_number, price, referrals, referral_to_open, mining_force, mining_duration_minuts, game_multiplayer)
VALUES
    (1, 0, 0, 0, 10, 4, 1),
    (2, 100, 1, 5, 20, 4, 2),
    (3, 200, 2, 10, 30, 4, 3),
    (4, 300, 3, 15, 40, 4, 4),
    (5, 400, 4, 20, 50, 4, 5),
    (6, 500, 5, 25, 60, 4, 6),
    (7, 600, 6, 30, 70, 4, 7);

-- Вставка данных в таблицу character_skins
INSERT INTO character_skins (character_name, character_lore, character_image_url, unlock_level)
VALUES
    ('Rookie', 'A fresh face in the crypto mining world.', 'https://example.com/rookie.png', 1),
    ('Miner', 'Experienced in the art of digital digging.', 'https://example.com/miner.png', 2),
    ('Crypto Knight', 'A valiant defender of blockchain realms.', 'https://example.com/crypto_knight.png', 3),
    ('Data Sorcerer', 'Mastering the arcane arts of algorithms.', 'https://example.com/data_sorcerer.png', 4),
    ('Blockchain Samurai', 'Slicing through transactions with precision.', 'https://example.com/blockchain_samurai.png', 5),
    ('Quantum Hacker', 'Bending the rules of digital reality.', 'https://example.com/quantum_hacker.png', 6),
    ('Crypto God', 'The ultimate form of digital existence.', 'https://example.com/crypto_god.png', 7);