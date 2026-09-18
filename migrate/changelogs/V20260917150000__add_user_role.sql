-- Adiciona role a users. Necessario para autorizar o export de
-- pesquisa (issue #26) a apenas researcher/admin, e para o front
-- diferenciar area de jogador vs area administrativa
-- (ROADMAP_PESQUISA_2027.md, Fase 2).

ALTER TABLE phishing_quest.users
    ADD COLUMN IF NOT EXISTS role VARCHAR(32) NOT NULL DEFAULT 'participant';

ALTER TABLE phishing_quest.users
    ADD CONSTRAINT chk_users_role CHECK (role IN ('participant', 'researcher', 'admin'));

COMMENT ON COLUMN phishing_quest.users.role IS 'participant (default, jogador comum) | researcher (acesso a export de dados de pesquisa) | admin (gestao de conteudo/usuarios)';
