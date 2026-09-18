-- Fluxo de revisao humana obrigatoria para itens (issue #30).
-- ROADMAP_PESQUISA_2027.md, Fase 4: um item errado ENSINA errado. Item
-- gerado por LLM (ou por qualquer automacao) nao pode chegar ao
-- participante sem que uma pessoa tenha assinado a revisao.
--
-- Estados: draft -> reviewed -> published. Somente 'published' e
-- servido ao jogo (ver ItemRepository.GetRandomUnseen*).

ALTER TABLE phishing_quest.items
    ADD COLUMN IF NOT EXISTS status VARCHAR(16) NOT NULL DEFAULT 'draft',
    ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS published_at TIMESTAMPTZ;

-- Itens que ja existiam ANTES deste fluxo foram criados manualmente e
-- ja estao em uso pelo jogo. Marca-los como 'draft' (o novo default)
-- esvaziaria silenciosamente o pool de itens da aplicacao, entao sao
-- promovidos a 'published' de uma vez. O default 'draft' vale para
-- todo item criado a partir de agora.
UPDATE phishing_quest.items
SET status = 'published',
    published_at = COALESCE(published_at, created_at)
WHERE status = 'draft';

ALTER TABLE phishing_quest.items
    DROP CONSTRAINT IF EXISTS chk_items_status;

ALTER TABLE phishing_quest.items
    ADD CONSTRAINT chk_items_status CHECK (status IN ('draft', 'reviewed', 'published'));

-- Um item publicado SEM revisor registrado violaria o criterio de
-- aceite da issue. A regra e garantida no dominio (Item.Publish) e
-- reforcada aqui, para que nem um UPDATE manual no banco consiga
-- publicar um item sem revisao humana.
--
-- NOT VALID e deliberado: os itens promovidos acima foram criados
-- ANTES deste fluxo e nao tem revisor registrado (nao havia onde
-- registrar). Validar retroativamente exigiria inventar um revisor
-- para eles, o que seria pior do que admitir a lacuna. Com NOT VALID
-- o Postgres nao checa as linhas existentes, mas passa a exigir a
-- regra em TODO INSERT e em todo UPDATE dessas linhas — ou seja,
-- nenhum item novo consegue ser publicado sem revisao, e qualquer
-- edicao futura de um item legado tambem passa a exigir revisor.
ALTER TABLE phishing_quest.items
    DROP CONSTRAINT IF EXISTS chk_items_published_requires_reviewer;

ALTER TABLE phishing_quest.items
    ADD CONSTRAINT chk_items_published_requires_reviewer
    CHECK (status <> 'published' OR reviewed_by IS NOT NULL) NOT VALID;

COMMENT ON COLUMN phishing_quest.items.status IS 'Estado no pipeline de curadoria: draft (rascunho, inclusive gerado por LLM) | reviewed (revisado por humano) | published (liberado para o jogo). Somente published e servido aos participantes.';
COMMENT ON COLUMN phishing_quest.items.reviewed_at IS 'Momento em que a revisao humana foi registrada';
COMMENT ON COLUMN phishing_quest.items.published_at IS 'Momento em que o item foi liberado para o jogo';

CREATE INDEX IF NOT EXISTS idx_items_status ON phishing_quest.items(status);
