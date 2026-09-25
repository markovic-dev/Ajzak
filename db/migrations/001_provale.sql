-- +goose Up
CREATE TABLE provale (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tekst TEXT NOT NULL,
    usage_count INTEGER NOT NULL DEFAULT 0
);

-- Ubacujemo par početnih provala
INSERT INTO provale (tekst) VALUES 
('Spejs nigrutin iz svemira stig''o, na krovu od zgrade sam zastavu podig''o!'),
('Otiš''o sam u prodavnicu po pivo i čips, sreo sam vanzemaljca nosio je gips.'),
('Kada vidim sarme na stolu odma proradi mi stomak!'),
('Nemam para ni za kartu, švercujem se trolom.');

-- +goose Down
DROP TABLE provale;